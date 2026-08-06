---
description: Post-implementation verification loop — run tests, fix issues, re-check until exit predicate is satisfied or the 3-loop safety cap is reached. Use after /m:implement or when verifying recent code changes.
argument-hint: [scope-or-check]
model: claude-sonnet-5
effort: high
disable-model-invocation: false
---
# /m:iterate - Verification Workflow

Run post-implementation verification, fix current issues, and repeat until the result is clean or blocked.

## Input

Context to verify: `$ARGUMENTS`

If no explicit target is given, verify the most recent implementation in context.

## Context Sources

- `.m/INDEX.md` for project patterns and test commands
- `.m/PROGRESS.md`
- `.m/GAPS.md`
- repo manifests and existing test scripts
- `~/.claude/m-learning/ADAPTATIONS.md` (if present) — apply the HIGH and MEDIUM `iterate` adaptations and the `test_approach` preference recorded there; proceed normally if it does not exist. Current-session instructions always override a learned adaptation.

## Workflow

### Phase Marker Protocol

This skill participates in the `/m:develop` phase gate. Follow this
protocol on every invocation, including standalone runs:

1. On entry, before running any tests or checks: run
   `mkdir -p .m && touch .m/phase-iterate-started` via Bash.
2. On successful completion (exit predicate satisfied and `PASSED`
   emitted): run `touch .m/phase-iterate-done`.
3. On loop-count exit (`BLOCKED`) or mid-run abort: leave `-started` in
   place and do NOT write `-done`.

If `.m/DEVELOP_ACTIVE` is present and its `current_phase:` line does not
read `iterate`, stop and tell the user — the pipeline is out of sync.

1. Determine the affected surface
2. Run the most relevant tests and checks first
3. Audit error handling, security basics, and obvious quality issues
4. Fix CURRENT issues when the fix is in scope
5. Log PRE-EXISTING issues to `.m/GAPS.md`
6. Re-run the checks — always re-run, never assume a fix worked
7. Evaluate the **exit predicate** (below) after each loop
8. Stop when the exit predicate is satisfied, or after 3 fix loops as a hard safety cap

For large verification passes, you may use focused subagents if the runtime supports them cleanly. Do not depend on external team-orchestration primitives.

## Exit Predicate

The loop exits when **all** of these are true:

1. **Tests green.** The relevant test suite runs and exits 0. If there are no tests, fall back to the best available verification (type-check, lint, build) and say so explicitly in the Loop Summary.
2. **Zero critical review findings.** If `/m:review` or `/m:review-fanout` produced findings against this change, every `critical`-severity finding is either fixed or explicitly waived by the user. `high` severity findings should also be resolved unless the user accepts them in writing.
3. **`.m/PROGRESS.md` updated.** The new state is written with: what was fixed, what was waived, and what (if anything) is still open.
4. **PRD Success Criteria satisfied (if present).** If a `.m/PRD-*.md` exists for this change and contains a `## 8. Success Criteria` section, every listed condition must be verifiably true. Each condition is a target-state predicate (e.g. exit code, observable user flow, response code, latency bound), not an activity description. If a condition cannot be verified in the current environment, list it under Remaining Issues with the exact reason and exit `BLOCKED`.

Token budget, loop count, and "it looks fine" are **not** valid exit conditions. If the predicate isn't satisfied and the 3-loop safety cap is reached, exit with `BLOCKED` and list exactly which predicate clause failed.

The hard cap exists to prevent runaway burn, not as a substitute for the predicate. A PASSED verdict requires the predicate, not just the cap.

## Context Accumulation

Each fix loop carries forward what was learned in prior loops:

- **Loop N** produces: failures found, fixes applied, new test output
- **Loop N+1** receives all of the above as input context
- Never re-investigate an issue that was already fixed and verified clean in a prior loop
- If a fix in loop N introduces a new failure in loop N+1, link them: "Fix for X in loop 1 caused Y in loop 2"
- After each loop, emit a brief progress line before continuing:
  `Loop {n}/3: {fixed_count} fixed, {remaining_count} remaining, {new_count} new`

This prevents circular debugging and makes the iteration log self-documenting.

## Progressive Output

Report progress at each milestone, not just at the end:

1. After initial check run: `Checks complete: {pass_count} passed, {fail_count} failed`
2. After each fix loop: `Loop {n}/3: {fixed_count} fixed, {remaining_count} remaining`
3. After final verdict: full report below

This lets the user see movement during long verification passes.

## Output

Finish with:

## Iteration Report

### Status
### Checks Run
### Loop Summary
Brief log of what each loop found and fixed (1-2 lines per loop).
### Issues Found and Fixed
### Remaining Issues
### Pre-Existing Gaps Logged
### Exit Predicate Status
- Tests green: yes / no — {command and exit code}
- Zero critical review findings: yes / no / n/a — {count and severity}
- PROGRESS.md updated: yes / no
- PRD Success Criteria satisfied: yes / no / n/a — {which criteria, or n/a if no `.m/PRD-*.md`}
### Verdict

Use `PASSED` or `BLOCKED` for the final verdict.

- `PASSED` requires all four exit-predicate clauses = yes (or `n/a` for clause 2 when no review was run, and `n/a` for clause 4 when no `.m/PRD-*.md` exists for this change).
- `BLOCKED` means at least one clause failed. Name which one(s) and why.

## Learning Signal

After the verdict is determined, append one JSON line to
`~/.claude/m-learning/signals/pipeline-events.jsonl` with the file tools —
never a Bash `echo`, matching the append convention in `/m:develop`:

```json
{"timestamp":"<ISO-8601 UTC>","type":"iterate_loop","project":"<repo dir basename>","loops":N,"verdict":"PASSED|BLOCKED","failed_clause":null,"per_loop":[{"loop":1,"fixed":0,"remaining":0,"new":0}]}
```

The `per_loop` array reuses the counts already emitted as the
`Loop {n}/3: {fixed_count} fixed, {remaining_count} remaining, {new_count} new`
progress line, one entry per loop actually run.

`failed_clause` names the first unsatisfied clause of the exit predicate —
one of `tests_green`, `zero_critical_findings`, `progress_updated`, or
`prd_criteria` — and is `null` on a `PASSED` verdict.

Signal writing is non-blocking. A failed append never changes the verdict,
never fails the run, and is reported as one line in chat.

The signals file is global and concurrent `/m:*` sessions append to it at
once. Build the whole record first and append it as one complete line in a
single write; never rewrite or reflow lines that are already in the file.

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (read in full before proceeding). `PASSED` requires the exit predicate with quoted commands and exit codes — never the loop cap, never an un-re-run fix.
- Re-run the relevant checks after every fix loop
- An intermittently failing test is characterized before it is judged: re-run it several times, report the observed pass/fail pattern as an open item, and never let a single green re-run count as conclusive
- Keep CURRENT issues separate from PRE-EXISTING gaps
- If there are no tests, say that explicitly and fall back to the best available verification
- If the repo is incomplete or missing build metadata, say exactly what could not be verified and why
- Flag plan defects directly. If the plan itself looks wrong, say so and escalate — improvising past a bad plan costs more than the pause
- When output drifts into self-correction or repeated apology, emit the next concrete action and resume. Apology spirals reinforce cautious hedging across the remaining loops, and the next action is the signal that ends them
- After a loop finishes clean, emit a positive checkpoint (`Loop N clean — continuing to {next check}`) before the next loop so the session frame reflects progress, not only failures
