<div align="center">

# `/m:*` — the m-pipeline

### A spec-driven software-delivery pipeline for [Claude Code](https://docs.claude.com/en/docs/claude-code)

*Requests are grilled into specifications, plans are challenged until zero gaps remain, and code ships only after gated review and a verification loop.*

![Claude Code](https://img.shields.io/badge/Claude%20Code-plugin-d97757)
![Commands](https://img.shields.io/badge/commands-14-3b6ea5)
![Skills](https://img.shields.io/badge/expert%20skills-5-3b6ea5)
![Review](https://img.shields.io/badge/review-Claude%20%2B%20Codex-4c9a6b)
![Gates](https://img.shields.io/badge/phases-hook--enforced-c0563a)

</div>

---

## The idea in one line

> **The gates are hooks, not honor rules** — no phase starts until the previous one *proves* it finished.

`/m:develop <request>` drives a change through five gated phases, each running as a discrete Skill invocation behind a marker-file gate. A `PreToolUse` hook denies any code mutation outside `.m/` until the active phase has been entered through its skill — so the model can't skip refine, can't implement past an unapproved plan, and can't call a run "done" without the verification loop.

The phases run in order, and each must prove it finished before the next can start:

1. **refine** — grill the raw request into an execution-ready spec.
2. **plan** — build the implementation plan, dual-engine sanity pass (Claude + Codex).
3. **implement** — write code to the approved plan only, following repo patterns.
4. **review** — evidence-backed review, sequential or parallel blind-lens fan-out.
5. **iterate** — test-and-fix loop until the exit predicate holds.

Project memory in `.m/` is read and written by every stage; support commands (index, status, analyze, learn) build it, and expert modes (go, react, biz, cr, security) clip onto implement.

> See the full architecture diagram and design notes at **[milorad.io](https://milorad.io)**.

---

## Install

This repository **is** a Claude Code plugin marketplace. In Claude Code:

```text
/plugin marketplace add milorad-teodorovic/m-pipeline
/plugin install m@m-pipeline
/reload-plugins
```

Commands, the five expert skills, and the phase-enforcement hook are available immediately. Run `/plugin` → **Installed** to confirm, and `/hooks` to see the phase hook listed.

> No host setup required — the engineering rules the commands depend on ship in `rules/` and resolve via `${CLAUDE_PLUGIN_ROOT}`.

---

## Quick start

**Full pipeline, end to end** (the common case):

```text
/m:develop add per-IP rate limiting to the public API
```

Claude refines the request into a spec, plans it (cross-checked by Codex if available), implements against the plan, reviews the diff, and runs a verify-fix loop — pausing for your input at each gate.

**Or drive the phases yourself**, one command at a time:

```text
/m:refine add per-IP rate limiting to the public API
/m:plan
/m:implement
/m:review
/m:iterate
```

**Or reach for a single tool:**

```text
/m:index          # first run in a repo — build the .m/ memory
/m:status         # where are we? what's left?
/m:cr             # security-review the current diff, evidence-only
/m:analyze the auth package's session handling
```

> **First time in a repo?** Run `/m:index` once. It scans the stack, patterns, and hotspots into `.m/`, which every later stage reads.

---

## The five-phase pipeline

Run together by `/m:develop`, or individually. Each is a real Skill with its own model tier.

| # | Command | What it does | Model |
|:-:|---------|--------------|:-----:|
| ① | `/m:refine` | Grills a raw request into an execution-ready spec — optimal-version reframe, then bounded-menu questions until ambiguity is gone. | Opus |
| ② | `/m:plan` | Builds the implementation plan and challenges it with a dual-engine (Claude + Codex) sanity pass — grilled until zero gaps remain. | Opus |
| ③ | `/m:implement` | Writes code **to the approved plan only**, following repo patterns. Plan defects escalate back to plan rather than being improvised past. | Opus |
| ④ | `/m:review` · `/m:review-fanout` | Evidence-backed review. Sequential for small diffs; parallel blind-lens fan-out (security, architecture, tests, performance, migrations, observability, api-contracts, compliance) + judge for large ones. | Opus |
| ⑤ | `/m:iterate` | Test-and-fix loop until the exit predicate holds (tests green · zero critical findings · progress logged). The 3-loop cap is `BLOCKED`, never `PASSED`. | Sonnet |

---

## All commands

**Pipeline** — the spine above, plus the orchestrator:

| Command | Purpose | Model |
|---------|---------|:-----:|
| `/m:develop` | Run all five phases end-to-end with hard phase gates and dual-engine review. | Opus |
| `/m:refine` · `/m:plan` · `/m:implement` · `/m:review` · `/m:review-fanout` · `/m:iterate` | The phases, standalone (see table above). | Opus / Sonnet |

**Support** — build, inspect, and learn from project memory:

| Command | Purpose | Model |
|---------|---------|:-----:|
| `/m:index` | Build or refresh persistent `.m/` project memory (stack, patterns, hotspots). | Opus |
| `/m:status` | "Where are we" dashboard — focus, gaps, tasks, worktrees. Logs bugs and progress. | Sonnet |
| `/m:research` | Isolated worktree research for unknowns before planning — advisory only. | Opus |
| `/m:analyze` | Deep analysis of code/docs/systems, with optional diagrams and grading. | Opus |
| `/m:feedback` | Store explicit workflow preferences (filesystem only, no inference). | Haiku |
| `/m:learn` | Turn stored feedback signals into per-skill behavioral adaptations. | Sonnet |
| `/m:help` | Print the workflow reference — order, purposes, side-effect tiers. | Haiku |

---

## Expert modes

Five specialist skills. Three **auto-activate** when matching files are edited; two are **invoked by hand**.

| Skill | Focus | Activation |
|-------|-------|------------|
| `m:go` | Senior Go engineering & review | Auto · `**/*.go`, `go.mod`, `go.work` |
| `m:react` | Senior React + TailwindCSS | Auto · `**/*.tsx`, `**/*.jsx`, `tailwind.config.*` |
| `m:biz` | Business-logic & domain mapping | Auto · `.business/**`, `**/BUSINESS.md` |
| `/m:cr` | Security review of changed code — evidence-only, read-only | Manual |
| `/m:security` | Standing-codebase OWASP/CWE audit + threat model | Manual |

---

## How it works

**Phase enforcement.** `/m:develop` writes marker files (`.m/DEVELOP_ACTIVE`, `.m/phase-<name>-started`/`-done`). The `enforce-develop-phase.py` `PreToolUse` hook denies `Edit`/`Write`/`MultiEdit` outside `.m/` until the active phase has been entered through its skill. Writes inside `.m/` are always allowed. This is what makes the gates real rather than advisory.

**Dual-engine (opt-in).** When you enable it per repo via `.m/pipeline.yml` `codex.enabled: true` (off by default), Codex runs as a second engine across the pipeline: `/m:plan` gets two blocking passes (Pass-1 architecture sanity, Pass-2 final-plan verdict), `/m:research` runs a parallel Codex researcher reconciled with Claude's, and `/m:review` / `/m:review-fanout` run Codex on every review with verdicts side-by-side. The stricter verdict always wins. Each run is **token-metered** against a budget (`token_budget`, default 200k) with a graceful fallback to Claude-only when the budget is reached, and an optional **fast mode**. Defaults and toggles: `references/pipeline-context.md`; full protocol: `references/codex-protocol.md`.

**Fan-out review.** `/m:review-fanout` spawns blind specialist subagents in parallel — each sees only its lens — then a judge pass reconciles and dedupes at `file:line`. Lens prompts: `references/lens-templates.md`.

**Project memory (`.m/`).** Structured per-repo state every stage reads and writes:

| File | Holds |
|------|-------|
| `INDEX.md` | Repo identity, stack, patterns, hotspots |
| `TASKS.md` · `PROGRESS.md` · `GAPS.md` | Work tracking |
| `RESEARCH.md` | Appended research findings |

**Side-effect tiers.** Every change is classified before implement runs:

| Tier | Behavior |
|------|----------|
| `read-only` | Full pipeline, no confirmation |
| `write-local` | Proceed after the initial scope confirmation |
| `write-external` | Pause before implement and any destructive step — confirm explicitly |

---

## Optional dependencies

The pipeline **degrades gracefully** when these are absent:

- **Codex CLI** (`codex` ≥ 0.123.0 on `PATH`) — enables the dual-engine passes across `/m:plan`, `/m:research`, and review when you opt in via `.m/pipeline.yml` `codex.enabled: true` (off by default). Token-metered per run with a budget + graceful fallback; optional fast mode (`codex.fast_mode`). Without it — or when left disabled — every stage runs Claude-only.
- **`atlassian` MCP** — enables Jira enrichment when a request matches a Jira key and a per-project `.m/jira.yml` exists. Set up with:
  ```text
  claude mcp add --transport http --scope user atlassian https://mcp.atlassian.com/v1/mcp
  ```
  Without it, Jira steps are skipped.

---

## Layout

```text
m-pipeline/
├── .claude-plugin/
│   ├── plugin.json
│   └── marketplace.json
├── commands/          # 14 slash commands (plugin name `m` → invoked as /m:refine, /m:plan, …)
├── skills/            # 5 expert-mode skills (/m:go, /m:react, /m:biz, /m:cr, /m:security)
├── references/        # codex-protocol · jira-context · lens-templates · pipeline-context · checklists
├── rules/             # rigor · self-serve · verification  (referenced via ${CLAUDE_PLUGIN_ROOT})
├── hooks/
│   ├── hooks.json
│   └── enforce-develop-phase.py
└── README.md
```

## License

MIT, see [LICENSE](LICENSE).

---

<div align="center">
<sub>Built for Claude Code · spec-driven delivery with real gates.</sub>
</div>
