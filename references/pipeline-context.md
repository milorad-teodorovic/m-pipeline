# Per-project pipeline config: `.m/pipeline.yml`

The `/m:*` pipeline reads project-specific compliance and high-stakes triggers from a
per-repository `.m/pipeline.yml` file, so the global commands stay project-agnostic. This
mirrors the per-project Jira mapping documented in `jira-context.md`. Each project that
needs compliance review, extra-rigorous gating on sensitive paths, or a second review
engine declares its own file under `.m/` at the repository root.

## Schema

```yaml
compliance:
  enabled: true                          # when true, /m:review Pass 4 and the
                                         # /m:review-fanout compliance lens fire automatically
  frameworks: [SOC2, GDPR, "EU AI Act"]  # informational; named in the compliance pass output
  rules:                                 # optional; concrete repo-specific checklist items,
    - "restricted files are never edited"      # each traced against the diff in /m:review
    - "*Enc fields use envelope encryption"    # Pass 4 and the /m:review-fanout compliance lens
high_stakes_paths:                       # repo-root-relative globs; a changed file matching
  - "server/pkg/crypto/envelope/**"      # any glob escalates Claude's pass depth and the
  - "**/erasure_service.go"              # large / security-sensitive size tier
second_engine:                           # second review/plan/research engine
  provider: none                         # codex | kimi | none (default). none = every
                                         # second-engine pass is skipped (Claude-only)
  model: gpt-5.6-sol                     # model passed to the provider CLI (see defaults below)
  reasoning_effort: high                 # effort passed to the provider (see interpretation table)
  fast_mode: false                       # codex only; ignored for kimi with a run-metadata note
  token_budget: 200000                   # cumulative second-engine tokens allowed per /m run
  on_budget_exceeded: fallback           # "fallback" (finish Claude-only) | "stop" (halt and save)
```

## Second engine: providers, defaults, and key interpretation

`second_engine.provider` selects which protocol file drives the passes:

| Provider | Protocol file | CLI floor |
|---|---|---|
| `codex` | `codex-protocol.md` | codex-cli `0.123.0` |
| `kimi` | `kimi-protocol.md` | kimi-code `0.29.0`, plus user-level deny rules — see below |
| `none` | — (every second-engine pass skipped silently) | — |

Defaults when `provider` is set but other keys are absent:

| Key | codex | kimi |
|---|---|---|
| `model` | `gpt-5.6-sol` | `kimi-code/k3` |
| `reasoning_effort` | `high` | `high` |
| `fast_mode` | `false` | — (not supported) |
| `token_budget` | `200000` | `200000` |
| `on_budget_exceeded` | `fallback` | `fallback` |

**Kimi prerequisite.** `kimi -p` auto-approves every tool call and has no read-only mode,
so `kimi-protocol.md` Section 2 gates its passes on user-level deny rules for `Write`,
`Edit`, and `Bash` in `~/.kimi-code/config.toml`. Without them, selecting `provider: kimi`
is accepted but every Kimi pass is skipped and the run proceeds Claude-only. `/m:setup`
offers to add the rules with per-write confirmation. Codex needs no equivalent: it runs
under its own `-s read-only` sandbox.

Key interpretation is provider-specific. A key that does not apply to the active provider
is ignored, and the run metadata notes the ignored key — it is never an error:

- `model` — codex: any model id the CLI accepts (`codex -m <model>`). kimi: a model alias
  defined in `~/.kimi-code/config.toml` (`kimi -m <alias>`), e.g. `kimi-code/k3`,
  `kimi-code/k3-256k`, `kimi-code/kimi-for-coding`.
- `reasoning_effort` — codex: passed as `model_reasoning_effort`; valid values depend on
  the model family (per the OpenAI config reference; for the gpt-5.6 family the snapshot
  checked 2026-07-25 is `low|medium|high|xhigh|max|ultra`, with `ultra` unsuitable for
  metered `/m` passes). kimi: valid values are `low|high|max`, but kimi-code v0.29.x has no
  per-invocation effort flag — effort follows the model's `default_effort` in
  `~/.kimi-code/config.toml`. When the configured value differs from that default, the run
  metadata notes the difference and the pass proceeds with the CLI default; `/m:setup` can
  change `default_effort` with per-write confirmation.
- `fast_mode` — codex only (per-invocation flags, `codex-protocol.md` Section 3). Ignored
  for kimi with a one-line run-metadata note.

## Legacy `codex:` section (deprecated fallback)

Older configs carry a `codex:` section instead of `second_engine:`. It remains readable:

```yaml
codex:
  enabled: true
  fast_mode: true
  model: gpt-5.6-sol
  reasoning_effort: xhigh
  token_budget: 200000
  on_budget_exceeded: fallback
```

Resolution rules:

1. An explicit `second_engine:` section always wins, even when a `codex:` section is also
   present (`second_engine.provider: none` + `codex.enabled: true` resolves to `none`).
2. With no `second_engine:` section, `codex.enabled: true` maps to
   `second_engine.provider: codex` and the remaining `codex.*` keys map by name;
   `codex.enabled: false` maps to `provider: none`.
3. Keys the legacy section omits take the codex-provider defaults in the table above
   (`model: gpt-5.6-sol`, `reasoning_effort: high`, `fast_mode: false`). These differ from
   the values documented before 2026-07-25 (`gpt-5.5` / `xhigh` / `fast_mode: true`), so a
   partial legacy block such as `codex: {enabled: true}` alone now resolves to a different
   model and effort than it once did. Pin the values explicitly if a repository depends on
   the older ones.
4. When the legacy mapping is used, the run metadata carries a one-line deprecation note
   pointing at this section, and names the resolved model and effort so a silently
   defaulted value is visible in the run output.

## Load rules

1. Loaded by `/m:develop`, `/m:review`, and `/m:review-fanout` from the current repository
   root, the same way `.m/jira.yml` is loaded.
2. If `.m/pipeline.yml` is missing, assume `compliance.enabled: false` and an empty
   `high_stakes_paths` list. The pipeline still applies its generic high-stakes categories
   (migration touching a production table, auth / money / tenant-isolation path, public API
   contract change). Those categories are built into the commands and do not depend on this
   file.
3. `high_stakes_paths` only adds project-specific sensitive paths on top of the generic
   categories; it never removes them.
4. `compliance.frameworks` is informational only. The compliance pass logic is the same
   regardless of which frameworks are listed; the list is surfaced in the report so the
   reviewer knows which obligations apply. `compliance.rules` (optional) lists concrete,
   repo-specific checklist items; when present, `/m:review` Pass 4 and the
   `/m:review-fanout` compliance lens trace **each rule** against the diff and flag any
   violation as a `[COMPLIANCE]` finding. When absent, the compliance pass relies on
   `frameworks` plus any `.business/` specs.
5. The `second_engine:` section controls second-engine participation (planning, research,
   review). Loaded by `/m:plan`, `/m:research`, `/m:review`, `/m:review-fanout`, and
   `/m:develop`. If the file, the `second_engine:` section, and a legacy `codex:` section
   are all absent, the provider is `none` and the pipeline runs Claude-only until the repo
   opts in. The pipeline never edits the user's global `~/.codex/config.toml` or
   `~/.kimi-code/config.toml`; only the `/m:setup` wizard may, per-write confirmed and
   backed up.

## Relationship to other `.m/` files

- `.m/jira.yml` — per-project Jira mapping (see `jira-context.md`).
- `.m/pipeline.yml` — per-project compliance, high-stakes, and second-engine config (this file).

Both are optional, one concern per file, and read on a "if present, else defaults" basis.

## Note on naming

This file lives under `.m/`, not at the repository root, specifically to avoid colliding
with a root-level `project.yml`, which is the XcodeGen project manifest used by some
Swift/macOS repositories in this workspace.

## Worktrees

A git worktree has its own working directory, so a worktree that needs the same triggers as
its main checkout needs its own `.m/pipeline.yml` (or a copy of the main one). Path-based
detection used to cover worktrees implicitly; config-based detection requires the file to be
present in whatever directory the pipeline runs from.
