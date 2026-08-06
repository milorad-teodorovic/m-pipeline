---
description: Analyze stored /m:* learning signals and generate ADAPTATIONS.md when evidence is sufficient. Use to inspect or refresh per-skill behavioral adaptations derived from user feedback.
argument-hint: [dry-run|explain <adaptation>|skill-name]
model: claude-sonnet-5
effort: medium
allowed-tools: Read, Grep, Glob, Edit, Write
---
# /m:learn - Adaptive Learning Analysis

Analyze stored learning signals for `/m:*` usage and generate `ADAPTATIONS.md` when there is enough evidence.

## Input

Optional subcommand: `$ARGUMENTS`

- no args: full analysis and write `ADAPTATIONS.md`
- `dry-run`: analyze and report changes without writing
- `explain <adaptation>`: trace the evidence behind one adaptation
- `<skill>`: analyze one skill more closely

## Workflow

1. Read all JSONL files from `~/.claude/m-learning/signals/`
2. If there is no usable data, say so and stop
3. Detect patterns across:
   - corrections
   - explicit preferences
   - outcomes
   - code style
   - project-specific behavior
   - pipeline events (see Event Types below)
4. Score confidence:
   - HIGH: strong repeated evidence
   - MEDIUM: enough evidence to be useful
   - LOW: track only, do not apply
5. For normal mode, write `~/.claude/m-learning/ADAPTATIONS.md`
6. For `dry-run`, show what would change without writing
7. For `explain`, show the evidence chain behind the selected adaptation

## Output Structure

`ADAPTATIONS.md` should stay concise and include only HIGH or MEDIUM confidence items:

- pipeline defaults
- skill-specific overrides
- code style preferences
- project-specific patterns

Keep evidence summaries short and remove contradicted adaptations.

## Event Types

Step 1 already reads every `.jsonl` file in the signals directory, so these
files need no separate read path. They differ from the older signals in one
way that matters for scoring: the pipeline records them as events happen,
rather than as a self-assessment written at the end of a run.

| Event `type` | Written by | Evidence about |
|---|---|---|
| `gate_denial` | `enforce-develop-phase.py` (`gate-denials.jsonl`) | pipeline discipline, per project and phase |
| `phase_reentry` | `/m:develop` | pipeline discipline, per project and phase |
| `iterate_loop` | `/m:iterate` | plan quality |
| `review_precision` | `/m:review`, `/m:review-fanout` | review quality |
| `engine_disagreement` | `/m:review`, `/m:review-fanout` | second-engine value |

### Behavioral mapping

Discipline events auto-apply at HIGH like any other adaptation, but only
through the preference categories that already exist in `/m:feedback`. An
event type maps to an adaptation or it does not apply at all:

| Observed pattern | Adaptation |
|---|---|
| Repeated `gate_denial` or `phase_reentry` on `implement` in one project | `plan_depth=deep` for that project |
| Repeated `gate_denial` or `phase_reentry` on `review` in one project | `review_strictness=high` for that project |
| Repeated `iterate_loop` with `loops` above 1 and `failed_clause` of `tests_green` | `test_approach=tests-first` for that project |
| Repeated `review_precision` where `findings_survived` is below half of `findings_total` | `review_strictness=high` for that project |
| Repeated `engine_disagreement` with `stricter_applied` true in one project | `review_strictness=high` for that project |

Every row names both a category and the value to write. A row that named
only a category would leave the model to invent the value, which is the
same defect as having no mapping at all.

The `review_precision` row reads a low survival ratio as over-flagging: the
reviewer produced findings that did not hold up under verification, so the
correction is a higher evidence bar before a finding is emitted, which is
what `review_strictness=high` means here. It is not a signal to review less.

An `engine_disagreement` where `stricter_applied` is false stays diagnostic
and maps to nothing. Repeated false values are evidence the second engine is
not changing outcomes for that project, which is a decision for the user to
make about configuration, not an adaptation to apply automatically.

An event that matches no row above is recorded at LOW and never applied,
however strong the evidence. This is deliberate: without a mapping onto an
existing preference category, applying an adaptation would mean inventing a
behavior change from a diagnostic signal.

Adaptations are written only to `~/.claude/m-learning/ADAPTATIONS.md`. No
adaptation may edit a command file or a skill file. Those files are scored
against a committed evaluation baseline, and an automated edit would
invalidate that baseline with no diff for the user to review.

## Application Rules

0. Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` (read in full before proceeding). Read every signal file before scoring; never promote a LOW signal to fill a slot.
1. HIGH: safe to apply silently
2. MEDIUM: apply, but mention when relevant
3. LOW: track only
4. Current session instructions always override learned behavior
5. If `ADAPTATIONS.md` does not exist, other `/m:*` commands should proceed without it
