---
description: Build or refresh persistent project memory under .m/ (INDEX, TASKS, PROGRESS, GAPS, RESEARCH). Use on first /m:* run in a repo or when index is stale. Foundation for downstream /m:* stages.
argument-hint: [focus-area]
model: claude-opus-4-8
effort: xhigh
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(git:*), Agent
---
# /m:index - Project Indexer

Build or refresh persistent project memory for the current repo.

## Input

Optional focus area: `$ARGUMENTS`

No arguments means full index. A specific focus means scan deeply only in that area while still preserving repo-wide context.

## Source Priority

Build the index from these sources, in this order:

1. Existing `.m/` files
2. `PROJECT_INDEX.md` and `PROJECT_INDEX.json`
3. `AGENTS.md` and `CLAUDE.md`
4. `.planning/`
5. repo manifests, lockfiles, and build configs when present
6. The current codebase structure and key files

## Repo-Specific Rules

- Repos that already maintain `PROJECT_INDEX.*`, `.planning/`, `AGENTS.md`, or `CLAUDE.md`: treat those as authoritative inputs; bootstrap `.m/` from them instead of inventing a parallel model from scratch
- Repos that already have `.m/INDEX.md`: prefer it and only supplement with `PROJECT_INDEX.*` or `.planning/`
- If `.m/` exists but is not ignored by git, add `.m/` to `.gitignore` surgically

## Persistence

The canonical memory files are:

- `.m/INDEX.md`
- `.m/TASKS.md`
- `.m/PROGRESS.md`
- `.m/GAPS.md`
- `.m/RESEARCH.md`

Only create `.m/PLAN.md` if the repo already uses it or the user explicitly asks to persist plans.

## Workflow

1. Detect the repo root and current project identity
2. Read existing memory and project guidance before scanning code
3. Parallelize independent discovery — gather these concurrently where possible:
   - **Group A** (no dependencies): repo manifests, lockfiles, build configs, `.gitignore`, existing `.m/` files
   - **Group B** (no dependencies): `PROJECT_INDEX.*`, `AGENTS.md`, `CLAUDE.md`, `.planning/`
   - **Group C** (after A): stack detection, directory structure scan, first-party vs vendor classification
   - **Group D** (after C): architecture patterns, test commands, maintainability hotspots
   Run groups A+B in parallel, then C, then D. Timeout any single scan at 5 seconds — skip and note what was missed rather than blocking the entire index.
4. From the gathered data, establish:
   - source completeness and repo health
   - first-party vs vendor, generated, bundled, or binary artifacts
   - stack and tooling
   - architecture, runtime control flow, and major flows
   - key directories and modules
   - error handling, logging, testing, auth, API, and permission patterns
   - state ownership, plugin or tool systems, and integration boundaries
   - shared packages and restricted files
   - practical run, test, and verification commands
   - maintainability hotspots such as oversized files, TODO/FIXME clusters, and config/env sprawl
5. Create or refresh `.m/INDEX.md`
6. Initialize missing tracking files with these sections:
   - `PROGRESS.md`: Current Focus, Recently Completed, In Progress, Next Up
   - `TASKS.md`: Active, Completed, Blocked
   - `GAPS.md`: Bugs, Missing Features, Technical Debt, Missing Tests, Security Concerns
   - `RESEARCH.md`: dated research entries
7. Ask at most 3 clarification questions, only for real ambiguity
8. Report what changed and which files were touched

## Output Format

Produce or refresh an index with:

- project name
- last indexed date
- repo health and verification caveats
- stack table
- architecture summary
- runtime and control-flow map
- key directories
- established patterns
- shared packages
- key file reference
- conventions
- maintainability hotspots
- restricted files
- test and run commands

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` for the entire index run. No shortcuts: do not bootstrap `.m/INDEX.md` from a guess of repo structure when the actual files exist; do not skip Group A/B/C/D scans because the repo "looks small". Use tools fully: Read manifests, lockfiles, and existing `.m/` and `PROJECT_INDEX.*` content before scanning code; spawn an `Explore` subagent (`model: haiku` — read-only discovery, no reasoning downgrade where it matters) if discovery would span more than three queries. Do not compress reasoning to save tokens — a thin index breaks every downstream `/m:*` stage that depends on it.
- Preserve user-authored content and history
- Update surgically instead of rewriting files wholesale
- Call out conflicting or stale source documents when you find them
- If the repo is a source-map dump, extracted bundle, or incomplete snapshot, say that clearly in the index
- If the user asked for a focused index, make that area deeper but still keep the project-level overview usable
