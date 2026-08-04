# PR Posting Suppression Gates — shared by /m:review and /m:review-fanout

The suppression-gate script below runs during the PR Posting Gate step of both review commands, after the detection and PR-target checks in the command body have passed. Run the gates in order; any positive match sets `SKIP=1` with a printed reason, and a set `SKIP=1` means: do not post, print the gate reason in chat, and continue to the chat Output block as today. Pass the resolved PR URL as the script's first positional argument; never substitute the URL into the script text.

```bash
# Required preconditions
PR_URL="$1"                                            # resolved PR URL, passed as the first positional argument
gh auth status >/dev/null 2>&1 || { echo "[gate] gh not authenticated — skip post"; SKIP=1; }

# Gate (b): author check
PR_AUTHOR=$(gh pr view "$PR_URL" --json author -q .author.login 2>/dev/null) || { echo "[gate] gh pr view failed — skip post"; SKIP=1; }
ME=$(gh api user -q .login 2>/dev/null) || { echo "[gate] gh api user failed — skip post"; SKIP=1; }
[ -n "$PR_AUTHOR" ] && [ "$PR_AUTHOR" = "$ME" ] && { echo "[gate] you are PR author — skip post"; SKIP=1; }

# Gate (c) + (d): state and draft
PR_META=$(gh pr view "$PR_URL" --json state,isDraft -q '[.state, (.isDraft|tostring)] | @tsv' 2>/dev/null) || { echo "[gate] gh pr view failed — skip post"; SKIP=1; }
IFS=$'\t' read -r PR_STATE PR_DRAFT <<< "$PR_META"
[ "$PR_STATE" != "OPEN" ] && { echo "[gate] PR state $PR_STATE — skip post"; SKIP=1; }
[ "$PR_DRAFT" = "true" ] && { echo "[gate] PR is draft — skip post"; SKIP=1; }

# Gate (e): idempotence
gh pr view "$PR_URL" --json comments,reviews -q '[.comments[].body, .reviews[].body] | .[]' 2>/dev/null | grep -q '<!-- m:review:posted -->' && { echo "[gate] prior /m:review post detected — skip post"; SKIP=1; }
```

The idempotence marker greps for `<!-- m:review:posted -->`, the first line of both commands' body templates — the templates stay in the command files because their report titles differ per command.
