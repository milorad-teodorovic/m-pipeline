---
description: Run the full /m delivery pipeline end-to-end (refine → plan → implement → review → iterate). Use when user wants a complete request delivered with quality gates, dual-engine review, and phase enforcement.
argument-hint: [request]
model: claude-opus-5
effort: xhigh
disable-model-invocation: true
---
# /m:develop - End-to-End Delivery Workflow

Run the full `/m:*` pipeline for the current request.

## Invocation Protocol (HARD GATE — read before doing anything else)

Every phase in this pipeline runs as a discrete Skill tool call. You may not
inline a phase's output into chat under any circumstances. If you catch
yourself about to emit a refined spec, an implementation plan, an
implementation narrative, a review report, or an iterate verdict without a
prior Skill call for that phase, stop immediately, delete the draft, and
invoke the skill.

The harness enforces this via a marker-file protocol that a PreToolUse hook
validates on every Edit, Write, and MultiEdit. Skipping a skill now causes
file mutations to be blocked until the skill is invoked.

### Marker files

- `.m/DEVELOP_ACTIVE` — single-line marker written at pipeline entry,
  deleted at pipeline exit. Contents:
  ```
  current_phase: <refine|plan|implement|review|iterate>
  ```
- `.m/phase-<name>-started` — created by each phase skill on entry.
- `.m/phase-<name>-done` — created by each phase skill on successful exit.

### Pipeline entry

Immediately after the user invokes `/m:develop`, before anything else:

1. Ensure `.m/` exists (`mkdir -p .m`).
2. Write `.m/DEVELOP_ACTIVE` with `current_phase: refine`.
3. Remove any stale `.m/phase-*-started` / `.m/phase-*-done` markers from
   a prior pipeline run.

### Phase transitions

Before moving from phase A to phase B:

1. Verify `.m/phase-<A>-done` exists. If missing, re-enter phase A via its
   skill — do NOT proceed.
2. Overwrite `.m/DEVELOP_ACTIVE` with `current_phase: <B>`.
3. Invoke `Skill(skill="m:<B>")` as the first action of phase B.

### Pipeline exit

On every terminal path (success, hard-block, abort, user interrupt):

1. Delete `.m/DEVELOP_ACTIVE`.
2. Leave the `phase-*-started` and `phase-*-done` markers in place so they
   are auditable; they will be cleaned up on the next pipeline entry.

### Hook blocking behavior

When `.m/DEVELOP_ACTIVE` is present and the current phase has not been
entered via its skill, any Edit/Write/MultiEdit outside `.m/` is denied by
the `enforce-develop-phase.py` PreToolUse hook. Writes under `.m/` are
always allowed so the protocol itself can run.

## Stage Order

The pipeline runs these stages in sequence. Each stage produces input the next stage depends on, which is what keeps the pipeline cheap:

```
index (if needed) → refine → plan → classify → implement → review → iterate → update
```

**Refine runs first.** Invoke `/m:refine` via the Skill tool before any other work. Refine surfaces the assumptions that plan and implement would otherwise guess at, so its output is what makes the downstream stages cheap.

**Plan runs after refine.** Invoke `/m:plan` via the Skill tool once refine completes. Plan grounds implementation in user-confirmed decisions, which is what prevents impl-time improvisation.

**Checkpoint line.** Before each stage transition, emit a line confirming the previous stage ran:
`[checkpoint] {stage_name} complete — output: {key artifact or verdict}`
If the line cannot be written because the stage produced no output, pause and run the missing stage before moving on.

## Input

Request to deliver: `$ARGUMENTS`

## Context Sources

Read these when available (reading context does NOT replace running refine/plan):

- `.m/INDEX.md`
- `.m/TASKS.md`
- `.m/PROGRESS.md`
- `.m/GAPS.md`
- `.m/RESEARCH.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`

## Workflow

0. **INDEX (if needed)** — If `.m/` is missing any core index file or is clearly stale, run `/m:index` first via the Skill tool. This runs before pipeline entry so the index exists before any phase marker or `.m/DEVELOP_ACTIVE` is written.
1. **ENTER PIPELINE** — follow the Invocation Protocol entry steps: create `.m/`, write `.m/DEVELOP_ACTIVE` with `current_phase: refine`, remove stale `phase-*` markers.
2. **REFINE** — Invoke `Skill(skill="m:refine")`. This is the grill stage. Skipping it silently collapses the spec to a lowest-common-denominator version. Wait for the skill to complete and write `.m/phase-refine-done` before proceeding. Do not emit any refined spec content inline.
3. **PLAN** — Before invoking, verify `.m/phase-refine-done` exists. Update `.m/DEVELOP_ACTIVE` to `current_phase: plan`. Invoke `Skill(skill="m:plan")` immediately. Do not prompt the user between refine and plan. Do not emit any plan content inline. Plan may BLOCK and spawn worktree-isolated `/m:research` when it encounters unknowns.
4. Classify the side-effect tier: `read-only`, `write-local`, or `write-external`.
5. **IMPLEMENT** — Verify `.m/phase-plan-done` exists. Update `.m/DEVELOP_ACTIVE` to `current_phase: implement`. Invoke `Skill(skill="m:implement")`. All code mutation happens inside this skill.
6. **REVIEW** — Verify `.m/phase-implement-done` exists. Update `.m/DEVELOP_ACTIVE` to `current_phase: review`. Invoke the appropriate review skill (see Review Selection below) via the Skill tool. The second engine runs in-stage on every review when `second_engine.provider` is `codex` or `kimi`.
7. **ITERATE** — Verify `.m/phase-review-done` exists. Update `.m/DEVELOP_ACTIVE` to `current_phase: iterate`. Invoke `Skill(skill="m:iterate")`. Run until the **exit predicate** is satisfied (tests green + zero critical review findings + `.m/PROGRESS.md` updated + PRD Success Criteria satisfied when a `.m/PRD-*.md` exists) or the 3-loop safety cap is reached. `PASSED` requires the predicate, not just the cap.
8. **EXIT PIPELINE** — Delete `.m/DEVELOP_ACTIVE`. Update `.m/TASKS.md`, `.m/PROGRESS.md`, and `.m/GAPS.md` with the outcome. Then append one outcome signal to `~/.claude/m-learning/signals/outcomes.jsonl` (create the file if absent) so `/m:learn` can derive adaptations — a single JSON line of the form `{"timestamp":"<ISO-8601>","type":"outcome","skill":"develop","request":"<one-line summary>","stages":"<e.g. refine→plan→implement→review-fanout→iterate>","verdict":"<PASSED|BLOCKED>","loops":<n>,"second_engine":"<codex:agree|codex:disagree|kimi:agree|kimi:disagree|n/a>"}` (use the `timestamp` key to match the existing signal schema). Append with the file tools, not `echo`. This is passive telemetry and is separate from the opt-in `/m:feedback` signals.

Emit a stage transition line between each step:
`[pipeline] {stage_name} → {next_stage_name} ({reason or key outcome})`

For the review stage, the transition line must also state which review variant ran and whether a second engine was invoked:
`[pipeline] implement → review-fanout ({N} lenses) → iterate (second opinion — codex: agree)`

## Review Selection

Pick the review command based on the change shape, not by default:

| Change shape | Command |
|---|---|
| 1–3 files, single lane, <100 lines changed | `/m:review` |
| 4–10 files OR touches 2+ layers (handler → service → repo) | `/m:review-fanout` |
| Go backend with security-sensitive code paths | `/m:review` (delegates security pass to `go-security-reviewer`) OR `/m:review-fanout` with the security + architecture lenses |
| Cross-stack change (Go + React + infra) | `/m:review-fanout` |
| Repo whose `.m/pipeline.yml` sets `compliance.enabled: true` | whichever command is already running **plus** the compliance pass |

Compliance and high-stakes triggers are read from the repo's `.m/pipeline.yml` (schema: `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`). If the file is absent, compliance is off and only the generic high-stakes categories in the next section apply.

## Second Engine (mandatory when a provider is selected)

Second-engine participation across the pipeline (plan, research, review) is driven by the `second_engine:` section of the repo's `.m/pipeline.yml` (schema, provider defaults, and the legacy `codex:` fallback: `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`; `provider: none` is the default — Claude-only).

When the provider is `codex` or `kimi`, the review stage runs the second engine automatically on **every** review — there is no `y/n` prompt and no high-stakes gating. Follow the active provider's protocol Section 12 (`${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` or `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md`):

1. After the review stage produces its verdict, run the Metered Invocation against the change set (Codex: native `codex exec review` with `--uncommitted`, `--base <branch>`, or `--commit <SHA>`; Kimi: the prompt-constructed review over the same target).
2. Present Claude's findings and the second engine's findings **side-by-side**. Do not merge silently. Second-engine findings are leads to confirm, never ground truth.
3. **Disagreement rule:** if Claude says `APPROVED` but the second engine flags criticals that survive re-verification, downgrade to `BLOCKED`. The more permissive verdict never wins by default.
4. The second engine is skipped (noted in metadata, never a pipeline failure) only when the provider is `none`, the CLI is unavailable, or the per-run token budget is reached (per `on_budget_exceeded`).

The high-stakes categories below still escalate Claude's own pass depth, size tier, and compliance pass; they do not gate the second engine.

## Size Selection

- **Small** (1–3 files, unambiguous scope): `refine → plan → implement → review → iterate`. Use `/m:review`.
- **Medium** (4–10 files OR multi-layer): `refine → plan → implement → review → iterate`. Use `/m:review-fanout` instead of `/m:review` if the change crosses 2+ layers.
- **Large or security-sensitive** (10+ files, migrations, auth/money, `high_stakes_paths` matches, API contracts): `refine → plan (with worktree research as needed) → implement → review-fanout (second engine runs in-stage when a provider is selected) → iterate`.

Every size runs the full pipeline: refine → plan → implement → review → iterate. No stage is optional. No skips. No "trivial enough" bypass. You must invoke `/m:refine` and `/m:plan` as Skill tool calls — not inline them, not summarize them, not skip them. Refine auto-chains to plan in all sizes. Plan handles research internally via BLOCK + worktree spawn when it encounters unknowns.

## Side-Effect Tiers

Inherit the tier from `/m:implement`. The tier gates how the pipeline proceeds:

- **read-only**: skip confirmation, run full pipeline automatically
- **write-local**: proceed after initial scope confirmation, no per-step gates
- **write-external** (migrations, API contracts, shared config): pause before implement and before any destructive step — confirm with user explicitly

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (loaded at session start) at every stage: no shortcuts, full tool use, `[USER-INTENT]`-only questions
- Treat critical or high-risk review and verification issues as gates, not soft suggestions
- `/m:iterate` may only emit `PASSED` when its four-clause exit predicate is green. A loop-count exit is `BLOCKED`, not `PASSED`
- The second engine runs automatically on plan, research, and review when `second_engine.provider` is `codex` or `kimi`; it is config-driven, not prompted. The default is `none` (Claude-only)
- When Claude's judge and the second engine disagree, the stricter verdict wins
- Do not auto-create worktrees
- Do not redesign UI unless the user explicitly asks
- Keep CURRENT issues separate from PRE-EXISTING gaps
- If the repo is an extracted snapshot, vendor dump, or is missing build/test metadata, call out the verification limits explicitly
- Apply `${CLAUDE_PLUGIN_ROOT}/rules/verification.md` at every stage. Self-challenge every finding before emitting it

## Output

Finish with:

## Delivery Summary

### Request
### Pipeline Stages Run
List the stages actually run, in order, with the specific review variant and whether Codex was invoked.
Example: `refine → plan → implement → review-fanout (7 lenses) → codex second-opinion (agree) → iterate (2 loops, PASSED)`
### Scope
### Checks Run
### Exit Predicate Status
- Tests green: yes / no — {command and exit code}
- Zero critical review findings: yes / no / n/a
- PROGRESS.md updated: yes / no
- PRD Success Criteria satisfied: yes / no / n/a
### Remaining Blockers
### Next Step
