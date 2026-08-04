---
description: Focused, isolated research using internet sources, official docs, and local code. Spawns in a worktree to prevent context pollution. Use when planning encounters unknowns or user asks "research", "investigate", "compare options".
argument-hint: [topic-or-question]
model: claude-opus-5
effort: xhigh
allowed-tools: Read, Grep, Glob, WebSearch, WebFetch, Agent, Bash(codex exec:*), Bash(codex --version), Bash(kimi -p:*), Bash(kimi --version), Bash(git:*), Bash(rm:*), Bash(mkdir:*), Bash(python3:*)
---
# /m:research - Focused Research Workflow

Research unfamiliar, risky, or multi-option work before planning or implementation.

## Input

Research topic or question: `$ARGUMENTS`

## Isolation

**Research MUST run in an isolated context to prevent pollution from existing plan or refine context.**

When invoked (by `/m:plan`, `/m:develop`, or the user directly), research spawns via:

```
Agent(isolation: "worktree")
```

The agent prompt must contain ONLY:
- The research question or topic
- Relevant file paths for local code inspection

The agent prompt must EXPLICITLY EXCLUDE:
- The refined spec from `/m:refine`
- Any plan-so-far from `/m:plan`
- Prior conversation context, decisions, or assumptions

**Fallback:** If worktree creation fails (dirty git state, conflicts), fall back to `Agent()` without worktree. Log the degradation: `[research] Worktree unavailable — running in subagent without filesystem isolation.`

## Advisory Status

**Research findings are advisory.** The user's plan wins unless the user explicitly decides to incorporate research findings. The plan stage presents research findings to the user; the user decides what (if anything) to adopt.

Research does not auto-revise plans. Research does not override user-confirmed decisions.

## Context Sources

The research agent reads these when relevant (within its isolated context):

- `.m/INDEX.md`
- `.m/GAPS.md`
- `PROJECT_INDEX.md`
- repo-local guidance such as `AGENTS.md` and `CLAUDE.md`
- the local code, docs, or config related to the topic

Use web research when the answer depends on external systems, changing facts, or official documentation.

## Dual-Engine Research (second engine)

When `second_engine.provider` is `codex` or `kimi` (see `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`), research runs as a **parallel dual-engine** pass: the second engine researches the same questions independently while Claude's research agent runs, then Claude reconciles both into one finding set. Follow the active provider's protocol Section 11 (`${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` or `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md`) — config resolution (Section 1), the Operating-Rules Preamble (Section 4), Secret Redaction (Section 5), the Metered Invocation (Section 6), Token Metering (Section 7), and the reconciliation rules (`[CORROBORATED]` where both agree; side-by-side `[claude]` vs `[codex]`/`[kimi]` where they disagree; merge unique findings preserving citations). When the provider is `none`, the CLI is unavailable, or the token budget is reached, run Claude-only research exactly as before.

## Workflow

1. Convert the request into 2-5 focused research questions
2. Resolve `second_engine` and, when a provider is selected, kick off its research pass (protocol Section 11) in parallel with the steps below
3. Inspect local code and repo patterns first
4. Use official docs and primary sources for external facts
5. Separate:
   - stable local observations
   - external facts
   - recommendations or opinions
6. Compare options, tradeoffs, risks, and unknowns
7. Reconcile Claude and second-engine findings per protocol Section 11 (when dual-engine ran)
8. Return findings as text to the parent session, then run Handoff Cleanup (protocol Section 13)

The research agent does NOT write to `.m/RESEARCH.md`. The parent session writes findings to `.m/RESEARCH.md` if and when the user confirms they should be persisted.

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (loaded at session start). Primary sources over recall; the parent session relies on the full finding set, not a pre-summarised verdict.
- Include concrete dates, versions, or protocol names when they matter
- Make uncertainty explicit
- If the repo is incomplete or extracted, distinguish what was observed locally from what had to be inferred
- Keep the final recommendation decisive unless the unknowns are genuinely blocking

## Output

Return:

## Research Complete

### Research Questions
### Findings
### Recommended Approach
### Risks
### Remaining Unknowns
