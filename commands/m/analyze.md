---
description: Deep analysis of code, docs, systems, or flows with cached results and optional D2/Typst rendering. Use for architecture deep-dives, maintainability grading, or when "analyze X" is asked.
argument-hint: [prompt|setup|list|open <slug>]
model: opus
effort: high
allowed-tools: Read, Grep, Glob, Bash(git:*), Bash(d2:*), Bash(typst:*), Agent
---
# /m:analyze - Deep Analysis Engine

Analyze code, docs, systems, or concepts. Cache the result when useful and optionally render diagrams or reports.

## Input

Analysis prompt or subcommand: `$ARGUMENTS`

## Subcommands

If `$ARGUMENTS` is one of these, handle it directly:

- `setup`: check whether `d2` and `typst` are installed; report versions if found and provide install commands if missing. Do not auto-install unless the user explicitly asks.
- `list`: list cached analyses from `.m/analyses/`
- `open <slug>`: read `.m/analyses/<slug>/analysis.json` and `summary.md`, summarize them, and continue from the cached result instead of re-analyzing

## Context Sources

Read these when relevant:

- `.m/INDEX.md`
- `.m/RESEARCH.md`
- `PROJECT_INDEX.md`
- `AGENTS.md`
- `CLAUDE.md`
- the files, directories, or documents named in the prompt

## Workflow

1. Parse the prompt and identify:
   - the input sources
   - the analysis domain
   - the decisions this analysis should inform
   - the likely deliverables
2. Classify the repo or artifact before drawing conclusions:
   - normal source repo
   - extracted source-map or decompiled snapshot
   - vendor-heavy tree
   - generated or bundled output
3. Separate first-party code from vendor, generated, binary, and cached artifacts
4. Read only the relevant inputs
5. When analyzing codebases, explicitly map:
   - entrypoints and runtime control flow
   - orchestration layers and state ownership
   - service, tool, and plugin boundaries
   - persistence, config, and environment surfaces
   - test surfaces and verification limits
6. Collect maintainability signals when relevant:
   - largest files or hotspots
   - TODO/FIXME density
   - tests present or absent
   - global mutable state or singleton patterns
   - sync filesystem usage in interactive or hot paths
   - environment flag sprawl
7. Ask at most 3 clarification questions if the scope, audience, or output is unclear
8. Run multiple focused analysis passes
9. For very large analysis requests, you may use focused subagents if the runtime supports them cleanly; otherwise do the passes yourself
10. Synthesize the result into one coherent analysis
11. If the user asks for grading, or if a grade would materially help, grade:
   - architecture
   - code style
   - maintainability
   - repo health
   - overall
12. Cache the result under `.m/analyses/<slug>/`

## Cache Files

When caching, write:

- `analysis.json`
- `summary.md`

If the user asks for rendered output, also save any generated `.d2`, `.svg`, `.typ`, `.pdf`, or markdown artifacts in the same directory.

## Rendering Rules

- If the user asks for diagrams, generate D2 source and render SVG when `d2` is available
- If the user asks for a report, generate Typst source and render PDF when `typst` is available
- If a renderer is missing, save the source file and tell the user exactly how to render it locally

## Output

Produce:

## Analysis Complete

### Scope
### Repo Health
### Architecture
### Style
### Key Findings
### Recommendations
### Risks
### Grades
### Open Questions
### Cached Output

If the user did not specify an output format, ask what they want next after presenting the summary.

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` for the entire analyze run. No shortcuts: do not grade architecture, style, or maintainability without Reading first-party code in this session; do not skip the maintainability-signal collection because it "feels obvious"; do not reuse cached analysis when the underlying code has changed. Use tools fully: spawn `Explore` (`model: haiku` — read-only breadth sweep) for breadth, run `d2`/`typst` for diagrams when requested, prefer the docs MCP for external references. Do not compress reasoning to save tokens — analysis is the input that downstream `/m:plan` and `/m:review` runs anchor to.
- Prefer primary sources and local code over generic summaries
- Make uncertainty explicit
- If the repo is incomplete, extracted, vendored, or not buildable from source, say that explicitly and lower confidence where appropriate
- Do not treat vendor or generated code as representative of project style unless the user asked for that specifically
- Reuse cached analysis when the user asks for a format change instead of re-running the whole analysis
