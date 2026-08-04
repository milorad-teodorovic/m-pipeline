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

## Application Rules

0. Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` (read in full before proceeding). Read every signal file before scoring; never promote a LOW signal to fill a slot.
1. HIGH: safe to apply silently
2. MEDIUM: apply, but mention when relevant
3. LOW: track only
4. Current session instructions always override learned behavior
5. If `ADAPTATIONS.md` does not exist, other `/m:*` commands should proceed without it
