---
description: Multi-pass sequential code review with unified, evidence-backed findings, a mandatory second engine (Codex or Kimi, config-driven via .m/pipeline.yml second_engine:), and per-project compliance pass. Use for small-to-medium diffs (1-3 files) or when a single deep sweep is preferable to parallel fan-out.
argument-hint: [target]
model: claude-opus-5
effort: xhigh
allowed-tools: Read, Grep, Glob, Write, Bash(git:*), Bash(gh:*), Bash(codex exec:*), Bash(codex --version), Bash(kimi -p:*), Bash(kimi --version), Bash(mkdir:*), Bash(python3:*), Bash(rm:*), Agent
---
# /m:review - Multi-Pass Review Workflow

Review code changes using Claude multi-pass analysis, producing a single evidence-backed report with hallucination verification.

For Go backend targets, this command delegates the security pass to the `go-security-reviewer` subagent (read-only, flow-simulation based). For a breadth-first parallel-lens review, use `/m:review-fanout`.

## Input

Target to review: `$ARGUMENTS`

If no explicit target is given, review the most recent plan or local code changes in context.

## Jira Context (run before Step 0)

Resolve a Jira issue for this review when possible, per `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md` (URL, PR-branch derivation, local branch name, bare-key validation, fetch, unauthenticated behavior — never fail the review over Jira). Place the **Jira Context** block (including acceptance criteria) at the very top of the review output — above **Review Metadata**. During the Deep Review pass, check whether the change actually satisfies the Jira acceptance criteria; any gap is a finding (severity = medium unless it blocks release).

## Context Sources

- `.m/jira.yml` (per-project Jira mapping)
- `.m/INDEX.md`
- `.m/GAPS.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`
- the actual changed files and directly impacted adjacent code

## Execution

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, before any review work: run
   `mkdir -p .m && touch .m/phase-review-started` via Bash.
2. On successful completion (review report emitted): run
   `touch .m/phase-review-done`.
3. On hard-block or mid-run abort: leave `-started` in place and do NOT
   write `-done`.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `review`, stop and tell the user — the pipeline is out of sync.

### Step 0: Determine Footprint

Before any review work, measure the change size with `git diff --shortstat` (unstaged), `git diff --shortstat --cached` (staged), and `git diff --name-only` (file list).

**Empty change set.** If all of the above show no changes (no diff, nothing staged, the target resolves to zero changed files, and the request text itself supplies no diff to review), there is nothing to review. Report "no change set to review" plainly, return an `N/A` verdict, and stop — do not fabricate findings or invent `file:line` references for nonexistent code, and do not return `APPROVED` or `BLOCKED`. A diff supplied inline in the request text is a valid change set even when the working tree is clean — review it as given.

Classify into a tier:

| Tier | Criteria | Passes |
|------|----------|--------|
| **Trivial** | 1 file, ≤5 lines, pure rename / comment / formatting / string-literal change with no logic, control-flow, or security surface | 1 (focused deep review of the changed lines only) |
| **Small** | 1-3 files, <100 lines changed | 1 (deep review only) |
| **Medium** | 4-10 files, 100-500 lines changed | 2 (scope mapping + deep review) |
| **Large** | 10+ files or 500+ lines changed | 3 (scope mapping + deep review + regression re-check) |
| **+Compliance** | Any size, repo's `.m/pipeline.yml` sets `compliance.enabled: true` | +1 compliance pass |

If the target is a PR URL or commit range, use `git diff --shortstat <base>...<head>` instead.

**Trivial fast-path.** When the change is Trivial, still run the Phase Marker Protocol and the Step 3 verification pass — the hallucination filter is never skipped — but limit the review to the changed lines and their direct callers, skip the blast-radius mapping, regression, and compliance passes, and return `APPROVED` directly when no real issue is found. Do not spend a multi-pass budget on a one-line rename. A change verified benign gets `APPROVED` with its evidence cited and nothing more. The fast-path does not apply if the single changed line touches auth, money, a query string, crypto, a `.sql` migration, or a file matching the repo's `.m/pipeline.yml` `high_stakes_paths` — those escalate to the normal tier regardless of size.

### Step 1: Review Passes (scaled by tier)

Run the applicable passes per the footprint tier:

**Pass 1 — Scope and impact mapping** *(Medium + Large only)*
Identify what changed and what it touches. Map the blast radius.

**Pass 2 — Deep review** *(all tiers)*
Security, correctness, code quality, and tests. This is the core pass.

**Pass 3 — Regression re-check** *(Large only)*
Re-examine every flagged area for false positives and missed context.

**Pass 4 — Compliance** *(repos whose `.m/pipeline.yml` sets `compliance.enabled: true`)*

**Detection:** Load `.m/pipeline.yml` (schema: `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`). Run this pass when `compliance.enabled: true`; skip it otherwise.

When enabled, **always** run the compliance pass — not only when data/PII/storage is touched. Apply, in order:
- The frameworks declared in `compliance.frameworks` (e.g. SOC2, GDPR, EU AI Act).
- Every entry in `compliance.rules` as a concrete checklist item traced against the diff.
- Any specs under `.business/` (BUSINESS.md, contracts, regulatory docs), which are the source of truth for domain compliance when present.

Where a rule names a path or field convention (restricted files, `*Enc` envelope-encryption fields, migration immutability), flag any change that violates it. Label all findings from this pass with `[COMPLIANCE]`.

### Step 2: Target Classification

Before finalizing findings, classify the target:

- normal source repo
- extracted source-map or decompiled snapshot
- vendor-heavy tree
- generated or bundled output

Scope style and maintainability judgments to first-party code unless the reviewed target explicitly includes vendor or generated artifacts.

## What To Check

Review against these categories:

- security threats
- bugs and behavioral regressions
- missing or weak tests
- code quality and clean-code issues
- repo-specific invariants from `AGENTS.md`, `CLAUDE.md`, or `.m/INDEX.md`
- repo-health risks such as missing manifests, absent tests, or incomplete source trees
- architectural liabilities such as oversized controller files, module-global mutable state, singleton runtime objects, sync filesystem calls in interactive paths, and excessive environment-flag branching
- prompt-pack or command-pack consistency when reviewing workflow tooling: documented commands must exist and command cross-references must be valid
- **[Compliance]** When the repo's `.m/pipeline.yml` enables compliance, check the declared `compliance.frameworks` and `compliance.rules` against the diff. See Pass 4 for the full pass.

## Step 3: Verification — Deduplicate and Validate

After all passes complete, perform a **final verification pass** over ALL findings.

For **every** finding:

1. **Read the actual code** at the reported `file:line` — confirm the file and line exist
2. **Verify the quoted snippet** matches the real file content
3. **Trace the data flow** — follow the claimed issue through the code path
4. **Check for existing mitigations** — middleware, validation, framework defaults that may already handle the issue
5. **Counter-argument** — what is the strongest reason this is NOT a problem? If the counter-argument holds, discard the finding
6. **If the finding cannot be verified from actual code, discard it** — this is the hallucination filter

Then:

- **Merge duplicates**: combine findings about the same issue
- **Rank** by severity (critical > high > medium > low)

Findings without concrete file:line evidence and verified code proof are NOT findings. Discard them.

## Step 4: Second Engine (mandatory when a provider is selected)

When `second_engine.provider` is `codex` or `kimi` (see `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`, including the legacy `codex:` fallback), the second-engine review runs on **every** review automatically; there is no `y/n` prompt and no high-stakes gating. Follow the active provider's protocol Section 12 (`${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` or `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md`) for the full flow — config resolution (Section 1), the Metered Invocation (Section 6), Token Metering (Section 7), side-by-side presentation under a **Second Opinion ({provider})** section, and the "stricter verdict wins" disagreement rule. Second-engine findings are leads to confirm against real code, never ground truth.

The second engine is skipped (noted in metadata, never a hard failure) only when the provider is `none`, the CLI is unavailable, or the per-run token budget is reached. `--second-opinion` in `$ARGUMENTS` is a no-op for enabling; it remains accepted for backward compatibility.

Record the invocation in the Review Metadata (`Second opinion: codex exec review --{mode}` / `kimi -p (review prompt)` or `none — {reason}`).

### Step 5: PR Posting Gate

When `/m:review` is invoked directly by the user (not as a phase of `/m:develop`) against a GitHub PR the user did not author, publish the review to the PR. In every other invocation mode the review remains chat-only.

**Detection.** Walk up from the current working directory looking for `.m/DEVELOP_ACTIVE`. If that file exists with a `current_phase:` line, the review is pipeline-invoked — **skip this step entirely**. Otherwise, the review is direct-invocation. Continue.

The remaining checks all apply only when the resolved target is a GitHub PR URL (or a PR number resolvable via `gh pr view`). For non-PR targets (local diff, commit SHA, file path), skip this step.

**Suppression gates (in order; any positive match skips the post).** Run the shared gate script in `${CLAUDE_PLUGIN_ROOT}/references/review-post-gate.md` and honor its `SKIP=1` outcome as defined there.

**Body template.** When posting, the body MUST begin with the HTML signature, followed by the same review content printed in chat:

```
<!-- m:review:posted -->
### /m:review report

{Review Metadata block}

{Findings block}

{Discarded Findings block}

{Pre-Existing Gaps block}

**Verdict:** {BLOCKED | APPROVED WITH WARNINGS | APPROVED}

<sub>Posted by /m:review.</sub>
```

**Posting branch.**

- If verdict is `APPROVED` AND findings count (after Step 3 verification) is zero:
  ```bash
  gh pr review "$PR_URL" --approve --body "$BODY"
  ```
- Otherwise (any other verdict, OR `APPROVED` with non-zero findings):
  ```bash
  gh pr comment "$PR_URL" --body "$BODY"
  ```

**Failure handling.** Any non-zero exit from `gh pr review --approve` (HTTP 422 own-PR, 403 branch-protection, 404, network) falls back to `gh pr comment` with the same body, and the fallback is noted in chat. Any non-zero exit from `gh pr comment` is logged in chat and the chat review is still printed in full. Posting failures NEVER block the chat output and NEVER change the verdict.

**Idempotence note.** The `<!-- m:review:posted -->` signature is the sole idempotence guard. A repeated invocation against the same PR after a successful post will detect the prior signature and skip — by design. To force a new post (e.g., the code has changed materially), edit out the prior signature in GitHub's UI, or delete the prior comment, then re-run.

## Persistence

Separate:

- CURRENT findings: introduced or relevant to the reviewed target
- PRE-EXISTING gaps: adjacent issues not caused by the reviewed target

Only append PRE-EXISTING gaps to `.m/GAPS.md`.

## Output

Return the review in chat only.

Use this format:

### Review Metadata

- **Footprint tier**: Small / Medium / Large
- **Passes run**: N
- **Compliance pass**: yes / no / n/a
- **Findings verified**: N of M survived verification
- **Second opinion**: {provider + invocation mode} / none — {disabled | unavailable | budget reached}
- **Second-engine tokens**: {run total from the meter} / {token_budget}

### Findings

List findings ordered by severity. For each finding include:

- **Title**
- **Severity**: critical / high / medium / low
- **Confidence**: high / medium / low
- **Location**: `file_path:line_number`
- **Evidence**: code snippet or diff excerpt proving the issue (verified against real code)
- **Why it matters**: impact description
- **Suggested fix** *(optional)*: only if the fix is obvious and non-trivial

### Discarded Findings

Brief list of findings that failed verification, with reason (hallucinated file/line, code doesn't match, mitigated by existing code, etc.). This section provides transparency on what was filtered.

### Pre-Existing Gaps

Summarize anything logged to `.m/GAPS.md`.

### Verdict

Return one of:

- `BLOCKED` — critical issues that must be fixed
- `APPROVED WITH WARNINGS` — non-critical issues worth addressing
- `APPROVED` — clean review
- `N/A` — no change set to review (empty diff, nothing staged, or no resolved target)

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (loaded at session start). Never skip a tier-required pass; never let the more permissive verdict win when the second engine disagrees.
- Prefer zero findings over weak findings
- Every finding MUST include file:line proof and a verified code snippet — no exceptions
- Every finding goes through the verification pass — no shortcuts
- Do not create separate review files unless the user explicitly asks
- If the repo is incomplete or extracted, note that as a review-confidence limiter instead of pretending verification was exhaustive
