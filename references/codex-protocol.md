# Codex Dual-Engine Protocol (referenced from `/m:plan`)

## Contents

- [1. Pre-flight: Codex CLI Availability Check](#1-pre-flight-codex-cli-availability-check)
- [2. Codex Invocation Template](#2-codex-invocation-template)
- [3. Secret Redaction Rule (applies to every handoff write)](#3-secret-redaction-rule-applies-to-every-handoff-write)
- [4. Pass-1: Architecture Sanity (blocking)](#4-pass-1-architecture-sanity-blocking)
- [5. Pass-2: Final Plan Review (blocking)](#5-pass-2-final-plan-review-blocking)
- [6. Disagreement Menu](#6-disagreement-menu)
- [7. Round Budget](#7-round-budget)
- [8. Handoff Cleanup](#8-handoff-cleanup)
- [9. Review Second-Opinion Gate](#9-review-second-opinion-gate)


This reference defines the dual-engine handoff used by `/m:plan` Pass-1 and Pass-2 and the gated second-opinion review used by `/m:review` and `/m:review-fanout`. The plan command and review commands link here so their main bodies stay focused on workflow.

## 1. Pre-flight: Codex CLI Availability Check

Before any observation work in `/m:plan` Phase 1, verify the Codex CLI is available for the dual-engine exchange. Run `codex --version` via Bash.

- If the command fails, is not on PATH, or reports a version older than `0.123.0`:
  - Print verbatim: `[WARN] codex unavailable — proceeding single-engine. Upgrade: npm install -g @openai/codex@latest`
  - Set a run-local flag `CODEX_DISABLED=true` and track it in working memory for the remainder of the invocation.
  - Skip Pass-1 and Pass-2 entirely. Do not prompt the user. Do not log to `.m/PLAN.md`.
- Otherwise (`codex --version` succeeds and reports `0.123.0` or newer): proceed. Set `CODEX_DISABLED=false`. Initialize a Codex round budget of 2 total invocations across Pass-1, Pass-2, and any final verification pass combined.

## 2. Codex Invocation Template

Both Pass-1 and Pass-2 use this exact Bash invocation. Do not change the flags without updating `.m/PRD-dual-engine-plan.md`.

```
codex exec \
  -m gpt-5.5 \
  -c model_reasoning_effort="xhigh" \
  -s read-only \
  -C "$PWD" \
  -o .m/handoff/codex-to-claude.md \
  --ephemeral \
  --skip-git-repo-check \
  "$(cat .m/handoff/claude-to-codex.md)"
```

If any Codex exec invocation exits non-zero at runtime, treat it identically to the pre-flight failure path: print the fallback warning, set `CODEX_DISABLED=true`, delete `.m/handoff/claude-to-codex.md` and `.m/handoff/codex-to-claude.md` if present, and continue single-engine.

## 3. Secret Redaction Rule (applies to every handoff write)

Before writing `.m/handoff/claude-to-codex.md`, redact any value matching the deny-list below. Replace matched values with the literal string `[REDACTED]`. This applies to both Pass-1 and Pass-2 writes.

- API keys: `sk-*`, `pk_*`, AWS `AKIA*` or `ASIA*`, Google `AIza*`
- ALL_CAPS environment variable values where the variable name matches `*_TOKEN`, `*_SECRET`, `*_PASSWORD`, or `*_KEY`
- Contents of any `.env*` file
- JWT tokens (tokens beginning `eyJ` with two `.` separators)
- PEM blocks (from `-----BEGIN * PRIVATE KEY-----` through the matching `-----END * PRIVATE KEY-----`)

Scope of handoff writes is plan prose only — decisions, file paths, acceptance criteria, rationale. Never include raw source code, `.env` contents, or log excerpts that may embed secrets.

## 4. Pass-1: Architecture Sanity (blocking)

Runs after observation gathering completes, before Phase 2 begins. Phase 2 must not start until Pass-1 completes or is skipped.

- If `CODEX_DISABLED` is true, or the round budget is zero: skip this pass and proceed to Phase 2.
- Otherwise:
  1. Run `mkdir -p .m/handoff` via Bash.
  2. Build the handoff payload — the raw `$ARGUMENTS`, every `[OBSERVATION]` entry gathered above, plus the request prose:

     > Review these observations against the repository. Identify missing architecture concerns, risks, or alternative approaches that the driver agent has not surfaced. Do not propose a full plan yet. Return a bulleted list of additional concerns, each with a concrete `file:line` reference when applicable.

  3. Apply the Secret Redaction Rule to the payload, then write the result to `.m/handoff/claude-to-codex.md`.
  4. Run the Codex Invocation Template via Bash. Decrement the round budget by 1.
  5. Read `.m/handoff/codex-to-claude.md`.
  6. Merge each Codex-surfaced concern into the observation list as a new `[OBSERVATION — codex]` entry, preserving any `file:line` references verbatim.
  7. Leave the handoff files in place for inspection; they will be cleaned up at the end of Phase 3.

The merged observation list (Claude's original `[OBSERVATION]` entries plus the new `[OBSERVATION — codex]` entries) is the input to Phase 2.

## 5. Pass-2: Final Plan Review (blocking)

Runs after the three exit-gate checks pass, before the plan document is emitted. The plan is not emitted until Pass-2 completes or is skipped.

- If `CODEX_DISABLED` is true, or the Codex round budget is zero: skip Pass-2, finalize tasks, and jump to Handoff Cleanup, then emit the plan.
- Otherwise:
  1. Build the handoff payload: the full final plan draft (summary, architecture decisions, tasks with acceptance criteria, risks) plus this explicit verdict-format instruction appended at the end:

     > End your output with a line of exactly `VERDICT: LGTM` if you have no disagreements with this plan, or `VERDICT: CHANGES` if you do. If `VERDICT: CHANGES`, list each disagreement above the verdict line as a numbered item: `Dn. <short description> | claude: <claude-pos + one-line rationale> | codex: <codex-pos + one-line rationale>`. Include concrete `file:line` references where applicable.

  2. Apply the Secret Redaction Rule to the payload. Write it to `.m/handoff/claude-to-codex.md` (overwriting any Pass-1 content).
  3. Run the Codex Invocation Template via Bash. Decrement the round budget by 1.
  4. Read `.m/handoff/codex-to-claude.md`. Parse the final non-empty line:
     - If it is exactly `VERDICT: LGTM`: finalize tasks, run Handoff Cleanup, emit the plan.
     - If it is exactly `VERDICT: CHANGES`: parse the numbered disagreements above it and enter the Disagreement Menu.
     - If neither verdict string is present on the last non-empty line: treat as `VERDICT: CHANGES` (conservative default). If no parseable disagreements are listed, hard-block and re-grill the user on the final plan content.

## 6. Disagreement Menu

For each numbered disagreement `Dn` Codex listed, present a single bounded-menu question to the user. Source tags are mandatory and must appear verbatim. Example format:

```
Decision D{n}: {short description}
  [claude] {claude-pos + rationale}
  [codex]  {codex-pos + rationale}

Pick:
  A) claude
  B) codex
  C) merge — describe
  D) neither — describe
```

Collect the user's picks for all disagreements. Apply them to the plan draft in order (A = keep Claude's position; B = replace with Codex's position; C/D = integrate the user's described variant).

After all picks are applied, run one final Codex verification pass if the round budget still permits it:

- If the round budget is zero: skip the final verification, run Handoff Cleanup, emit the plan.
- Otherwise: write the merged plan to `.m/handoff/claude-to-codex.md` (with redaction), run the Codex Invocation Template, decrement the round budget, parse the verdict.
  - `VERDICT: LGTM` → run Handoff Cleanup, emit the plan.
  - `VERDICT: CHANGES` → **hard-block**. Convert the remaining disagreements into new Phase 2 grill questions. Do NOT invoke Codex again for the remainder of this run (set `CODEX_DISABLED=true`). Loop back to Phase 2.

## 7. Round Budget

Total Codex invocations across Pass-1, Pass-2, and the final verification pass MUST NOT exceed 2 per `/m:plan` run. Track the budget explicitly. Once exhausted, every further Codex step is skipped as if `CODEX_DISABLED=true`.

## 8. Handoff Cleanup

At the end of Phase 3 — on every terminal path (successful emit, pre-flight fallback, mid-run Codex failure, or hard-block re-grill exit) — delete both handoff files if they exist:

```
rm -f .m/handoff/claude-to-codex.md .m/handoff/codex-to-claude.md
```

Do not persist any dialogue artifact. Only the final plan output (chat or `.m/PLAN.md`) remains.

## 9. Review Second-Opinion Gate

Used by `/m:review` and `/m:review-fanout` after the primary review verdict, gated on:

- Migration touching a production table
- Auth, money, or tenant-isolation code path
- Public API contract change
- Any file matching the repo's `.m/pipeline.yml` `high_stakes_paths` (see `pipeline-context.md`)
- `--second-opinion` appears in `$ARGUMENTS`

Gate flow:

1. Ask the user explicitly: *"Change touches {category}. Run Codex second-opinion review? (y/n)"* — never auto-run.
2. On yes, invoke `codex review` with the flag matching the target:

```
codex review --uncommitted    # staged + unstaged + untracked
codex review --base main      # against a base branch
codex review --commit <SHA>   # a specific commit
```

3. For targeted lenses, pipe a constraining prompt via stdin:

```
echo "Focus on: tenant isolation, SQL safety, and migration atomicity. Report findings with file:line and a verbatim code snippet only." \
  | codex review --uncommitted -
```

4. Present Claude's findings and Codex's findings side-by-side under a dedicated **Second Opinion (Codex)** section. Do not merge silently.
5. **Disagreement rule:** if Claude says `APPROVED` but Codex flags criticals that survive re-verification, downgrade to `BLOCKED`. The more permissive verdict never wins by default.
6. If `codex` is not on PATH, note it in metadata and skip the gate — do not fail the review.
