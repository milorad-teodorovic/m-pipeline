---
description: Store explicit /m:* workflow preferences (approve, reject, prefer, style). Use only when the user explicitly wants persistent learning for /m:* behavior.
argument-hint: [approve|reject|prefer|style|show|stats|reset] ...
model: claude-haiku-4-5
effort: low
allowed-tools: Read, Edit, Write
disable-model-invocation: true
---
# /m:feedback - Adaptive Learning Feedback

Use this command only when the user explicitly wants persistent Claude-side learning for the `/m:*` workflow.

## Input

Subcommand and arguments: `$ARGUMENTS`

Store signal files under `~/.claude/m-learning/signals/` — this is the directory `/m:learn` reads. `ADAPTATIONS.md` lives one level up at `~/.claude/m-learning/ADAPTATIONS.md`.

## Subcommands

### `approve <skill> <detail>`

Write an approval signal to `signals/feedback.jsonl`.

### `reject <skill> <detail>`

Write a rejection signal to `signals/feedback.jsonl`.

### `prefer <category> <value>`

Supported categories:

- `plan_depth`
- `question_count`
- `iteration_tolerance`
- `research_depth`
- `review_strictness`
- `agent_count`
- `commit_style`
- `test_approach`

Write a preference signal to `signals/feedback.jsonl`.

### `style <language> <detail>`

Write a language-specific style preference signal to `signals/feedback.jsonl`.

### `show`

Read and display `~/.claude/m-learning/ADAPTATIONS.md`, or say it does not exist yet.

### `stats`

Show signal counts by file and whether `ADAPTATIONS.md` exists.

### `reset`

Ask for confirmation first. Only clear learning data if the user confirms. Clear by overwriting each signal file under `~/.claude/m-learning/signals/` to empty with the Write tool — do **not** delete the files (`Write` is the only mutation tool granted here, and an emptied file is an equivalent reset). Report which files were cleared.

### No arguments

Show:

```text
/m:feedback <subcommand>
  approve <skill> <detail>
  reject <skill> <detail>
  prefer <category> <value>
  style <language> <detail>
  show
  stats
  reset
```

## Valid Skills

- `index`
- `status`
- `refine`
- `research`
- `plan`
- `implement`
- `review`
- `iterate`
- `develop`
- `analyze`
- `cr`

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` (read in full before proceeding). Write each signal exactly once with the file-system tools, never Bash echo.
- Store signals as timestamped JSONL entries
- Keep this layer opt-in; do not suggest it unless the user wants persistence
