---
description: Diagnose, then interactively configure (guided wizard — provider → model → effort → provider extras) the second review/plan engine for the /m pipeline — codex, kimi, or none. Covers CLI, auth, global config, per-repo second_engine block. Use to check "is the second engine ready for /m", to switch providers, or after installing/upgrading a CLI.
argument-hint: "[--check]   (read-only diagnosis; omit to run the guided config wizard)"
model: claude-sonnet-5
effort: medium
allowed-tools: Read, Grep, Glob, Edit, Write, Bash(codex --version), Bash(codex features list:*), Bash(kimi --version), Bash(kimi doctor), Bash(mkdir:*)
---
# /m:setup - Second-Engine Setup Doctor

Diagnose the second-engine prerequisites the `/m` pipeline relies on for dual-engine planning, research, and review; report each check; and — unless `--check` is passed — run a guided wizard that asks which provider to use (codex, kimi, or none) and, per provider, its model and effort, then applies the choices only after explicit, per-change confirmation.

## Input

Flags: `$ARGUMENTS`
- _(no flag)_ — run the diagnosis (Phase 1), then the **guided configuration wizard** (Phase 2).
- `--check` — read-only. Run the diagnosis and print remediation, but never run the wizard or apply any change.

## Safety contract

- **Diagnosis is read-only.** Every value reported must come from a command run or a file read in this session — never assert a config value without reading it.
- **No silent writes.** Any change to `~/.codex/config.toml`, `~/.kimi-code/config.toml`, or a repo's `.m/pipeline.yml` is shown verbatim (exact lines) and applied only after the user confirms that specific change. `--check` disables the wizard and all fix offers.
- **Never touch credentials.** Do not read `~/.codex/auth.json` wholesale (it holds tokens) — extract only the `auth_mode` value with `grep -oE`. Never read `~/.kimi-code/credentials` or `~/.kimi-code/oauth`. Never run `codex login` or `kimi login` for the user; print the command for them to run.
- Before editing a global config, back it up: `~/.codex/config.toml` → `~/.claude/backups/codex-config-<timestamp>.toml`, `~/.kimi-code/config.toml` → `~/.claude/backups/kimi-config-<timestamp>.toml`.

## Phase 1 — Diagnose (always runs, read-only)

Run each check and record `✓` (ready) / `⚠` (works, sub-optimal) / `✗` (broken, needs action).

1. **Resolved provider** — walk up from the cwd to the nearest `.m/`; read its `pipeline.yml` and report the resolved `second_engine:` values (applying the legacy `codex:` fallback per `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md`), or `absent → provider: none (Claude-only)`. Note when the legacy mapping is in use.
2. **Codex CLI** (when relevant): `codex --version` — `✓` if `0.123.0` or newer · `✗` if missing/old → remediation `npm install -g @openai/codex@latest`.
3. **Codex auth mode**: `grep -oE '"auth_mode"[[:space:]]*:[[:space:]]*"[^"]*"' ~/.codex/auth.json` — `✓` `chatgpt` (fast-mode eligible) · `⚠` `apikey` (fast mode falls back to standard pricing) · `✗` absent → remediation `codex login`.
4. **Codex global config**: `grep -nE '^model|^model_reasoning_effort|^service_tier|fast_mode|^\[features\]' ~/.codex/config.toml` — report values. `model` is informational only: report it, never flag it or propose changing it.
5. **Kimi CLI** (when relevant): `kimi --version` — `✓` if `0.29.0` or newer · `✗` if missing/old → remediation: install/upgrade the Kimi Code CLI, then `kimi login`.
6. **Kimi config health**: `kimi doctor` — report its verdict. Then `grep -nE '^default_model|^effort|^model|support_efforts|default_effort' ~/.kimi-code/config.toml` — report `default_model` and each model alias with its `default_effort` (valid efforts `low|high|max`).
7. **Statusline snapshot** (codex only) — report whether `~/.claude/.codex-limits.json` exists (populates on the first metered Codex run; absence is normal before any run).

Run checks 2–4 when the resolved or candidate provider is codex, 5–6 when kimi; on a plain diagnosis with no provider resolved, run all of them so the user sees both options' readiness. Emit a readiness table: one row per check with `✓/⚠/✗`, current value, recommended value.

## Phase 2 — Guided configuration wizard (skipped entirely when `--check` is present)

Before asking, **read the current values to pre-fill every default**: the resolved `second_engine:` block (or the built-in defaults in `${CLAUDE_PLUGIN_ROOT}/references/pipeline-context.md` when absent), `~/.codex/config.toml`, and `~/.kimi-code/config.toml`. Present each current value as the pre-selected default. Never ask for a value you can read.

Ask these in order, one bounded-menu question per step:

1. **Provider — codex, kimi, or none?** → sets `second_engine.provider`.
   - **none** → write only `provider: none` to the per-repo block; the remaining steps are skipped and both global configs are left untouched. Every second-engine pass in `/m:plan|research|review|review-fanout` is then skipped silently. Go to the write step.
   - **codex** → continue with steps 2c–4c.
   - **kimi** → continue with steps 2k–3k.
2c. **Codex fast mode — on or off?** → sets `second_engine.fast_mode` and the global `service_tier` / `[features] fast_mode`. Note: fast mode is ~1.5× faster at ~2.5× the credit rate and requires ChatGPT auth (check 3); API-key auth falls back to standard pricing.
3c. **Codex model?** → sets `second_engine.model` and global `model`. Present the current value as the default plus a non-exhaustive menu, and allow free-form entry of any id the CLI accepts (`codex -m <model>`): `gpt-5.6-sol` (pipeline default) · `gpt-5.5` · or type any other id. These names are a dated snapshot (checked 2026-07-25); the authoritative list is at `developers.openai.com/codex/config-reference`. Do not hard-fail an unknown id.
4c. **Codex reasoning effort?** → sets `second_engine.reasoning_effort` and global `model_reasoning_effort`. Valid values depend on the model family (gpt-5.6 snapshot: `low|medium|high|xhigh|max|ultra`; `ultra` is unsuitable for metered `/m` passes). Present the current value as the default.
2k. **Kimi model?** → sets `second_engine.model`. Menu from the model aliases actually defined in `~/.kimi-code/config.toml` (e.g. `kimi-code/k3`, `kimi-code/k3-256k`, `kimi-code/kimi-for-coding`, `kimi-code/kimi-for-coding-highspeed`), current `default_model` pre-selected.
3k. **Kimi effort?** → sets `second_engine.reasoning_effort` (valid `low|high|max`). Note for the user: kimi-code has no per-invocation effort flag, so the pipeline follows the model's `default_effort` in `~/.kimi-code/config.toml`; if the chosen value differs from the current `default_effort`, offer to update that key in the global config (confirmed write, backed up) so the choice takes effect.
4k. **Add Kimi deny rules?** (required for Kimi to run) → `kimi -p` auto-approves every tool call, offers no read-only mode, and neither scratch-directory isolation nor project-scoped config contains its writes (all tested — `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md` Section 6.1). User-level deny rules are the only verified containment, because `UserConfiguredDenyPermissionPolicy` is evaluated before `AutoModeApprovePermissionPolicy`. Without them the protocol's Section 2 gate skips every Kimi pass and the pipeline runs Claude-only. Offer to append these to `~/.kimi-code/config.toml` (confirmed write, backed up), one block per tool for `Write`, `Edit`, and `Bash`:
   ```toml
   [[permission.rules]]
   decision = "deny"
   scope = "user"
   pattern = "Write"
   reason = "m-pipeline: second-engine passes are read-only"
   ```
   State plainly that these rules apply to **all** Kimi usage on this machine, not only `/m` passes, so a user who also uses Kimi interactively for editing should decline — and that declining means Kimi passes stay disabled and the pipeline runs Claude-only.

**Write step** — show the exact changes verbatim, then apply only on an explicit yes (each write confirmed separately):

- **Per-repo `second_engine:` block** (`./.m/pipeline.yml`): build the block from the answers. Preserve any existing `compliance:` / `high_stakes_paths:` content and the keys the wizard does not touch — keep the current `token_budget` / `on_budget_exceeded`, defaulting to `200000` / `fallback` when absent. If a legacy `codex:` block exists, offer to replace it with the equivalent `second_engine:` block (shown verbatim); if the user declines, leave it and note that the legacy fallback still applies. Create the file (and `.m/`) if missing; otherwise edit surgically. Example for a kimi run:
  ```yaml
  second_engine:
    provider: kimi
    model: kimi-code/k3
    reasoning_effort: high
    token_budget: 200000
    on_budget_exceeded: fallback
  ```
- **Global `~/.codex/config.toml`** (codex only): show the exact `model`, `model_reasoning_effort`, `service_tier`, and `[features] fast_mode` lines to add or change. On yes: back up first, apply with Edit/Write, then re-run `codex features list | grep -E '^fast_mode'` and report the new effective state.
- **Global `~/.kimi-code/config.toml`** (kimi only, and only when step 3k chose a different effort): show the exact `default_effort` line change under the chosen model's `[models."<alias>"]` section. On yes: back up first, apply with Edit/Write, then re-run the check-6 grep and report the new value.
- **CLI / auth** (`✗` on the CLI or auth checks): never auto-run — print the install/login commands for the user to run, then suggest re-running `/m:setup`.

If the user declines a specific write at its confirmation, skip that write, leave the file unchanged, and report it as declined.

## Output

A summary with: each check's `✓/⚠/✗`, the wizard answers and exactly what was written to `./.m/pipeline.yml` and the global configs (each only after the user confirmed it — note any write the user declined), and what the user must still run manually (CLI install, login). Close by noting that `/m:plan`, `/m:research`, `/m:review`, and `/m:review-fanout` use the second engine automatically when `second_engine.provider` is `codex` or `kimi`, token-metered per run (see `${CLAUDE_PLUGIN_ROOT}/references/codex-protocol.md` / `${CLAUDE_PLUGIN_ROOT}/references/kimi-protocol.md`).

## Rules

- Apply `${CLAUDE_PLUGIN_ROOT}/rules/rigor.md` and `${CLAUDE_PLUGIN_ROOT}/rules/self-serve.md`. Resolve every Phase-1 check by running its command or reading the file — quote exact current vs recommended values, never guess. The only user-facing questions permitted are the Phase-2 wizard's configuration questions and the per-write confirmations — and all of these only when `--check` is absent. Read the current config values to pre-fill every wizard default; never ask for a value you can read.
