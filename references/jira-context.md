---
name: Jira Context
description: Shared rules for pulling Jira stories into /m:* workflows via the `atlassian` MCP server.
---

# Jira Context — shared workflow fragment

This file is the single source of truth for how `/m:refine`, `/m:plan`, `/m:implement`, and `/m:review` pull Jira stories. Each of those commands references this file.

## Prerequisites

1. The `atlassian` MCP server is installed at user scope:
   `claude mcp add --transport http --scope user atlassian https://mcp.atlassian.com/v1/mcp`
2. The user has run `/mcp` once to complete the OAuth handshake.
3. If the MCP is unauthenticated, stop and tell the user: **"Run `/mcp` to authenticate with Atlassian, then retry."** Do not attempt to proceed without the story.

## Per-project config: `.m/jira.yml`

Each project that wants Jira integration declares its own `.m/jira.yml` at the repo root:

```yaml
# Atlassian Cloud site — the hostname that appears in browse URLs
site: yourcompany.atlassian.net

# Default Jira project key for this repo (e.g. PROJ, ABC)
# Used to validate bare keys and as a fallback when no URL is given.
projectKey: PROJ

# Regex that extracts a Jira issue key from a git branch name.
# Must contain one capturing group yielding "PROJ-123".
# Default works for feature/PROJ-123-slug, PROJ-123-slug, hotfix/PROJ-99.
branchPattern: "([A-Z][A-Z0-9]+-\\d+)"
```

If `.m/jira.yml` is missing, assume: `branchPattern = ([A-Z][A-Z0-9]+-\d+)`, no projectKey validation for bare keys, and derive `site` from the URL whenever a full URL is supplied.

## Detection — refine / plan / implement

Scan `$ARGUMENTS` for a Jira reference **before** starting the workflow phases:

1. **Full URL**: `https?://[^/\s]+\.atlassian\.net/browse/[A-Z][A-Z0-9]+-\d+` — extract the trailing `KEY`.
2. **Bare key**: `\b[A-Z][A-Z0-9]+-\d+\b` — only treat as a Jira key when `.m/jira.yml` exists **and** the prefix matches `projectKey` (prevents false positives like `HTTP-404`).

If multiple keys are present, prefer the first URL; if no URL, the first bare key that passes validation.

## Detection — review (from PR URL)

When `$ARGUMENTS` is a GitHub PR URL (already handled by `/m:review`), additionally:

1. Get the head branch name:
   ```bash
   gh pr view "$PR_URL" --json headRefName -q .headRefName
   ```
2. Load `.m/jira.yml` if present and use its `branchPattern`; otherwise fall back to `([A-Z][A-Z0-9]+-\d+)`.
3. Run the regex against the branch name and capture group 1 as the Jira key.
4. If the PR body or `$ARGUMENTS` themselves also contain a Jira URL, prefer that over the branch-derived key.
5. If no key is found, proceed with the review without Jira context — do **not** fail.

## Fetch (via atlassian MCP)

Once a key is resolved:

1. Use the `atlassian` MCP server's Jira tools (tool names are prefixed `mcp__atlassian__*` at runtime — discover and select the one that fetches a single issue by key).
2. Retrieve at minimum: `summary`, `description`, `status`, `assignee`, `acceptance criteria` (often in custom fields or at the bottom of the description), and the latest 3–5 comments.
3. Cache the fetched content in working memory for the current turn. Do **not** write it to a file unless the user asks.

## Use

- Treat the Jira story as the authoritative requirements input.
- Prepend a **Jira Context** block to the workflow output that includes: key, title, link, status, and a 2–4 line summary.
- If `$ARGUMENTS` adds scope or contradicts the story, surface the conflict explicitly in the output under **Conflicts with Jira**.
- For `/m:review`, add the Jira Context block above **Findings** so reviewers see requirements before issues.

## Failure modes

- **MCP not installed**: tell the user the exact `claude mcp add` command above.
- **MCP not authenticated**: tell the user to run `/mcp`.
- **Issue not found / permission denied**: report the key and continue without Jira context.
- **Ambiguous key** (no `.m/jira.yml`, bare key only): ask the user to confirm the key or paste the full URL.
