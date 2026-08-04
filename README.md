<div align="center">

# `/m:*` — the m-pipeline

### A spec-driven software-delivery pipeline for [Claude Code](https://docs.claude.com/en/docs/claude-code)

*Requests are grilled into specifications, plans are challenged until zero gaps remain, and code ships only after gated review and a verification loop.*

![Claude Code](https://img.shields.io/badge/Claude%20Code-plugin-d97757)
![Commands](https://img.shields.io/badge/commands-15-3b6ea5)
![Skills](https://img.shields.io/badge/expert%20skills-5-3b6ea5)
![Review](https://img.shields.io/badge/second%20engine-Codex%20%7C%20Kimi-4c9a6b)
![Gates](https://img.shields.io/badge/phases-hook--enforced-c0563a)

</div>

---

## The idea in one line

> **The gates are hooks, not honor rules** — no phase starts until the previous one *proves* it finished.

`/m:develop <request>` drives a change through five gated phases, each running as a discrete Skill invocation behind a marker-file gate. A `PreToolUse` hook denies any code mutation outside `.m/` until the active phase has been entered through its skill — so the model can't skip refine, can't implement past an unapproved plan, and can't call a run "done" without the verification loop.

The phases run in order, and each must prove it finished before the next can start:

1. **refine** — grill the raw request into an execution-ready spec.
2. **plan** — build the implementation plan, second-engine sanity pass (Codex or Kimi when configured).
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

Claude refines the request into a spec, plans it (cross-checked by the configured second engine), implements against the plan, reviews the diff, and runs a verify-fix loop — pausing for your input at each gate.

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
| ① | `/m:refine` | Grills a raw request into an execution-ready spec — optimal-version reframe, then bounded-menu questions until ambiguity is gone. | Opus 5 |
| ② | `/m:plan` | Builds the implementation plan and challenges it with a second-engine (Codex or Kimi) sanity pass — grilled until zero gaps remain. | Opus 5 |
| ③ | `/m:implement` | Writes code **to the approved plan only**, following repo patterns. Plan defects escalate back to plan rather than being improvised past. | Opus 5 |
| ④ | `/m:review` · `/m:review-fanout` | Evidence-backed review. Sequential for small diffs; parallel blind-lens fan-out (security, architecture, tests, performance, migrations, observability, api-contracts, compliance) + judge for large ones. | Opus 5 |
| ⑤ | `/m:iterate` | Test-and-fix loop until the exit predicate holds (tests green · zero critical findings · progress logged · PRD criteria met). The 3-loop cap is `BLOCKED`, never `PASSED`. | Sonnet 5 |

---

## All commands

**Pipeline** — the spine above, plus the orchestrator:

| Command | Purpose | Model |
|---------|---------|:-----:|
| `/m:develop` | Run all five phases end-to-end with hard phase gates and second-engine review. | Opus 5 |
| `/m:refine` · `/m:plan` · `/m:implement` · `/m:review` · `/m:review-fanout` · `/m:iterate` | The phases, standalone (see table above). | Opus 5 / Sonnet 5 |

**Support** — build, inspect, and learn from project memory:

| Command | Purpose | Model |
|---------|---------|:-----:|
| `/m:index` | Build or refresh persistent `.m/` project memory (stack, patterns, hotspots). | Opus 5 |
| `/m:status` | "Where are we" dashboard — focus, gaps, tasks, worktrees. Logs bugs and progress. | Sonnet 5 |
| `/m:research` | Isolated worktree research for unknowns before planning — advisory only. | Opus 5 |
| `/m:analyze` | Deep analysis of code/docs/systems, with optional diagrams and grading. | Opus 5 |
| `/m:setup` | Diagnose and configure the second engine — provider, model, effort, per-repo block. | Sonnet 5 |
| `/m:feedback` | Store explicit workflow preferences (filesystem only, no inference). | Haiku 4.5 |
| `/m:learn` | Turn stored feedback signals into per-skill behavioral adaptations. | Sonnet 5 |
| `/m:help` | Print the workflow reference — order, purposes, side-effect tiers. | Haiku 4.5 |

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

**Second engine (opt-in, per repo).** Pick a provider via `.m/pipeline.yml` `second_engine.provider: codex | kimi | none` (default `none`; legacy `codex:` sections still work as a deprecated fallback). When a provider is selected it runs across the pipeline: `/m:plan` gets two blocking passes (Pass-1 architecture sanity, Pass-2 final-plan verdict), `/m:research` runs a parallel second researcher reconciled with Claude's, and `/m:review` / `/m:review-fanout` run it on every review with verdicts side-by-side. The stricter verdict always wins, and second-engine findings are leads to verify, never ground truth. Each run is **token-metered** against a budget (`token_budget`, default 200k) with graceful fallback to Claude-only. Schema and per-provider defaults: `references/pipeline-context.md`; protocols: `references/codex-protocol.md`, `references/kimi-protocol.md`. Configure interactively with `/m:setup`.

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

- **Codex CLI** (`codex` ≥ 0.123.0 on `PATH`) — the `codex` provider for the second-engine passes across `/m:plan`, `/m:research`, and review, selected via `.m/pipeline.yml` `second_engine.provider: codex`. Token-metered per run with a budget + graceful fallback; optional fast mode.
- **Kimi Code CLI** (`kimi` ≥ 0.29.0 on `PATH`) — the `kimi` provider, selected via `second_engine.provider: kimi`. Same passes and metering; review runs over a diff the pipeline prepares. **Prerequisite:** `kimi -p` auto-approves every tool call and has no read-only mode, so the protocol gates its passes on user-level deny rules for `Write`/`Edit`/`Bash` in `~/.kimi-code/config.toml`. Without them, Kimi passes are skipped and the run continues Claude-only. `/m:setup` adds the rules with your confirmation; the details and what was tested are in `references/kimi-protocol.md` §6.1. Without either CLI — or with `provider: none` (the default) — every stage runs Claude-only.
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
├── commands/          # 15 slash commands (plugin name `m` → invoked as /m:refine, /m:plan, …)
├── skills/            # 5 expert-mode skills (/m:go, /m:react, /m:biz, /m:cr, /m:security)
├── references/        # codex-protocol · kimi-protocol · jira-context · lens-templates · pipeline-context · checklists
├── rules/             # rigor · self-serve · verification · code-quality · testing  (referenced via ${CLAUDE_PLUGIN_ROOT})
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
