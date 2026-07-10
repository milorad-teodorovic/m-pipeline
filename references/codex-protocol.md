# Codex Dual-Engine Protocol (referenced from `/m:plan`, `/m:research`, `/m:review`, `/m:review-fanout`, `/m:develop`)

## Contents

- [1. Configuration: `.m/pipeline.yml` `codex:` section](#1-configuration-mpipelineyml-codex-section)
- [2. Pre-flight: Availability and Enablement](#2-pre-flight-availability-and-enablement)
- [3. Fast-Mode Flag Construction](#3-fast-mode-flag-construction)
- [4. Operating-Rules Preamble (parity)](#4-operating-rules-preamble-parity)
- [5. Secret Redaction Rule (applies to every handoff write)](#5-secret-redaction-rule-applies-to-every-handoff-write)
- [6. Metered Codex Invocation (used by every pass)](#6-metered-codex-invocation-used-by-every-pass)
- [7. Token Metering and Budget Enforcement](#7-token-metering-and-budget-enforcement)
- [8. Pass-1: Architecture Sanity (plan, blocking)](#8-pass-1-architecture-sanity-plan-blocking)
- [9. Pass-2: Final Plan Review (plan, blocking)](#9-pass-2-final-plan-review-plan-blocking)
- [10. Disagreement Menu (plan)](#10-disagreement-menu-plan)
- [11. Research: Parallel Dual-Engine Researcher](#11-research-parallel-dual-engine-researcher)
- [12. Review: Mandatory Second Engine](#12-review-mandatory-second-engine)
- [13. Handoff Cleanup](#13-handoff-cleanup)

This reference defines the dual-engine handoff used by `/m:plan` (Pass-1 and Pass-2), the parallel dual-engine research used by `/m:research`, and the mandatory second-engine review used by `/m:review` and `/m:review-fanout`. The pipeline commands link here so their main bodies stay focused on workflow.

Codex participation is **configuration-driven and mandatory when enabled**. The previous behavior — optional dual-engine gated behind a per-invocation `y/n` prompt and a fixed two-call round budget — has been replaced by the `codex:` config section (Section 1) plus token metering (Section 7). When Codex is enabled, every applicable pass runs Codex automatically; there is no per-pass permission prompt. When Codex is disabled, every Codex pass is skipped silently and the command runs Claude-only.

## 1. Configuration: `.m/pipeline.yml` `codex:` section

All Codex behavior is controlled by the `codex:` section of the per-project `.m/pipeline.yml` file (full schema in `pipeline-context.md`). Read this section once at the start of any command that uses Codex and hold the resolved values in working memory for the remainder of the run.

```yaml
codex:
  enabled: false              # master switch. true = run Codex; false (default) = Claude-only.
  fast_mode: false            # true = add fast-mode flags (Section 3, ~2.5x credit rate); off by default.
  model: gpt-5.6-sol          # model passed to Codex.
  reasoning_effort: high      # model_reasoning_effort passed to Codex.
  token_budget: 200000        # cumulative Codex tokens allowed per /m run before metering triggers.
  on_budget_exceeded: fallback   # "fallback" (finish Claude-only) | "stop" (halt and save progress).
```

**Defaults when `.m/pipeline.yml` is absent, or the `codex:` section is missing or partial:**

| Key | Default |
|---|---|
| `enabled` | `false` |
| `fast_mode` | `false` |
| `model` | `gpt-5.6-sol` |
| `reasoning_effort` | `high` |
| `token_budget` | `200000` |
| `on_budget_exceeded` | `fallback` |

Codex is **off by default** — the pipeline runs Claude-only until you opt in. To enable the dual-engine passes for a repository, set `codex.enabled: true` in that repo's `.m/pipeline.yml` (requires the `codex` CLI on `PATH`). Fast mode is a separate opt-in via `codex.fast_mode: true` (faster, but ~2.5x the credit rate and ChatGPT-auth only).

Bind the resolved values to named shell variables used by the rest of this protocol:

```bash
# Resolve config (parse .m/pipeline.yml codex: section; fall back to defaults when keys are absent).
CODEX_ENABLED=false           # from codex.enabled, default false (opt-in)
CODEX_FAST=false              # from codex.fast_mode, default false (opt-in)
CODEX_MODEL=gpt-5.6-sol       # from codex.model, default gpt-5.6-sol
CODEX_EFFORT=high             # from codex.reasoning_effort, default high (xhigh for high-stakes repos)
CODEX_BUDGET=200000           # from codex.token_budget, default 200000
CODEX_ON_EXCEED=fallback      # from codex.on_budget_exceeded, default fallback
```

## 2. Pre-flight: Availability and Enablement

Run this once before the first Codex pass of a command.

1. If `CODEX_ENABLED` is `false`: set `CODEX_DISABLED=true`, skip every Codex pass for the rest of the run, do not prompt, do not warn. The command proceeds Claude-only. Stop here.
2. Otherwise, verify the CLI: run `codex --version` via Bash.
   - If the command fails, is not on PATH, or reports a version older than `0.123.0`: print verbatim `[WARN] codex enabled but unavailable — proceeding Claude-only. Upgrade: npm install -g @openai/codex@latest`, set `CODEX_DISABLED=true`, and proceed Claude-only.
   - Otherwise: set `CODEX_DISABLED=false`. Initialize the token meter: `mkdir -p .m/handoff && echo 0 > .m/handoff/codex-meter.txt`.

Because Codex is mandatory when enabled, a genuine CLI failure is surfaced loudly (it is not silent), but it never hard-blocks the run — the command degrades to Claude-only. This is the same terminal behavior as the `on_budget_exceeded: fallback` path in Section 7. If a project requires Codex to be present (hard-fail), that is a project policy decision and is not the default.

## 3. Fast-Mode Flag Construction

Fast mode is driven by the `codex.fast_mode` toggle (off by default), applied as per-invocation flags so the toggle is authoritative for `/m` regardless of any global `~/.codex/config.toml` setting. The on-path adds the fast flags; the off-path explicitly disables fast so a global fast default cannot leak into a `/m` run. The pipeline never edits the global `config.toml`.

The official fast-mode keys (OpenAI Codex docs, `developers.openai.com/codex/speed`) are the config value `service_tier = "fast"` and the feature flag `features.fast_mode = true`. As per-invocation overrides these are `-c 'service_tier="fast"'` / `--enable fast_mode` to turn it on, and `--disable fast_mode` (the documented `-c features.fast_mode=false`, verified to flip the effective state via `codex features list`) to turn it off.

```bash
if [ "$CODEX_FAST" = "true" ]; then
  FAST_FLAGS=(-c 'service_tier="fast"' --enable fast_mode)
else
  FAST_FLAGS=(--disable fast_mode)   # repo opt-out wins over the global fast default
fi
```

Fast mode increases speed (~1.5x measured on `gpt-5.5`) and consumes credits at ~2.5x the standard rate. It requires a ChatGPT-authenticated Codex login (API-key auth falls back to standard pricing). It applies to every metered invocation below.

## 4. Operating-Rules Preamble (parity)

Every handoff payload Claude writes for Codex (Pass-1, Pass-2, research, review instructions) MUST begin with this preamble so Codex operates under the same floors as Claude. This is the Claude→Codex half of instruction parity; the Codex-as-driver half lives in `~/.codex/skills/m-pipeline/`.

```
Operating rules for this task (non-negotiable floors):
- Verify before asserting. Every file, struct, function, table, column, or API
  signature you reference must come from a file you actually read. Recalled
  knowledge is not a citation. Quote code verbatim, not from memory.
- Cite exact file:line. Re-read to confirm line numbers before reporting.
- Self-challenge every finding: state the strongest counter-argument and trace
  it through real code. If the counter-argument holds, drop the finding. Prefer
  zero findings over weak findings. "I don't know" is a valid answer.
- No shortcuts: do not skip a check because the change looks small; do not
  substitute pattern recognition for tracing where inputs come from.
- Stay read-only. Do not modify files. Report findings only.
```

## 5. Secret Redaction Rule (applies to every handoff write)

Before writing `.m/handoff/claude-to-codex.md`, redact any value matching the deny-list below. Replace matched values with the literal string `[REDACTED]`. This applies to every handoff write (Pass-1, Pass-2, research, review).

- API keys: `sk-*`, `pk_*`, AWS `AKIA*` or `ASIA*`, Google `AIza*`
- ALL_CAPS environment variable values where the variable name matches `*_TOKEN`, `*_SECRET`, `*_PASSWORD`, or `*_KEY`
- Contents of any `.env*` file
- JWT tokens (tokens beginning `eyJ` with two `.` separators)
- PEM blocks (from `-----BEGIN * PRIVATE KEY-----` through the matching `-----END * PRIVATE KEY-----`)

Scope of handoff writes is plan/research/review prose only — decisions, file paths, acceptance criteria, rationale. Never include raw source code, `.env` contents, or log excerpts that may embed secrets. (Native `codex exec review` reads the repo directly under its own sandbox; the redaction rule governs only the prose payloads Claude authors.)

## 6. Metered Codex Invocation (used by every pass)

Every Codex pass in this protocol uses the single helper below. It emits the agent's final message to `.m/handoff/codex-to-claude.md`, captures the JSONL event stream for token metering, updates the run meter, and enforces the budget. Do not call `codex` outside this helper from within the pipeline.

```bash
# Inputs the caller sets before invoking:
#   CODEX_MODE         = "exec" | "review"
#   CODEX_PROMPT       = exec prompt text, OR review instruction text (review mode)
#   CODEX_REVIEW_FLAG  = "--uncommitted" | "--base <branch>" | "--commit <SHA>"  (review mode only)
# Plus the config vars from Section 1 and FAST_FLAGS from Section 3.

mkdir -p .m/handoff
EV=.m/handoff/codex-events.jsonl
OUT=.m/handoff/codex-to-claude.md
METER=.m/handoff/codex-meter.txt
[ -f "$METER" ] || echo 0 > "$METER"

# Budget pre-check: if already at/over budget, do not spend more (see Section 7).
SPENT=$(cat "$METER" 2>/dev/null || echo 0)
if [ "$CODEX_DISABLED" = "true" ] || [ "$SPENT" -ge "$CODEX_BUDGET" ]; then
  echo "[codex] budget reached or disabled — skipping pass (Claude-only)."
else
  if [ "$CODEX_MODE" = "review" ]; then
    printf '%s' "$CODEX_PROMPT" | codex exec review $CODEX_REVIEW_FLAG \
      -c model="$CODEX_MODEL" -c model_reasoning_effort="$CODEX_EFFORT" "${FAST_FLAGS[@]}" \
      -o "$OUT" --json --skip-git-repo-check - > "$EV" 2>&1
    RC=$?
  else
    codex exec \
      -m "$CODEX_MODEL" -c model_reasoning_effort="$CODEX_EFFORT" "${FAST_FLAGS[@]}" \
      -s read-only -C "$PWD" -o "$OUT" --json --ephemeral --skip-git-repo-check \
      "$CODEX_PROMPT" > "$EV" 2>&1
    RC=$?
  fi

  if [ "$RC" -ne 0 ]; then
    echo "[WARN] codex exec exited $RC — proceeding Claude-only for the rest of this run."
    CODEX_DISABLED=true
    rm -f .m/handoff/claude-to-codex.md
  else
    # Meter: tokens spent this pass. codex-cli >=0.137 emits per-turn usage on
    # each `turn.completed` event (input_tokens + output_tokens) and no longer
    # carries a `total_tokens` field; older builds emitted a single
    # `total_tokens` under payload.info.total_token_usage. Sum the turn.completed
    # usage and fall back to the legacy field so both schemas are covered.
    THIS=$(python3 - "$EV" <<'METERPY'
import json, sys
turn_total = 0
legacy_max = 0
try:
    for line in open(sys.argv[1]):
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except ValueError:
            continue
        if obj.get("type") == "turn.completed":
            u = obj.get("usage") or {}
            turn_total += (u.get("input_tokens") or 0) + (u.get("output_tokens") or 0)
        info = ((obj.get("payload") or {}).get("info") or {})
        tot = (info.get("total_token_usage") or {}).get("total_tokens")
        if isinstance(tot, int):
            legacy_max = max(legacy_max, tot)
except OSError:
    pass
print(turn_total or legacy_max or 0)
METERPY
)
    [ -z "$THIS" ] && THIS=0
    TOTAL=$(( SPENT + THIS ))
    echo "$TOTAL" > "$METER"
    echo "[codex] this pass: ${THIS} tok | run total: ${TOTAL}/${CODEX_BUDGET} tok"
    # Persist the real Codex rate-limit/usage snapshot for the statusline
    # (global, cross-run): payload.rate_limits of the last token_count event.
    python3 - "$EV" "$CODEX_MODEL" <<'PYEOF'
import json, sys, os, time
events, model = sys.argv[1], sys.argv[2]
snap = None
try:
    for line in open(events):
        try:
            obj = json.loads(line)
        except ValueError:
            continue
        payload = obj.get("payload") or {}
        if payload.get("type") == "token_count" and payload.get("rate_limits"):
            snap = payload["rate_limits"]
except OSError:
    pass
if snap:
    try:
        json.dump({"model": model, "ts": int(time.time()), "rate_limits": snap},
                  open(os.path.expanduser("~/.claude/.codex-limits.json"), "w"))
    except OSError:
        pass
PYEOF
  fi
fi
```

After every successful pass, read `.m/handoff/codex-to-claude.md` for the agent's final message, then apply the Section 7 budget check before the next pass.

## 7. Token Metering and Budget Enforcement

The two-call round budget has been replaced by token metering. `.m/handoff/codex-meter.txt` accumulates the cumulative Codex token total across every pass in a single `/m` run (plan Pass-1 + Pass-2, research, review). The budget (`CODEX_BUDGET`, default 200000) is the ceiling.

After each pass updates the meter, evaluate:

- If `TOTAL < CODEX_BUDGET`: continue normally.
- If `TOTAL >= CODEX_BUDGET`, apply `CODEX_ON_EXCEED`:
  - **`fallback`** (default): print verbatim `[codex] token budget ${CODEX_BUDGET} reached — finishing this run Claude-only.`, set `CODEX_DISABLED=true`, and complete all remaining passes Claude-only. The work finishes; Codex cost is capped.
  - **`stop`**: print verbatim `[codex] token budget ${CODEX_BUDGET} reached — stopping and saving progress.`, persist current state (the in-progress plan/review/research draft and any updated `.m/` files), write a one-line resume note to `.m/PROGRESS.md` describing what remains, run Handoff Cleanup (Section 13), and halt the command. Do not silently finish; the user decides whether to continue.

The meter is per run. The first pass of a fresh `/m` command initializes it to `0` (Section 2). Handoff Cleanup (Section 13) removes the meter file at every terminal path so the next run starts clean.

## 8. Pass-1: Architecture Sanity (plan, blocking)

Runs after observation gathering completes, before Phase 2 begins. Phase 2 must not start until Pass-1 completes or is skipped (skipped only when `CODEX_DISABLED=true`).

1. Run `mkdir -p .m/handoff` via Bash.
2. Build the handoff payload: the Operating-Rules Preamble (Section 4), then the raw request, every `[OBSERVATION]` entry gathered above, plus this instruction:

   > Review these observations against the repository. Identify missing architecture concerns, risks, or alternative approaches that the driver agent has not surfaced. Do not propose a full plan yet. Return a bulleted list of additional concerns, each with a concrete `file:line` reference when applicable.

3. Apply the Secret Redaction Rule (Section 5), then write the result to `.m/handoff/claude-to-codex.md`.
4. Set `CODEX_MODE=exec` and `CODEX_PROMPT="$(cat .m/handoff/claude-to-codex.md)"`, then run the Metered Codex Invocation (Section 6).
5. Read `.m/handoff/codex-to-claude.md`.
6. Merge each Codex-surfaced concern into the observation list as a new `[OBSERVATION — codex]` entry, preserving any `file:line` references verbatim.
7. Apply the Section 7 budget check. Leave the handoff files in place; they are cleaned up at the end of Phase 3.

The merged observation list (Claude's original `[OBSERVATION]` entries plus the new `[OBSERVATION — codex]` entries) is the input to Phase 2.

## 9. Pass-2: Final Plan Review (plan, blocking)

Runs after the three exit-gate checks pass, before the plan document is emitted. The plan is not emitted until Pass-2 completes or is skipped.

1. Build the handoff payload: the Operating-Rules Preamble (Section 4), the full final plan draft (summary, architecture decisions, tasks with acceptance criteria, risks), plus this verdict-format instruction appended at the end:

   > End your output with a line of exactly `VERDICT: LGTM` if you have no disagreements with this plan, or `VERDICT: CHANGES` if you do. If `VERDICT: CHANGES`, list each disagreement above the verdict line as a numbered item: `Dn. <short description> | claude: <claude-pos + one-line rationale> | codex: <codex-pos + one-line rationale>`. Include concrete `file:line` references where applicable.

2. Apply the Secret Redaction Rule. Write it to `.m/handoff/claude-to-codex.md` (overwriting any Pass-1 content).
3. Set `CODEX_MODE=exec` and `CODEX_PROMPT="$(cat .m/handoff/claude-to-codex.md)"`, then run the Metered Codex Invocation (Section 6).
4. Read `.m/handoff/codex-to-claude.md`. Parse the final non-empty line:
   - If it is exactly `VERDICT: LGTM`: finalize tasks, run Handoff Cleanup, emit the plan.
   - If it is exactly `VERDICT: CHANGES`: parse the numbered disagreements above it and enter the Disagreement Menu (Section 10).
   - If neither verdict string is present on the last non-empty line: treat as `VERDICT: CHANGES` (conservative default). If no parseable disagreements are listed, hard-block and re-grill the user on the final plan content.
5. Apply the Section 7 budget check.

If `CODEX_DISABLED=true` (disabled, CLI failure, or budget reached): skip Pass-2, finalize tasks, run Handoff Cleanup, emit the plan.

## 10. Disagreement Menu (plan)

For each numbered disagreement `Dn` Codex listed, present a single bounded-menu question to the user. Source tags are mandatory and must appear verbatim:

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

After all picks are applied, run one final Codex verification pass if the token budget still permits it:

- If `CODEX_DISABLED=true` or the meter is at/over budget: skip the final verification, run Handoff Cleanup, emit the plan.
- Otherwise: write the merged plan to `.m/handoff/claude-to-codex.md` (with redaction and preamble), run the Metered Codex Invocation, parse the verdict.
  - `VERDICT: LGTM` → run Handoff Cleanup, emit the plan.
  - `VERDICT: CHANGES` → **hard-block**. Convert the remaining disagreements into new Phase 2 grill questions. Do NOT invoke Codex again for the remainder of this run (set `CODEX_DISABLED=true`). Loop back to Phase 2.

## 11. Research: Parallel Dual-Engine Researcher

Used by `/m:research` (and by `/m:plan`'s worktree research spawn) when `CODEX_ENABLED` is true. Codex researches the same questions independently and in parallel with Claude's research agent; Claude then reconciles both into one finding set. This mirrors the plan dual-engine model.

1. After forming the 2–5 focused research questions, write them (plus relevant file paths and the Operating-Rules Preamble) to `.m/handoff/claude-to-codex.md`, redacted per Section 5, with this instruction appended:

   > Research these questions against the local repository, official documentation, and primary sources. Separate stable local observations, external facts (with versions/dates), and recommendations. For each question, give a decisive recommendation plus the strongest counter-argument. Cite `file:line` for local claims and source URLs for external claims. List remaining unknowns explicitly.

2. Run Claude's research agent (the existing `Agent(isolation: "worktree")` flow) and the Metered Codex Invocation (Section 6, `CODEX_MODE=exec`) concurrently when possible.
3. Read `.m/handoff/codex-to-claude.md`. **Reconcile**:
   - Where both engines agree, mark the finding `[CORROBORATED]`.
   - Where they disagree, present both positions side-by-side under the relevant research question — `[claude]` vs `[codex]` — and do not silently pick a winner. The user decides what to adopt (research is advisory).
   - Merge unique findings from each engine, preserving citations verbatim.
4. The reconciled set is the research output. Apply the Section 7 budget check; if the budget is reached mid-research, finish reconciliation with whatever Codex returned and continue Claude-only per `CODEX_ON_EXCEED`.

If `CODEX_DISABLED=true`, run Claude-only research exactly as before.

## 12. Review: Mandatory Second Engine

Used by `/m:review` and `/m:review-fanout`. When `CODEX_ENABLED` is true, Codex review runs on **every** review — it is no longer gated behind a high-stakes category or a `y/n` prompt. (The high-stakes categories still drive Claude's own pass depth and the compliance pass; they no longer gate Codex.)

1. After Claude's primary review produces its verdict, set `CODEX_MODE=review` and the target flag:
   - `CODEX_REVIEW_FLAG="--uncommitted"` for staged + unstaged + untracked changes
   - `CODEX_REVIEW_FLAG="--base <branch>"` to review against a base branch
   - `CODEX_REVIEW_FLAG="--commit <SHA>"` for a specific commit
2. Set `CODEX_PROMPT` to a constraining instruction that begins with the Operating-Rules Preamble (Section 4), for example:

   > Focus on: correctness bugs, security threats, regressions, and clean-code violations introduced by these changes. Report each finding with `file:line` and a verbatim code snippet only. Do not restate unchanged code.

3. Run the Metered Codex Invocation (Section 6). Native `codex exec review` reads the diff directly under its own sandbox; no source is written into the handoff payload.
4. Present Claude's findings and Codex's findings side-by-side under a dedicated **Second Opinion (Codex)** section. Do not merge silently. For `/m:review-fanout`, place this alongside the judge verdict, not inside the judge's merged list.
5. **Disagreement rule:** if Claude says `APPROVED` but Codex flags criticals that survive re-verification (the Step 3 / judge hallucination filter applied to Codex's findings too), downgrade to `BLOCKED`. The more permissive verdict never wins by default.
6. Apply the Section 7 budget check. If `CODEX_DISABLED=true` (disabled, CLI failure, or budget reached), note it in the Review Metadata and skip the second engine — do not fail the review.

## 13. Handoff Cleanup

At the end of every command that used this protocol — on every terminal path (successful emit, pre-flight fallback, mid-run Codex failure, budget stop, or hard-block re-grill exit) — delete the handoff and meter files if they exist:

```bash
rm -f .m/handoff/claude-to-codex.md .m/handoff/codex-to-claude.md \
      .m/handoff/codex-events.jsonl .m/handoff/codex-meter.txt
```

Do not persist any dialogue artifact. Only the final plan/research/review output (chat or the relevant `.m/*.md` file) remains.
