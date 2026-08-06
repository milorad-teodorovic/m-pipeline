---
description: Implement an approved plan or a direct request using repo patterns. Use after /m:plan completes or for clear direct asks. Writes code.
argument-hint: [plan-or-request]
model: claude-opus-5
effort: xhigh
disable-model-invocation: false
---
# /m:implement - Implementation Workflow

Implement the approved plan or a clear direct request.

## Input

Approved plan or implementation instructions: `$ARGUMENTS`

## Jira Context (run before implementation)

If `$ARGUMENTS` contains a Jira reference, resolve and fetch it per `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md` **before** starting implementation. Treat Jira acceptance criteria as the verification target; if the implementation cannot satisfy all of them, list the gaps under **Deviations**.

## Context Sources

Read these first when available:

- `.m/jira.yml` (per-project Jira mapping)
- `.m/INDEX.md`
- `.m/PLAN.md` (the approved implementation plan, when the project persists one — the written handoff from `/m:plan` for a standalone `/m:implement` run that does not share the planning session's context)
- `.m/PROGRESS.md`
- `.m/GAPS.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`
- `~/.claude/m-learning/ADAPTATIONS.md` (if present) — apply the HIGH and MEDIUM `implement` adaptations and the `test_approach` preference recorded there; proceed normally if it does not exist. Current-session instructions always override a learned adaptation.

## Workflow

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, before any code mutation: run
   `mkdir -p .m && touch .m/phase-implement-started` via Bash.
2. On successful completion (all planned tasks implemented and any
   proportionate verification green): run `touch .m/phase-implement-done`.
3. On abort, unrecoverable test failure, or mid-run hand-off: leave
   `-started` in place and do NOT write `-done`.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `implement`, stop and tell the user — the pipeline is out of sync.

1. Confirm the requested scope and the relevant plan
2. Identify the touched modules, languages, test surfaces, and verification limits
3. Implement in dependency order
4. Run proportionate verification as you go
5. Update `.m/PROGRESS.md` with meaningful status changes

## Side-Effect Tier

Before starting, classify the implementation scope:

- **read-only** (analyze, research, search): no confirmation needed
- **write-local** (create/edit files, run tests): proceed after confirming scope
- **write-external** (DB migrations, API changes, config that affects shared systems): confirm each step explicitly with the user before executing

This classification drives how aggressively to proceed vs pause for confirmation. Default to `write-local` when unclear.

## Mode Selection

- Small and clear: implement directly
- Medium: implement directly, but keep explicit checkpoints and self-review
- Large or multi-surface: you may use one or more focused subagents only if the runtime supports them cleanly and file ownership can stay clear; otherwise implement directly

## Implementation Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (read in full before proceeding): run the test runs the plan calls for, stay on `[CONFIRMED]` plan elements, escalate only genuine plan defects.
- Match established project patterns exactly
- Reuse shared types, components, utilities, and services before creating new ones
- Keep the change to the minimum that satisfies the request. Do not add features, refactor adjacent code, or make "improvements" beyond what was asked — a bug fix does not need the surrounding code cleaned up. The right amount of complexity is the minimum needed for the current task; do not introduce abstractions, configuration knobs, or defensive layers the request did not call for. Validate input only at system boundaries such as user input and external API responses, not at every internal call site.
- Solve the problem, not the test. Implement a solution that works correctly for all valid inputs, not only the cases the tests cover, and never hard-code expected values to make a test pass. Tests verify correctness; they do not define the solution. If a test itself looks wrong, flag it instead of coding to satisfy it.
- Prefer tests first when proportionate, especially for business logic, auth, money, parsing, or data integrity
- For Go: use existing error and logging patterns, validate input at the boundary, and avoid raw SQL string building with user input
- For React and frontend work: preserve the current design system and layout language unless the user explicitly asked for a redesign
- If the repo is an extracted snapshot or is missing build metadata, prefer minimal, auditable changes and state what could not be verified
- If you encounter ambiguity in a task that requires a design decision, STOP and escalate back to the plan stage as a plan defect. Flag plan defects directly — if the plan itself looks wrong, say so. The plan stage owns design decisions; implement stays grounded in confirmed patterns rather than inferred ones.
- Track deviations from the plan instead of silently drifting
- Emit a brief progress line after each meaningful checkpoint:
  `[impl] {module}: {what was done} ({files_touched} files)`
  This keeps the user oriented during longer implementations without waiting for the final summary

## Output

Finish with:

## Implementation Summary

### Scope
### Files Changed
### Tests Run
### Deviations
### Ready For

If the work is incomplete or blocked, say so directly.
