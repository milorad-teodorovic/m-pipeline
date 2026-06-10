# Verification Rules

These rules apply to every `/m:*` command and every agent. They codify the verification discipline that the go-security-reviewer agent already enforces, promoted to a global floor so every stage of the pipeline is held to the same bar.

## Hard rules

1. **No persona priming.** Never open a prompt with "you are the world's best programmer" or equivalent. It measurably degrades planning and review output. Describe the *role and responsibilities* procedurally; do not motivate.
2. **Read before you cite.** Any file, struct, function, flag, table, column, or API signature you reference in output must come from a file you read *in this session*. Recalled knowledge is not a citation.
3. **Verify exact line numbers.** Diffs show hunk-relative positions. Cited line numbers must match an actual Read() of the current file (±0). Re-read and fix mismatches before emitting findings.
4. **Quote verbatim.** Code snippets in findings must be copied from Read/Grep output, never reconstructed from memory.
5. **No phantom conclusions.** Recognizing a function name (`Raw`, `Exec`, `dangerouslySetInnerHTML`, etc.) is not a finding. Trace where the arguments come from. If the inputs are hardcoded, server-controlled, or already sanitized, there is no finding.

## Self-challenge gate

Before emitting ANY finding, recommendation, or assertion that depends on code the user will act on:

1. **State the strongest counter-argument.** What is the strongest reason this might NOT be a real issue, or this recommendation might be wrong?
2. **Trace the counter-argument** through actual code. Read the upstream validators, middleware, framework defaults, or call sites that could disprove your claim.
3. **Decide:**
   - Counter-argument holds → **drop the finding entirely.** Do not hedge, do not mention it.
   - Counter-argument fails → **include the finding** and document (a) the counter-argument you considered, (b) the specific code that disproves it.

Prefer zero findings over weak findings. Prefer "I don't know" over a confident guess.

## "I don't know" is a valid answer

If you cannot verify a claim from actual code in this session, you have three options — use one of them, never a fourth:

- **Move the item to "Needs verification"** with an explicit list of what is missing.
- **Ask the user to point you at the source** (file path, branch, PR, doc).
- **Say "I don't know" directly** and stop.

Boundaries:
- Treat unread struct fields, table names, API signatures, and config values as unknown until a Read confirms them. State the gap rather than filling it.
- Ground "probably" and "usually" intuitions in a Read of the actual code before asserting them.
- Treat the codebase as the source of truth over training-data defaults.

## Practice habits

These habits prevent the silent quality regressions the verification floor is designed to catch:

- **Read what you cite.** If you need to know what a function does, read it. Turn count is not a quality metric.
- **Run the counter-argument step every time.** It feels like padding, and it is the single most effective hallucination filter.
- **Give each reviewer lens its own scoped context.** Security, architecture, tests, and performance each run with independent inputs. Parallel is not serial-with-fewer-turns.
- **Push back on the first plan.** The first plan is almost always a lowest-common-denominator version. Challenge scope, edge cases, or the "optimal version ignoring time/labor" framing at least once.
- **Iterate until the exit predicate is satisfied.** Exit predicate is tests green + review score met + zero open architectural notes. Token budget is not the exit predicate.
- **Log every `// TODO`, `// for now`, and silent `error swallowing` in Deviations.** Surface them in the summary so the user sees what is still open.

## Verification budget

Spend verification effort proportionate to blast radius:

| Change scope | Verification floor |
|---|---|
| Read-only analysis | Self-challenge each claim once |
| Local file edits, tests stay green | Self-challenge + Read-back of every edited region |
| Migrations, auth, money, or cross-service state | Self-challenge + Read-back + explicit tenant/rollback trace + second-model review |
| Public API / contract changes | Everything above + explicit consumer impact list |

## Handoff discipline

When one `/m:*` stage hands work to the next:

- **Plan → Implement:** the plan must reference concrete file paths that were Read in the planning session. Abstract bullets without paths are not a plan.
- **Implement → Review:** the implementation must list exactly which files were touched and which tests were run. "Ran tests" without a command and exit code is not a test run.
- **Review → Iterate:** every finding must carry `file:line` evidence or be discarded before handoff. Iterate does not re-open findings that failed verification upstream.
- **Iterate → done:** the exit predicate must be stated explicitly in the verdict, not implied.

## Enforcement

These rules define the verification floor for every `/m:*` stage. When you notice a step you skimmed, pause and complete it before moving on — the corner you re-do costs less than the rework that an unverified claim triggers downstream.
