---
description: Show project status from .m/ — current focus, gaps, tasks, worktrees. Subcommands update task tracking. Use for "where are we", "what's left", "log a bug", "track progress".
argument-hint: [gaps|task ...|done ...|progress ...|bug ...|debt ...|worktrees|cleanup]
model: claude-sonnet-5
effort: medium
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(git:*)
---
# /m:status - Project Status & Gap Tracker

Show current project state, detect gaps, and update task tracking.

## Input

Interpret `$ARGUMENTS` as one of:

- empty: full dashboard
- `gaps`
- `task <description>`
- `done <task id or description>`
- `progress <update>`
- `bug <description>`
- `debt <description>`
- `worktrees`
- `cleanup`

## Memory Sources

Prefer `.m/` if present. If `.m/` is missing, use `PROJECT_INDEX.*`, `.planning/`, `AGENTS.md`, and `CLAUDE.md` for a read-only dashboard. If the user asks for tracked state updates and `.m/` is missing, bootstrap minimal `.m/` state first or tell the user that `/m:index` is required.

## Default Dashboard

Present:

- project identity
- index freshness
- repo health and verification caveats
- current focus
- active and completed tasks
- gap counts by category
- recent analyses from `.m/analyses/*`
- active worktrees, if any

Keep the output compact and actionable.

**Reading the state correctly (avoid these misreads):**

- Determine index freshness by actually Reading `.m/INDEX.md` and checking its contents and modification recency. Do not report `INDEX.md` as missing without confirming via `ls`/Read — it is frequently the newest `.m/` file.
- Treat `.m/phase-*-started` / `.m/phase-*-done` markers as a **live** pipeline only when `.m/DEVELOP_ACTIVE` is also present. Without `DEVELOP_ACTIVE`, leftover `phase-*` markers are historical — from a finished or aborted run, or archived — so report them as past activity, never as an in-progress pipeline.

## `gaps`

Run an active gap scan using the current repo patterns:

1. Read `.m/INDEX.md` if it exists
2. Scan for TODO, FIXME, HACK, empty bodies, stub returns, commented-out code, and obvious pattern violations
3. Scan for maintainability smells such as oversized files, god objects, module-global mutable state, singleton runtime objects, sync filesystem usage in interactive paths, and excessive `process.env` branching
4. Look for missing tests around changed or critical code
5. Look for auth, validation, rate-limit, and ownership gaps on sensitive paths
6. For workflow or command packs, check that documented commands actually exist and that cross-references are valid
7. If the repo appears incomplete, extracted, or vendor-heavy, log that as a repo-health gap when it blocks reliable verification
8. Update `.m/GAPS.md` with categorized findings and timestamps
9. Separate newly discovered issues from already logged ones

## `task`, `done`, `progress`, `bug`, `debt`

Update `.m/TASKS.md`, `.m/PROGRESS.md`, and `.m/GAPS.md` surgically:

- `task`: add a new active task with a stable identifier
- `done`: move the task to completed and reflect it in progress
- `progress`: update Current Focus or In Progress
- `bug`: add a high-severity bug entry
- `debt`: add a medium-severity technical debt entry

## `worktrees`

List active worktrees by combining:

- `git worktree list`
- repo-local worktree folders such as `.claude/worktrees/` when present

For each worktree, report:

- path
- branch
- current focus from `.m/PROGRESS.md` if present
- task counts from `.m/TASKS.md` if present
- obvious state such as clean, dirty, or stale

Do not assume worktrees are siblings of the main repo. Respect repo-specific layouts (e.g. a `<repo>/.claude/worktrees/` directory).

## `cleanup`

Never remove worktrees automatically.

1. Detect candidates
2. Show why each is a candidate
3. Ask for confirmation before any removal
4. Do not remove branches or unmerged worktrees unless the user explicitly confirms

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md` (loaded at session start). Use file-system tools, not Bash, when updating `.m/` tracking files.
- Preserve existing task and gap history
- Prefer updating existing entries over duplicating them
- If repo state is inconsistent, call it out instead of guessing
