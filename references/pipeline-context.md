# Per-project pipeline config: `.m/pipeline.yml`

The `/m:*` pipeline reads project-specific compliance and high-stakes triggers from a
per-repository `.m/pipeline.yml` file, so the global commands stay project-agnostic. This
mirrors the per-project Jira mapping documented in `jira-context.md`. Each project that
needs compliance review or extra-rigorous gating on sensitive paths declares its own file
under `.m/` at the repository root.

## Schema

```yaml
compliance:
  enabled: true                          # when true, /m:review Pass 4 and the
                                         # /m:review-fanout compliance lens fire automatically
  frameworks: [SOC2, GDPR, "EU AI Act"]  # informational; named in the compliance pass output
high_stakes_paths:                       # repo-root-relative globs; a changed file matching
  - "server/pkg/crypto/envelope/**"      # any glob triggers the Codex second-opinion gate and
  - "**/erasure_service.go"              # the large / security-sensitive size tier
```

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
   reviewer knows which obligations apply.

## Relationship to other `.m/` files

- `.m/jira.yml` — per-project Jira mapping (see `jira-context.md`).
- `.m/pipeline.yml` — per-project compliance and high-stakes triggers (this file).

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
