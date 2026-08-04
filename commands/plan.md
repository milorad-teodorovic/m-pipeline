---
description: Create an actionable implementation plan grounded in repo patterns and user-confirmed decisions. Uses second-engine (Codex or Kimi, config-driven) sanity passes and a grill loop until zero gaps. Use after /m:refine or when starting a non-trivial change.
argument-hint: [refined-request]
model: claude-opus-5
effort: xhigh
allowed-tools: Read, Grep, Glob, Bash, Agent, TaskCreate
---
# /m:plan - Master Planner

Create a detailed, actionable implementation plan through rigorous gap analysis and user-confirmed decisions. The plan stage is the primary value producer in the pipeline.

## Core Constraint: Trace Every Plan Element to a User Statement

**Every plan element traces to an explicit user statement from this session, or to the user-authored refined specification supplied as this plan's input.** This is the anchor for the rest of the file.

- Architecture preference unstated → BLOCK and ask.
- Error handling strategy unspecified → BLOCK and ask.
- Codebase pattern unconfirmed for this change → BLOCK and ask.
- "Probably needs X" is an observation. "User confirmed X in round 2" is a plan element.

When you don't know, say "I don't know" and BLOCK. The plan grounds unknowns in user confirmation rather than inferred patterns, hallucinated architecture, or training-data defaults.

## Input

Refined prompt or task description: `$ARGUMENTS`

## Jira Context (run before Phase 1)

If `$ARGUMENTS` contains a Jira reference, resolve and fetch it per `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md` **before** codebase analysis. Ground acceptance criteria in the Jira acceptance criteria; if the story is thin, call that out under **Risks or Open Items**.

## Context Sources

Read these first when available:

- `.m/jira.yml` (per-project Jira mapping)
- `.m/INDEX.md`
- `.m/GAPS.md`
- `.m/RESEARCH.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`
- `~/.claude/m-learning/ADAPTATIONS.md` (if present) — apply the HIGH and MEDIUM `plan` adaptations and pipeline defaults recorded there; proceed normally if it does not exist. Current-session instructions always override a learned adaptation.

Always prefer established repo patterns over invention — but only when the user confirms the pattern applies.

## Workflow

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, before any codebase analysis or Codex pre-flight: run
   `mkdir -p .m && touch .m/phase-plan-started` via Bash.
2. On successful completion (plan emitted, exit gate passed): run
   `touch .m/phase-plan-done`.
3. On BLOCK (for gaps, research spawn, or hard-block disagreement): leave
   `-started` in place and do NOT write `-done` until the plan is
   eventually emitted.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `plan`, stop and tell the user — the pipeline is out of sync.

### Phase 1: Codebase Analysis

#### Pre-flight: Second-Engine Check

Before any observation work, resolve `second_engine` from `.m/pipeline.yml` (schema, provider defaults, and legacy `codex:` fallback in `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`) and run the pre-flight check in the active provider's protocol Section 2 — `${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` for `codex`, `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md` for `kimi`. When a provider is selected, Pass-1 and Pass-2 are **mandatory** — there is no per-pass permission prompt. The passes are skipped (the plan proceeds Claude-only) only when the provider is `none`, the CLI is unavailable, or the per-run token budget is reached.

The Metered Invocation (Section 6), Operating-Rules Preamble (Section 4), Secret Redaction Rule (Section 5), and Token Metering (Section 7) of the active provider's protocol apply to every Pass-1 and Pass-2 handoff. Do not duplicate those rules here — read the reference and apply them verbatim.

#### Observation Gathering

Explore the codebase — read the request, map relevant code paths, classify repo health, identify existing patterns. When the codebase map would span more than three searches, spawn `Explore` subagent(s) (`model: haiku`) for the breadth sweep; keep the synthesis, grilling, and the worktree `/m:research` spawn on the orchestrator's `opus`.

Present each finding as a one-line `[OBSERVATION]` entry that names the concrete file or path it came from.

Observations are input for the grill. They do NOT become plan elements without user confirmation.

#### Pass-1: Second-Engine Architecture Sanity (blocking)

Runs after observation gathering completes, before Phase 2 begins. Phase 2 must not start until Pass-1 completes or is skipped. Follow the active provider's protocol Section 8 for the full protocol — payload build, redaction, metered invocation, merge of `[OBSERVATION — codex]` / `[OBSERVATION — kimi]` entries.

### Phase 2: Grill-Based Plan Construction

Build the plan iteratively through user-confirmed decisions. This phase loops until all gaps are resolved.

**For each plan area** (architecture, file changes, data model, security, testing, error handling, implementation order):

1. Identify what the user HAS explicitly stated vs what is MISSING
2. Batch all missing items as gaps
3. Present gaps as bounded-menu questions:

```
BLOCKED — {N} gaps require your input

Q1. {short question about gap}
  A) {concrete option with its tradeoff}
  B) {concrete option with its tradeoff}
  C) {concrete option with its tradeoff}
  D) Other — describe

Q2. {next gap}
  ...
```

4. Wait for user answers
5. If answers create new gaps → new grill round. Loop until zero gaps. **Soft round cap:** there is no hard limit, but if the grill reaches round 4 and answers are still spawning fresh gaps, pause and surface a convergence check — show the user the still-open gaps and ask whether to keep grilling or to move the remainder to DEFERRED (under Risks) and proceed. This guards against a gap-spawns-gap loop without forcing premature closure; the user, not a counter, decides when to stop.

**Complete-input fast path.** When the refined spec settles every plan area and no genuine gap survives observation gathering, do not manufacture gaps: confirm the collapsed plan in one round, and when no requester round-trip is possible, record each spec-settled decision as `[CONFIRMED-BY-SPEC]` with its spec source and emit the plan — the spec is the user's statement.
6. When user confirms a section:
   - Mark items as `[CONFIRMED]`
   - Create a draft task via TaskCreate: high-level title + checklist sub-items in description
   - Each task must have acceptance criteria in the description

**Labeling protocol:**
- `[OBSERVATION]` — codebase finding, not a decision
- `[PROPOSED]` — Claude's suggestion, needs user confirmation before becoming plan
- `[CONFIRMED]` — user confirmed, now a plan element. Include which grill round confirmed it.
- `[CONFIRMED-BY-SPEC]` — settled by the user-authored refined spec when no requester round-trip is possible, now a plan element. Include the spec section it traces to.

**Anti-assumption enforcement:**
- When presenting a `[PROPOSED]` item, always include your reasoning AND the strongest counter-argument
- If you catch yourself writing a plan element without a `[CONFIRMED]` trace, stop and convert it to a gap question
- "The codebase already does X" is an observation, not a confirmation. The user must still confirm X applies to this change.

**Research trigger:** If you encounter an unknown requiring external research (unfamiliar library, protocol, integration point), BLOCK — announce `BLOCKED — need research on <topic>`, state that isolated worktree research is being spawned, and spawn it.

Spawn via `Agent(isolation: "worktree")` with ONLY the research question and relevant file paths. Do NOT include the refined spec, plan-so-far, or any conversation context in the agent prompt. Present research findings to user. User decides what to incorporate — research is advisory, the plan (user's plan) wins.

### Phase 3: Exit Gate

Before emitting the final plan, verify ALL of the following:

1. **Every task has acceptance criteria** — specific, testable, complete
2. **Zero impl-time decisions** — no task requires the implementer to choose an approach, pick a pattern, or decide on error handling. The metric is decision count, not file count. A 3-file change with zero ambiguity passes. A 1-file change with "pick auth strategy" fails.
3. **All gaps resolved** — or user said "enough" (remaining gaps become DEFERRED items under Risks)

If any check fails:
- Identify failing items
- Convert to new grill round questions
- Loop back to Phase 2

If all checks pass, run Pass-2 (below) before emitting the plan document.

**Task splitting rule:** Before finalizing, check every task against the zero-decisions constraint. If a task requires the implementer to make any design decision, split it or escalate the missing decision as a new gap.

#### Pass-2: Second-Engine Final Plan Review (blocking)

Runs after the three exit-gate checks pass, before the plan document is emitted. The plan is not emitted until Pass-2 completes or is skipped. Follow the active provider's protocol Section 9 for the full Pass-2 protocol, Section 10 for the Disagreement Menu, Section 7 for token metering and budget enforcement, and Section 13 for the handoff cleanup that runs on every terminal path.

## Output

Produce a slim context document. All actionable content lives in tasks.

## Implementation Plan

### Summary
One paragraph: what we're building, why, and the key decisions that shape the approach.

### Architecture Decisions
For each major decision:
- **Decision**: what was chosen
- **Confirmed**: which grill round, user's exact choice
- **Rationale**: why this approach
- **Alternatives rejected**: what else was evaluated

### Risks or Open Items
- Active risks with mitigation
- DEFERRED items (gaps user chose not to resolve now)
- Verification limits (missing tests, build configs, etc.)

All other plan content (file changes, implementation steps, test strategy, error handling) lives in the tasks created during the grill.

## Persistence

- If `.m/PLAN.md` already exists for the current task, update it surgically
- Otherwise keep the plan in chat unless the user explicitly asks to persist it

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (loaded at session start). The grill loop and the second-engine Pass-1/Pass-2 checks are mandatory, not friction to optimize away.

## Self-Check

Before finishing, verify:

- every acceptance criterion from the refined spec is addressed by at least one task
- the plan follows existing patterns (confirmed by user, not assumed)
- shared utilities were considered
- restricted files, migrations, contracts, or generated code are flagged
- the implementation order encoded in tasks is dependency-safe
- every `[CONFIRMED]` item traces to an explicit user statement, and every `[CONFIRMED-BY-SPEC]` item cites the spec section that settles it

End by asking whether to proceed to `/m:implement` or adjust the plan.
