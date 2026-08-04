---
description: Parallel-lens code review — blind specialist subagents (security, architecture, tests, performance, migrations, observability, api-contracts, compliance) reconciled by a judge pass. Use for medium-to-large diffs (4+ files) or cross-stack changes.
argument-hint: [target]
model: claude-opus-5
effort: xhigh
allowed-tools: Read, Grep, Glob, Write, Bash(git:*), Bash(gh:*), Bash(codex exec:*), Bash(codex --version), Bash(kimi -p:*), Bash(kimi --version), Bash(mkdir:*), Bash(rm -f .m/handoff/:*), Bash(rm -rf .m/handoff/:*), Agent
---
# /m:review-fanout - Parallel-Lens Review

Fan out a change set to several specialist reviewer subagents **in parallel**, each blind to the others, then reconcile their findings with a judge pass. Complements `/m:review` — use this when the target is medium/large, crosses domain boundaries, or needs breadth rather than a single deep serial sweep.

## When to use this over `/m:review`

- Change touches 4+ files or crosses layer boundaries (handler → service → repo → migration)
- Change affects Go + React + infra at once
- You want a second opinion on architecture without running the whole sequential pipeline twice
- Pre-merge gate on a project whose `.m/pipeline.yml` requires compliance + security + architecture sign-off

For small, single-file changes, use `/m:review`. Fanout has higher token cost and is only worth it when the lenses would genuinely disagree.

## Input

Target: `$ARGUMENTS`

Same resolution rules as `/m:review` — explicit URL, PR URL, local diff, or Jira key.

## Jira Context

Follow the same Jira resolution flow as `/m:review` (see `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md`). Prepend a **Jira Context** block to the final judge output.

## Context Sources

- `.m/jira.yml`, `.m/INDEX.md`, `.m/GAPS.md`, `PROJECT_INDEX.md`
- Repo `AGENTS.md` / `CLAUDE.md`
- `${CLAUDE_PLUGIN_ROOT}/rules/verification.md` — all lenses and the judge must apply these rules
- The actual changed files and directly impacted adjacent code

## Execution

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, before any review work: run
   `mkdir -p .m && touch .m/phase-review-started` via Bash.
2. On successful completion (judge verdict emitted): run
   `touch .m/phase-review-done`.
3. On hard-block or mid-run abort: leave `-started` in place and do NOT
   write `-done`.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `review`, stop and tell the user — the pipeline is out of sync.

### Step 0: Footprint + Lens Selection

Measure the change with `git diff --shortstat` (unstaged), `git diff --shortstat --cached` (staged), and `git diff --name-only` (file list).

**Empty change set.** If there is no diff, nothing staged, and the target resolves to zero changed files, there is nothing to review — report "no change set to review" plainly, return an `N/A` verdict, and stop. Do not spawn lenses or fabricate findings for nonexistent code.

Select which lenses to spawn based on what the diff actually touches. **Do not spawn lenses that have nothing to review** — lens-count is not a quality metric.

| Lens | Spawn when the diff touches… | Subagent |
|---|---|---|
| **security** | Auth, input handling, query construction, crypto, cookies, secrets, file I/O | `go-security-reviewer` (Go targets) or `code-reviewer` with a security lens prompt |
| **architecture** | New packages, cross-layer calls, interface changes, module boundaries, dependency flow | `code-reviewer` with an architecture lens prompt |
| **tests** | Any business logic, auth, money, parsing, data integrity | `test-runner` + a test-quality lens |
| **performance** | Hot paths, loops, DB queries, N+1 risk, concurrency | `code-reviewer` with a perf lens prompt |
| **migrations** | `*.sql`, schema, data backfills, Alembic/GORM migrations | `code-reviewer` with a migration-safety lens prompt |
| **observability** | Logging, metrics, tracing, error wrapping, panic recovery | `code-reviewer` with an observability lens prompt |
| **api-contracts** | HTTP/gRPC handlers, request/response types, OpenAPI, versioned endpoints | `code-reviewer` with a contracts lens prompt |
| **compliance** | Repo whose `.m/pipeline.yml` sets `compliance.enabled: true` | `code-reviewer` with the per-project compliance lens from `/m:review` Pass 4 |

Minimum: always include security + architecture. Add the rest only when the diff justifies them.

**Subagent model tier (cost control).** When spawning each lens via the Agent tool, set the `model` parameter by blast radius rather than letting every lens inherit the orchestrator's `opus`:

- **`opus`** — `security`, `migrations`, and `compliance`. These carry the highest blast radius; the Go security-review policy and the `verification.md` migration / second-model floor require top-tier reasoning, so they are never downgraded.
- **`sonnet`** — `architecture`, `tests`, `performance`, `observability`, and `api-contracts`. Each is scoped to a small file subset where Sonnet matches Opus on benchmarks, and the Opus judge pass reconciles their output.

The judge pass always runs on the orchestrator's own model (`opus`), never a downgraded subagent. Tiering changes only which model each blind lens runs on — it never reduces the number of lenses or the rigor each one applies.

### Step 1: Blind Fan-Out

Spawn all selected lens subagents **in parallel, in a single message with multiple Agent tool calls** — each with the `model` from the Subagent model tier above. Each lens receives:

- Only the files relevant to its lens (not the full diff)
- A lens-specific system prompt (see Lens Prompt Templates below)
- The shared verification rules from `${CLAUDE_PLUGIN_ROOT}/rules/verification.md`
- **No output from any other lens.** Reviews must be blind; agreement theater is the failure mode to avoid.

Each lens returns a structured report:

```
Lens: {lens_name}
Files reviewed: {list}
Findings: N (critical/high/medium/low)

Critical:
- [file:line] description + evidence snippet + why it matters

High:
- ...

Clean areas:
- ...

Needs verification:
- ...
```

Findings without exact `file:line` and a verbatim code snippet from Read() are discarded by the lens before returning (per verification.md).

### Step 2: Judge Pass

After all lenses return, the orchestrator (this command, in the main thread) runs a judge pass. **Do not spawn another subagent for the judge** — the judge needs full cross-lens context the subagents cannot share.

Judge responsibilities:

1. **Deduplicate.** Multiple lenses flagging the same `file:line` → merge into one finding, credit all lenses.
2. **Resolve contradictions.** If security says "X is unsafe" and architecture says "X is fine because of Y middleware", Read the middleware and decide. The decision must be shown.
3. **Re-verify critical and high findings.** For every `critical` and `high` severity finding, Read the cited file and confirm the evidence. Discard any that fail re-verification (hallucination filter). This matches the all-findings verification floor `/m:review` applies in its Step 3.
4. **Cross-lens synthesis.** Look for issues only visible across lenses:
   - Security says "input validated", architecture says "validator bypassed by new path" → cross-cut finding
   - Tests say "happy path covered", contracts say "new error case added" → test gap
5. **Rank** by merged severity.

### Step 3: Second Engine (mandatory when a provider is selected)

When `second_engine.provider` is `codex` or `kimi` (see `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`, including the legacy `codex:` fallback), the second-engine review runs on **every** fanout review automatically; there is no `y/n` prompt and no high-stakes gating. Follow the active provider's protocol Section 12 (`${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` or `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md`) for the full flow — config resolution (Section 1), the Metered Invocation (Section 6), and Token Metering (Section 7). `--second-opinion` in `$ARGUMENTS` is a no-op for enabling; accepted for backward compatibility. The second engine is skipped (noted in metadata, never a hard failure) only when the provider is `none`, the CLI is unavailable, or the per-run token budget is reached.

**Fanout-specific handling.** Emit the second engine's findings under a dedicated **Second Opinion ({provider})** section alongside the judge verdict — do **not** merge it into the judge's findings list automatically. Apply Section 12's "stricter verdict wins" rule: if the judge says `APPROVED` but the second engine flags criticals that survive the hallucination filter, downgrade the verdict to `BLOCKED` and surface the disagreement at the top of the output.

### Step 4: PR Posting Gate

When `/m:review-fanout` is invoked directly by the user (not as a phase of `/m:develop`) against a GitHub PR the user did not author, publish the review to the PR. In every other invocation mode the review remains chat-only.

For fanout, the `{Findings block}` in the body template refers to the judge-reconciled merged list (the same list shown in chat under `### Findings`) — not the per-lens output.

**Detection.** Walk up from the current working directory looking for `.m/DEVELOP_ACTIVE`. If that file exists with a `current_phase:` line, the review is pipeline-invoked — **skip this step entirely**. Otherwise, the review is direct-invocation. Continue.

The remaining checks all apply only when the resolved target is a GitHub PR URL (or a PR number resolvable via `gh pr view`). For non-PR targets (local diff, commit SHA, file path), skip this step.

**Suppression gates (in order; any positive match skips the post).** Run the shared gate script in `${CLAUDE_PLUGIN_ROOT}/references/review-post-gate.md` and honor its `SKIP=1` outcome as defined there.

**Body template.** When posting, the body MUST begin with the HTML signature, followed by the same review content printed in chat:

```
<!-- m:review:posted -->
### /m:review-fanout report

{Review Metadata block}

{Findings block}

{Discarded Findings block}

{Pre-Existing Gaps block}

**Verdict:** {BLOCKED | APPROVED WITH WARNINGS | APPROVED}

<sub>Posted by /m:review-fanout.</sub>
```

**Posting branch.**

- If verdict is `APPROVED` AND merged findings count (after judge re-verification) is zero:
  ```bash
  gh pr review "$PR_URL" --approve --body "$BODY"
  ```
- Otherwise (any other verdict, OR `APPROVED` with non-zero findings):
  ```bash
  gh pr comment "$PR_URL" --body "$BODY"
  ```

**Failure handling.** Any non-zero exit from `gh pr review --approve` (HTTP 422 own-PR, 403 branch-protection, 404, network) falls back to `gh pr comment` with the same body, and the fallback is noted in chat. Any non-zero exit from `gh pr comment` is logged in chat and the chat review is still printed in full. Posting failures NEVER block the chat output and NEVER change the verdict.

**Idempotence note.** The `<!-- m:review:posted -->` signature is the sole idempotence guard. A repeated invocation against the same PR after a successful post will detect the prior signature and skip — by design. To force a new post (e.g., the code has changed materially), edit out the prior signature in GitHub's UI, or delete the prior comment, then re-run.

## Output

Return in chat only. Format:

### Review Metadata

- **Footprint**: {tier}
- **Lenses spawned**: {list}
- **Findings per lens**: security:N, architecture:N, ...
- **Merged findings**: N after dedup
- **Re-verified critical+high**: N of M survived
- **Second opinion**: {provider + invocation mode} — {agree | disagree with Claude judge} / none — {disabled | unavailable | budget reached}
- **Second-engine tokens**: {run total from the meter} / {token_budget}

### Findings

Grouped by merged severity (critical → low). For each:

- **Title**
- **Severity** / **Confidence**
- **Location**: `file:line`
- **Flagged by**: {lens names} — shows which lenses agreed
- **Evidence**: verbatim snippet
- **Why it matters**
- **Suggested fix** (optional, only if obvious)

### Cross-Lens Findings

Findings only visible by correlating two or more lenses. These are the highest-value output of fanout; call them out explicitly.

### Discarded

Brief list of findings that failed judge re-verification, with reason.

### Pre-Existing Gaps

Append to `.m/GAPS.md`.

### Verdict

- `BLOCKED` — any unresolved critical
- `APPROVED WITH WARNINGS` — non-critical issues only
- `APPROVED` — clean
- `N/A` — no change set to review (empty diff, nothing staged, or no resolved target)

## Lens Prompt Templates

Use the templates in `${CLAUDE_PLUGIN_ROOT}/references/lens-templates.md` as the `prompt` field when spawning each lens via the Agent tool. The reference file contains one template per lens (security, architecture, tests, performance, migrations, observability, api-contracts, compliance) with focus areas, method, and required output format. Read the matching section for each lens you spawn — do not duplicate the template here.

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (read in full before proceeding) to every lens and the judge. Fanout is breadth-first by design — never spawn lenses serially, never let the more permissive verdict win silently.
- Lenses run in **parallel and blind**. Breaking either property breaks the pattern.
- Every lens applies `${CLAUDE_PLUGIN_ROOT}/rules/verification.md`. No exceptions.
- Judge re-verifies every critical and high finding. No exceptions.
- Spawn only lenses the diff justifies. Empty lenses waste tokens and dilute the judge pass.
- Do not create separate review files. Output is chat-only unless the user asks otherwise.
- If any lens returns zero findings, say so explicitly in the metadata — silence is not a pass.
