# Self-Serve Before Ask (HARD)

Before emitting ANY question to the user — clarifying, bounded-menu, or inline — classify the question first. The user's time is the scarcest resource in the loop, and a question that grep, Read, or an MCP call could have answered is a tax on it. This rule is a positive mandate that sits next to `verification.md` (which forbids asserting without reading) and `rigor.md` (which forbids shortcuts).

Reaching for a question because asking is faster than reading is the failure mode this file exists to prevent.

## Question classification

Every candidate question carries one of three tags:

- **[FACTUAL]** — the answer exists in the repository, the running system, an MCP-reachable external system, or vendor documentation. Examples: struct field names, table or column names, JSON shapes, function signatures, file paths, framework or library defaults, current branch state, the exact test command, environment variables, config values, the contents of a Jira story, current dashboard state, the result of a test or lint run.
- **[USER-INTENT]** — the answer is irreducibly subjective and lives only in the user's head. Examples: scope tradeoffs, business rules with no recorded source, preferences between equally valid options, deadlines, stakeholder context, which of two existing patterns should be canonical going forward, which optional acceptance criterion the user actually cares about.
- **[MIXED]** — the question contains both. Split it: tool-resolve the factual half and ask only the user-intent residue.

A question that cannot be cleanly tagged `[USER-INTENT]` is factual by default. Resolve it via tools.

## Pre-flight check (run before every user-facing question)

1. Can `grep`, `Read`, `Glob`, or `Bash` settle this in under 60 seconds against the current working tree? If yes, settle it. Do not ask.
2. Is there an MCP server that owns this domain (`atlassian` for Jira and Confluence, `context7` for library and framework docs, `claude-in-chrome` or `playwright` for live browser state, `computer-use` for native app state)? If yes, call the MCP. Do not ask.
3. Have I already Read the file the question is about? If yes, the answer is in the file — re-read the relevant region. Do not ask.
4. Am I asking because asking is faster than reading? If yes, read it.

If all four checks pass and the question still has a residue that no tool can resolve, that residue is `[USER-INTENT]` and may be asked.

## Forbidden ask patterns

Replace these with the listed action. Each pattern is a sign that the self-serve pass was skipped.

- "Can you check what fields are on `<struct>`?" → Read the file that defines the struct.
- "What is the signature of `<function>`?" → Grep for the function definition and Read the file.
- "Does `<framework>` support X?" → Call the `context7` MCP and Read the relevant doc.
- "What is in Jira ticket `<KEY>`?" → Call the `atlassian` MCP `getJiraIssue` tool.
- "Is the test passing?" → Run the test command via Bash and quote the exit code.
- "What does `<error>` mean?" → Read the source that emits the error and run the failing command.
- "What's the current branch?" → Run `git status` or `git rev-parse --abbrev-ref HEAD`.
- "Are there any TODOs in the file?" → Grep the file.
- "Is the dev server running?" → Check the process or hit the URL.
- "What's in the config?" → Read the config file.
- "Where is X defined?" → Grep for X.
- "Could you confirm the schema of table `<name>`?" → Read the migration files or the schema introspection output.

## Permitted ask patterns

These are the shapes that survive the pre-flight check.

- "Option A favors X, Option B favors Y. Which tradeoff matches your goal?"
- "Scope question: include legacy callers, or new-API only?"
- "Is there a deadline or stakeholder constraint that isn't in the code?"
- "Pattern P1 and Pattern P2 both exist in the repo. After reading both, P1 looks dominant in newer code (`<file:line>`) and P2 in older code (`<file:line>`). Should new code follow P1?"
- "The acceptance criteria leave A through E open. Which subset is must-have for this delivery?"
- "Two business rules conflict (rule X in `<source>`, rule Y in `<other source>`). Which one wins for this change?"

Each permitted pattern names the tradeoff explicitly and shows that the factual ground has already been covered.

## Output discipline

Every question emitted to the user must be prefixed with `[USER-INTENT]` in chat. The prefix is a forcing function: if the question cannot wear the prefix cleanly, it is factual and the prefix is the signal to go resolve it instead.

Example, refine Phase 2 menu under this rule:

```
Q1. [USER-INTENT] Should new uploads block the request, or run async?
  A) Block — simpler, but the p99 latency rises by ~400 ms based on the
     existing handler in handlers/upload.go:120.
  B) Async — adds a queue, but keeps the handler under the 500 ms SLO
     documented in PROJECT_INDEX.md.
  C) Other — describe.
```

The factual context (latency, SLO, file references) is resolved via Read before the question goes out, not asked back to the user.

## Interaction with refine Phase 2

The 3–5 clarifying-question floor in `m-pipeline/commands/m/refine.md` Phase 2 applies to `[USER-INTENT]` questions only. If the self-serve pass drains the candidate question list below 3, that is the correct outcome — emit fewer questions and a fuller `Technical Context` section. A weak refine that asks the user about facts the codebase already settled is worse than a refine with two strong intent questions and a complete context block.

## Interaction with verification.md

`verification.md` rule "I don't know is a valid answer" still holds. It applies *after* self-serve has been exhausted, not as a way to skip self-serve. Order of operations: self-serve first, then `Needs verification`, then ask, then `I don't know`. Reaching for `I don't know` before self-serve is a shortcut.

## Interaction with rigor.md

`rigor.md` rule 2 ("Use tools to their full extent") is the negative form of this rule. This file is the positive form: not just "do not assert without tools," but "do not ask the user before tools." Together they close the loop on both directions of laziness.

## Conflicts and precedence

- When this rule conflicts with `principles.md` "Surgical Changes," surgical wins — do not expand scope while self-serving.
- When this rule conflicts with `code-quality.md` floors, the floors win — do not skip a function-length split because self-serving covered the fact.
- When this rule conflicts with `caveman-output.md`, caveman wins for chat output formatting only — never for the work, the tool calls, or the verification.

## Enforcement

This file is loaded at every session start through `~/.claude/rules/`. It applies to every `/m:*` skill, every spawned subagent, every standalone Claude Code conversation, and every plan, review, iterate, and implement run regardless of size or "feel." It is a floor, not a preference. The pre-flight check runs before every question to the user, every time.
