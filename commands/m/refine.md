---
description: Turn a raw request into an execution-ready PRD via active grilling. Use when user says "ask questions", "examine", "challenge this", "what am I missing", "stress test", "let's align", or wants a refined spec before planning or implementation.
argument-hint: [request]
model: opus
effort: high
---
# /m:refine - Request Refinement (Grill Stage)

Turn a raw request into an implementation-ready specification by **actively grilling** the request — not defensively checking it. This is the highest-leverage stage in the pipeline: a weak refine silently collapses plans to the lowest-common-denominator solution, and a strong one makes every downstream stage cheaper.

The posture is *"surface ambiguity proactively"*, not *"ask only when something would waste time"*. The first draft of any spec is almost always wrong in ways the requester cannot see yet.

## Input

Raw request: `$ARGUMENTS`

## Jira Context (run before workflow phases)

If `$ARGUMENTS` contains a Jira reference (full `*.atlassian.net/browse/KEY` URL, or a bare `KEY` that matches `.m/jira.yml` `projectKey`), fetch the story via the `atlassian` MCP server **before** Phase 1. Follow the shared rules in `${CLAUDE_PLUGIN_ROOT}/references/jira-context.md`.

- Load `.m/jira.yml` if present for `site`, `projectKey`, and `branchPattern`.
- Use `mcp__atlassian__*` tools to fetch: summary, description, status, acceptance criteria, recent comments.
- If the `atlassian` MCP is not installed or not authenticated, stop and instruct the user to run `/mcp` (or `claude mcp add --transport http --scope user atlassian https://mcp.atlassian.com/v1/mcp` if missing entirely).
- Prepend a **Jira Context** block (key, title, link, status, 2–4 line summary) to the final output.
- Treat the Jira story as authoritative requirements. If `$ARGUMENTS` contradicts it, list conflicts under **Conflicts with Jira**.

## Context Sources

Read these first when available:

- `.m/jira.yml` (per-project Jira mapping)
- `.m/INDEX.md`
- `.m/GAPS.md`
- `.m/RESEARCH.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`
- `~/.claude/m-learning/ADAPTATIONS.md` (if present) — apply the HIGH and MEDIUM `refine` adaptations and code-style preferences recorded there; proceed normally if it does not exist. Current-session instructions always override a learned adaptation.

## Workflow

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, immediately after reading context sources: run
   `mkdir -p .m && touch .m/phase-refine-started` via Bash.
2. On successful completion (spec delivered, no BLOCK): run
   `touch .m/phase-refine-done`.
3. On abort, hard-block, or unresolved gap: leave `-started` in place
   and do NOT write `-done`. The pipeline will refuse to advance.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `refine`, stop and tell the user — the pipeline is out of sync.

### Phase 0: Optimal-Version Reframe (grill opener)

Before analyzing anything, ask the requester:

> *"If time and labor were not a consideration, what would the optimal version of this look like? Don't plan — just describe the end state."*

Their answer (or the absence of one) is the anchor for the rest of the refine. The goal is to surface the *real* target before assumptions collapse it. Per community research: default AI planning assumes a solo dev with two jobs and no scaffolds, so the first plan is always smaller than it should be. The optimal-version reframe forces a truer anchor.

Skip this phase only when the request is a trivial bug fix with an obvious scope.

### Phase 1: Analysis

Classify the request (bug/feature/refactor/review/research/analysis/infrastructure), then analyze:

1. **Hidden requirements** — edge cases, implicit assumptions, things the requester likely forgot
2. **Affected surfaces** — files, modules, APIs, data flows touched by this change
3. **Reuse opportunities** — existing patterns, utilities, shared code that should be leveraged
4. **Security and data impact** — attack surface changes, PII handling, auth implications
5. **Acceptance criteria** — concrete, testable criteria for done
6. **Risks** — what could go wrong or waste implementation time

Reuse known repo context from `.m/INDEX.md` and `.m/GAPS.md`.

### Phase 2: Grill — Bounded-Menu Clarifying Questions

**Self-Serve Gate (run BEFORE drafting any question).** Apply `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` to every candidate question:

- `[FACTUAL]` questions (struct fields, table names, file paths, signatures, framework defaults, Jira content, test results, config values, current state) MUST be resolved via Read, Grep, Glob, Bash, or MCP. Move the resolved answer into Technical Context. Do NOT ask the user.
- `[USER-INTENT]` questions (scope tradeoffs, business rules, preferences between equally valid options, deadlines, stakeholder context) become bounded-menu questions below.
- `[MIXED]` questions are split. Tool-resolve the factual half. Ask only the intent residue.

Every question that reaches the user MUST be prefixed `[USER-INTENT]` in the menu. If a question cannot wear the prefix cleanly, it is factual — go resolve it via tools instead.

**Mandatory for anything larger than a trivial fix.** Do not skip this phase to save turns. The 3–5 question floor applies to `[USER-INTENT]` questions only. If the self-serve gate drains the candidate list below 3, emit fewer questions and a fuller Technical Context section — that is the correct outcome.

1. **Validate file references** against actual repo state.
2. **Emit 3–5 clarifying questions, each as a bounded menu** of 2–4 selectable options (plus an explicit "none of these / I'll describe it" escape hatch). No open-ended prose questions — menus force a decision and prevent runaway question trees. Format each as:

   ```
   Q{n}. {short question}
     A) {concrete option with its tradeoff}
     B) {concrete option with its tradeoff}
     C) {concrete option with its tradeoff}
     D) Other — describe
   ```

3. **Forced round-trip.** Do not accept the first spec without at least one grill round. Even if the answers all point the same way, restate the collapsed spec back to the requester and confirm.
4. **Anti-agreement-theater.** If the requester's answer conflicts with what the codebase pattern would dictate, say so explicitly: *"You picked A, but the existing pattern in `file:line` is B. Confirm A is intentional?"* — do not silently go along with the first answer.

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` for the entire refine. No shortcuts: do not skip Phase 0 (optimal-version reframe) for anything larger than a trivial bug, do not accept the first spec without at least one grill round, do not paraphrase a requirement to make it easier to satisfy. Use tools fully: validate every cited file path against actual repo state, fetch Jira via the `atlassian` MCP rather than improvising, prefer `context7` for library questions. Do not compress reasoning to save tokens — the grill is the value producer; collapsing it silently downgrades every downstream stage.
- Prefer concrete acceptance criteria over generic summaries
- If UI work is involved, preserve the existing design system unless the user explicitly asks for a redesign
- If repo health limits confidence, say so in Assumptions
- Keep the refinement actionable enough to hand directly to `/m:plan` or `/m:implement`

## Output

Return:

## Refined Specification

### Goal
### Classification
### Scope
### Out of Scope
### Success Criteria
[Target state expressed as observable conditions, not activities. Read by `/m:iterate` as exit-predicate input. Example: "`go test ./...` exits 0", "user completes checkout without redirect loop", "API returns 401 when token missing". Avoid "improve X" or "make it work".]
### Acceptance Criteria
### Technical Context
### Security or Data Impact
### Test Requirements
### Assumptions
### Recommended Next Command
