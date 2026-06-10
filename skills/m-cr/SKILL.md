---
name: m-cr
description: Security review of changed code — evidence-only, read-only. Use for manual security verification of PRs, commits, or local changes.
allowed-tools: Read, Grep, Glob, Bash(git:*), Bash(gh pr view:*), Bash(gh pr diff:*), Bash(gh api:*)
model: opus
effort: max
disable-model-invocation: true
---

ultrathink

Perform a security and integrity review for: $ARGUMENTS

## Core Rules

- **Evidence first** — every finding needs verbatim code proof from Read()
- **Read-only** — no edits, no fixes unless explicitly asked
- **Changed code only** — do not audit the entire codebase, only what changed
- **No style commentary** — security and integrity defects only
- **Prefer zero findings over weak findings**
- **If you can't prove it, put it in "Needs Verification" — never guess**

## Step 1: Determine Target & Intent

- If `$ARGUMENTS` is a GitHub PR URL → `gh pr view <url>` and `gh pr diff <url>`
- If `$ARGUMENTS` is empty → review staged + unstaged changes. If clean, fall back to `HEAD~1...HEAD`
- If `$ARGUMENTS` is a commit/range/path → review that target
- If `gh` cannot resolve a PR URL, say so and stop

Collect:
- **Intent**: What is this change trying to do? (from PR title, description, commit messages, or diff context)
- **Changed files**: `git diff --name-only` for the target
- **The diff**: `git diff` for the target

### Jira Context (best-effort)

Resolve a Jira issue to enrich **Intent**. Follow the shared rules in `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md`. Resolution order:

1. Explicit Jira URL in `$ARGUMENTS`.
2. If `$ARGUMENTS` is a PR URL, derive from `gh pr view "$PR_URL" --json headRefName -q .headRefName` and apply `.m/jira.yml` `branchPattern` (default `([A-Z][A-Z0-9]+-\d+)`). Fall back to scanning the PR body for a Jira URL.
3. Local review: current branch name → same regex.
4. Bare `KEY` in `$ARGUMENTS` — only when `.m/jira.yml.projectKey` matches the prefix.

If a key resolves, fetch it with `mcp__atlassian__*` tools (summary, description, acceptance criteria) and add a brief **Jira Context** block to the output under **Intent**. If no key or the MCP is not authenticated, proceed without Jira context — **do not fail the review**.

## Step 2: Review Changed Code

Read() every changed file. For each change, look for:

- **Injection**: SQL concatenation, command injection, XSS, path traversal, template injection, SSRF
- **Auth/Authz**: missing middleware, IDOR, privilege escalation, JWT issues, session problems
- **Data exposure**: PII in logs, secrets in code, error messages leaking internals, overly broad API responses
- **Crypto**: weak algorithms, hardcoded keys, missing TLS, insecure random
- **State & integrity**: race conditions, partial mutations without rollback, inconsistent error handling, data loss on failure paths
- **Infrastructure**: misconfigured CORS, missing security headers, container running as root, unpinned deps, CI/CD injection

For each suspicious pattern: trace the data flow. Read() the surrounding code. Check for existing mitigations (middleware, validation, framework defaults).

## Step 3: Verify Every Finding

For each potential finding before including it:

1. Read() the actual code — quote verbatim
2. Trace the full path from input to vulnerable operation
3. Check for existing mitigations that handle it
4. **Counter-argument**: what's the strongest reason this ISN'T a problem?
5. If counter-argument holds → drop it silently
6. If counter-argument fails → include with evidence of why it fails

## Step 4: Output

### Intent
[2-4 sentence summary of what the change does and why]

### Findings

| # | Title | Severity | File(s) | Impact |
|---|-------|----------|---------|--------|

For each finding:

```
### Finding N: [title]
**Severity**: CRITICAL / HIGH / MEDIUM
**Files**: file:line
**Trace**: [step-by-step through actual code]
**Evidence**: [verbatim code from Read()]
**Counter-argument**: [considered] → [why it fails]
**Impact**: [what goes wrong in practice]
```

### Needs Verification
[Concerns where evidence is incomplete — what to check and how]

### Checked and Cleared
[Areas reviewed that are properly handled, with code reference]

## Step 5: Iterate

If the user provides feedback or asks about specific areas, re-read the relevant code and refine findings. Drop retracted findings, investigate new concerns.
