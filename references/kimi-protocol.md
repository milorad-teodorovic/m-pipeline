# Kimi Second-Engine Protocol (referenced from `/m:plan`, `/m:research`, `/m:review`, `/m:review-fanout`, `/m:develop`)

## Contents

- [1. Configuration: `.m/pipeline.yml` `second_engine:` section](#1-configuration-mpipelineyml-second_engine-section)
- [2. Pre-flight: Availability and Enablement](#2-pre-flight-availability-and-enablement)
- [3. Fast Mode (not applicable)](#3-fast-mode-not-applicable)
- [4. Operating-Rules Preamble (parity)](#4-operating-rules-preamble-parity)
- [5. Secret Redaction Rule (applies to every handoff write)](#5-secret-redaction-rule-applies-to-every-handoff-write)
- [6. Metered Kimi Invocation (used by every pass)](#6-metered-kimi-invocation-used-by-every-pass)
- [7. Token Metering and Budget Enforcement](#7-token-metering-and-budget-enforcement)
- [8. Pass-1: Architecture Sanity (plan, blocking)](#8-pass-1-architecture-sanity-plan-blocking)
- [9. Pass-2: Final Plan Review (plan, blocking)](#9-pass-2-final-plan-review-plan-blocking)
- [10. Disagreement Menu (plan)](#10-disagreement-menu-plan)
- [11. Research: Parallel Dual-Engine Researcher](#11-research-parallel-dual-engine-researcher)
- [12. Review: Mandatory Second Engine](#12-review-mandatory-second-engine)
- [13. Handoff Cleanup](#13-handoff-cleanup)

This reference defines the Kimi half of the second-engine handoff. It mirrors
`codex-protocol.md` section-for-section so the pipeline commands can dispatch on
`second_engine.provider` and follow the same section numbers in either file. Kimi
participation is **configuration-driven and mandatory when selected**: when
`second_engine.provider` is `kimi`, every applicable pass runs Kimi automatically with no
per-pass prompt; when the provider is `none` or `codex`, every Kimi pass in this file is
skipped.

## 1. Configuration: `.m/pipeline.yml` `second_engine:` section

All Kimi behavior is controlled by the `second_engine:` section of the per-project
`.m/pipeline.yml` (full schema, provider defaults, key-interpretation table, and the legacy
`codex:` fallback in `pipeline-context.md`). Read it once at the start of any command that
uses the second engine and hold the resolved values for the remainder of the run.

Bind the resolved values to named shell variables used by the rest of this protocol:

```bash
# Resolve config (parse .m/pipeline.yml second_engine: section; apply the
# pipeline-context.md kimi defaults when keys are absent).
KIMI_MODEL=kimi-code/k3       # from second_engine.model, default kimi-code/k3
KIMI_EFFORT=high              # from second_engine.reasoning_effort, default high (see note)
KIMI_BUDGET=200000            # from second_engine.token_budget, default 200000
KIMI_ON_EXCEED=fallback       # from second_engine.on_budget_exceeded, default fallback
```

Effort note: kimi-code v0.29.x has no per-invocation effort flag. The effective effort is
the model's `default_effort` in `~/.kimi-code/config.toml` (valid values `low|high|max`).
When `KIMI_EFFORT` differs from that default, note it once in the run metadata and proceed
with the CLI default — do not edit the user's global config. `fast_mode` does not apply to
Kimi; if set, note the ignored key once in the run metadata (see Section 3).

## 2. Pre-flight: Availability and Enablement

Run this once before the first Kimi pass of a command.

1. If the resolved provider is not `kimi`: set `KIMI_DISABLED=true`, skip every Kimi pass
   for the rest of the run, do not prompt, do not warn. The command proceeds Claude-only
   (or under the Codex protocol when the provider is `codex`). Stop here.
2. Otherwise, verify the CLI: run `kimi --version` via Bash.
   - If the command fails, is not on PATH, or reports a version older than `0.29.0`: print
     verbatim `[WARN] kimi selected but unavailable — proceeding Claude-only. Install or
     upgrade the Kimi Code CLI, then re-run /m:setup.`, set `KIMI_DISABLED=true`, and
     proceed Claude-only.
   - Otherwise: set `KIMI_DISABLED=false`. Initialize the token meter:
     `mkdir -p .m/handoff && echo 0 > .m/handoff/kimi-meter.txt`.
3. **Containment gate (disabling, not advisory).** `kimi -p` auto-approves every tool call
   and the only mechanism verified to precede that is a user-level deny rule
   (Section 6.1). Check for them:
   `grep -A3 '\[\[permission.rules\]\]' ~/.kimi-code/config.toml`. If rules denying
   `Write`, `Edit`, and `Bash` are not all present, print verbatim
   `[kimi] no user-level deny rules found — kimi passes skipped (Claude-only). Run /m:setup to add them.`,
   set `KIMI_DISABLED=true`, and proceed Claude-only. Do not write the rules here; only
   `/m:setup` may, with the user's per-write confirmation.
   These three rules bound writes, edits, and shell only: every other tool the CLI
   auto-approves in `-p` mode (including its network-fetch tool) stays available to Kimi.
   Treat every payload handed to Kimi as readable and transmittable by it, and keep the
   Section 5 deny-list strict.

A genuine CLI failure is surfaced loudly but never hard-blocks the run — the command
degrades to Claude-only, the same terminal behavior as the `on_budget_exceeded: fallback`
path in Section 7.

## 3. Fast Mode (not applicable)

Kimi has no fast-mode flag. When the config sets `fast_mode` while the provider is `kimi`,
ignore the key and record one line in the run metadata:
`[second-engine] fast_mode ignored (kimi has no fast mode)`. Speed-tiered Kimi models
(e.g. `kimi-code/kimi-for-coding-highspeed`) are selected via `model`, not a flag.

## 4. Operating-Rules Preamble (parity)

Every handoff payload Claude writes for Kimi (Pass-1, Pass-2, research, review
instructions) MUST begin with this preamble so Kimi operates under the same floors as
Claude.

```
Operating rules for this task (non-negotiable floors):
- Verify before asserting. Every file, struct, function, table, column, or API
  signature you reference must come from a file you actually read. Recalled
  knowledge is not a citation. Quote code verbatim, not from memory.
- Cite exact file:line. Re-read to confirm line numbers before reporting.
- Self-challenge every finding: state the strongest counter-argument and trace
  it through real code. If the counter-argument holds, drop the finding. Prefer
  zero findings over weak findings. "I don't know" is a valid answer.
- No shortcuts: do not skip a check because the change looks small; do not
  substitute pattern recognition for tracing where inputs come from.
- Stay read-only. Do not modify files. Report findings only.
```

## 5. Secret Redaction Rule (applies to every handoff write)

Before writing `.m/handoff/claude-to-kimi.md`, redact any value matching the deny-list
below. Replace matched values with the literal string `[REDACTED]`. This applies to every
handoff write (Pass-1, Pass-2, research, review).

- API keys: `sk-*`, `pk_*`, AWS `AKIA*` or `ASIA*`, Google `AIza*`
- ALL_CAPS environment variable values where the variable name matches `*_TOKEN`, `*_SECRET`, `*_PASSWORD`, or `*_KEY`
- Contents of any `.env*` file
- JWT tokens (tokens beginning `eyJ` with two `.` separators)
- PEM blocks (from `-----BEGIN * PRIVATE KEY-----` through the matching `-----END * PRIVATE KEY-----`)

Scope of handoff writes is plan/research/review prose only — decisions, file paths,
acceptance criteria, rationale. Never include raw source code, `.env` contents, or log
excerpts that may embed secrets. (Under Section 6.1 Kimi's working directory is the
scratch directory: it sees only what Claude writes there. The redaction rule therefore
governs every payload Claude authors — including `.m/handoff/kimi-scratch/changes.diff`,
whose context lines must never carry values matching the deny-list.)

## 6. Metered Kimi Invocation (used by every pass)

Every Kimi pass uses the single helper below. Do not call `kimi` outside this helper from
within the pipeline.

### 6.1 Security properties of `kimi -p` (read before changing this section)

`kimi -p` is **not** a sandboxed or read-only invocation, and the CLI offers no way to make
it one. These properties were verified against kimi-code v0.29.1 on 2026-07-25 and are the
reason the helper below runs every pass in a disposable scratch directory:

- Prompt mode forces `permission: "auto"` (`resolvePromptSession`, `forcePromptPermission`)
  and installs `session.setApprovalHandler(() => ({ decision: "approved" }))`.
  `AutoModeApprovePermissionPolicy` then approves **every** tool call with no context
  inspection. Verified empirically: a `-p` run instructed to create a file wrote it to disk
  with no prompt.
- `--yolo`, `--auto`, and `--plan` are all rejected when combined with `-p`
  (`Cannot combine --prompt with --plan.`), so kimi's only read-only mode is unavailable
  here. There is no equivalent of Codex's `-s read-only`.
- The `Read`/`Write`/`Edit` tools do carry a mechanical guard: `resolvePathAccess` throws
  `PathSecurityError` for sensitive basenames (`.env*`, `id_rsa`, `id_ed25519`,
  `credentials`, `*.pem`, `*.key`) because `DEFAULT_WORKSPACE_ACCESS_POLICY` sets
  `checkSensitive: true`. This fires independently of the permission chain — verified
  empirically (`"id_rsa" matches a sensitive-file pattern ... Access is blocked`).
- **`Bash` is not covered by that guard.** A shell command can read any file the user can
  read. Kimi declined to do so when asked directly, but that is model behavior, not
  enforcement — treat it as unprotected.

**What does not contain it (tested, do not re-attempt).** These were tried against v0.29.1
on 2026-07-25 and each failed to stop a write:

- Setting `cwd` to a scratch directory. `DEFAULT_WORKSPACE_ACCESS_POLICY` uses
  `guardMode: "absolute-outside-allowed"`, so only *relative* `..` traversal is refused
  (`"../x" is not an absolute path...`). Kimi canonicalizes and retries with an absolute
  path, which succeeds anywhere the user can write.
- `[[permission.rules]]` deny blocks in a project-scoped `.kimi-code/local.toml`, placed
  either in the scratch directory or at the repository root. Neither was applied to the
  `-p` run; the write still succeeded.

**What does contain it: user-level deny rules.**
`UserConfiguredDenyPermissionPolicy` is evaluated **before** `AutoModeApprovePermissionPolicy`
in `createPermissionDecisionPolicies`, so deny rules in the user's own
`~/.kimi-code/config.toml` are the only mechanism verified to sit ahead of blanket
auto-approval. The pipeline never writes that file on its own — `/m:setup` offers it with
per-write confirmation, because the rules apply to *all* Kimi usage on the machine, not
only `/m` passes:

```toml
[[permission.rules]]
decision = "deny"
scope = "user"
pattern = "Write"
reason = "m-pipeline: second-engine passes are read-only"
```

Repeat for `Edit` and `Bash`. The pattern grammar is `ToolName` or `ToolName(argPattern)`;
`decision` is one of `allow|deny|ask` and `scope` one of
`turn-override|session-runtime|project|user`.

**The gate.** Because containment depends on user configuration the pipeline does not own,
Section 2 treats its absence as a disabling condition rather than a warning: with no
user-level deny rules for `Write`, `Edit`, and `Bash`, Kimi passes are skipped and the run
proceeds Claude-only. Failing safe is deliberate — a second opinion is worth less than an
unbounded write path next to untrusted diff content.

**Scratch directory (still used).** Passes run with `cwd` set to `.m/handoff/kimi-scratch/`
and the payload written there as a file. This is not a security boundary; it keeps Kimi's
working directory free of repository content, blocks the relative-traversal case, and gives
Handoff Cleanup (Section 13) a single path to delete.

**Consequence for review depth.** Because `cwd` is the scratch directory, Kimi reviews the
diff it is given and cannot trace call sites across the repository the way Codex can under
its own sandbox. This is a deliberate trade: Kimi's findings are narrower, and (per
Section 12) they are leads to confirm rather than verdicts.

Token accounting: the stream-json stdout carries the conversation events and a final
`session.resume_hint` meta event with the session id, but no usage numbers. Per-turn usage
is recorded on disk in the session's wire log at
`~/.kimi-code/sessions/*/session_<id>/agents/*/wire.jsonl` as
`{"type":"usage.record","usage":{"inputOther":N,"output":N,"inputCacheRead":N,"inputCacheCreation":N}}`
events. Spend for the meter is `inputOther + inputCacheCreation + output` — cache reads are
excluded so budgets track real cost.

```bash
# Inputs the caller sets before invoking:
#   KIMI_PROMPT = the full pass prompt (preamble + payload + instruction)
# Plus the config vars from Section 1.

mkdir -p .m/handoff/kimi-scratch
EV=.m/handoff/kimi-events.jsonl
OUT=.m/handoff/kimi-to-claude.md
METER=.m/handoff/kimi-meter.txt
SCRATCH="$PWD/.m/handoff/kimi-scratch"
[ -f "$METER" ] || echo 0 > "$METER"

SPENT=$(cat "$METER" 2>/dev/null || echo 0)
if [ "$KIMI_DISABLED" = "true" ] || [ "$SPENT" -ge "$KIMI_BUDGET" ]; then
  echo "[kimi] budget reached or disabled — skipping pass (Claude-only)."
else
  # cwd is the scratch dir, never the repo root (Section 6.1, Layer 1).
  ( cd "$SCRATCH" && kimi -p "$KIMI_PROMPT" -m "$KIMI_MODEL" --output-format stream-json ) > "$EV" 2>&1
  RC=$?

  if [ "$RC" -ne 0 ]; then
    echo "[WARN] kimi exited $RC — proceeding Claude-only for the rest of this run."
    KIMI_DISABLED=true
    rm -f .m/handoff/claude-to-kimi.md
  else
    # Extract the final assistant message and the session id, then sum usage
    # from the session's wire.jsonl (cache reads excluded).
    THIS=$(python3 - "$EV" <<'KIMIPY'
import glob, json, os, sys
final, session = "", None
try:
    for line in open(sys.argv[1]):
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except ValueError:
            continue
        if obj.get("role") == "assistant" and isinstance(obj.get("content"), str):
            final = obj["content"]
        if obj.get("type") == "session.resume_hint":
            session = obj.get("session_id")
except OSError:
    pass
try:
    open(".m/handoff/kimi-to-claude.md", "w").write(final)
except OSError:
    pass
spend, measured = 0, False
if session:
    pattern = os.path.expanduser(f"~/.kimi-code/sessions/*/{session}/agents/*/wire.jsonl")
    for wire in glob.glob(pattern):
        try:
            for line in open(wire):
                try:
                    obj = json.loads(line)
                except ValueError:
                    continue
                if obj.get("type") == "usage.record":
                    u = obj.get("usage") or {}
                    spend += (u.get("inputOther") or 0) + (u.get("inputCacheCreation") or 0) + (u.get("output") or 0)
                    measured = True
        except OSError:
            continue
# UNMEASURED is distinct from a real zero: the caller charges the whole
# remaining budget so a broken meter cannot silently uncap spending.
print(spend if measured else "UNMEASURED")
KIMIPY
)
    if [ -z "$THIS" ] || [ "$THIS" = "UNMEASURED" ]; then
      echo "[WARN] kimi token usage could not be measured — charging the remaining budget and disabling further kimi passes this run."
      TOTAL="$KIMI_BUDGET"
      KIMI_DISABLED=true
    else
      TOTAL=$(( SPENT + THIS ))
    fi
    echo "$TOTAL" > "$METER"
    echo "[kimi] this pass: ${THIS} tok | run total: ${TOTAL}/${KIMI_BUDGET} tok"
  fi
fi
```

After every successful pass, read `.m/handoff/kimi-to-claude.md` for the agent's final
message, then apply the Section 7 budget check before the next pass.

## 7. Token Metering and Budget Enforcement

`.m/handoff/kimi-meter.txt` accumulates the cumulative Kimi token total across every pass
in a single `/m` run (plan Pass-1 + Pass-2, research, review). The budget (`KIMI_BUDGET`,
default 200000) is the ceiling.

After each pass updates the meter, evaluate:

- If `TOTAL < KIMI_BUDGET`: continue normally.
- If `TOTAL >= KIMI_BUDGET`, apply `KIMI_ON_EXCEED`:
  - **`fallback`** (default): print verbatim `[kimi] token budget ${KIMI_BUDGET} reached — finishing this run Claude-only.`, set `KIMI_DISABLED=true`, and complete all remaining passes Claude-only.
  - **`stop`**: print verbatim `[kimi] token budget ${KIMI_BUDGET} reached — stopping and saving progress.`, persist current state, write a one-line resume note to `.m/PROGRESS.md`, run Handoff Cleanup (Section 13), and halt the command.

The meter is per run. The first pass of a fresh `/m` command initializes it to `0`
(Section 2). Handoff Cleanup (Section 13) removes the meter file at every terminal path.

## 8. Pass-1: Architecture Sanity (plan, blocking)

Runs after observation gathering completes, before Phase 2 begins. Phase 2 must not start
until Pass-1 completes or is skipped (skipped only when `KIMI_DISABLED=true`).

1. Run `mkdir -p .m/handoff` via Bash.
2. Build the handoff payload: the Operating-Rules Preamble (Section 4), then the raw
   request, every `[OBSERVATION]` entry gathered above, plus this instruction:

   > Review these observations against the repository. Identify missing architecture concerns, risks, or alternative approaches that the driver agent has not surfaced. Do not propose a full plan yet. Return a bulleted list of additional concerns, each with a concrete `file:line` reference when applicable.

3. Apply the Secret Redaction Rule (Section 5), then write the result to `.m/handoff/claude-to-kimi.md`.
4. Set `KIMI_PROMPT="$(cat .m/handoff/claude-to-kimi.md)"`, then run the Metered Kimi Invocation (Section 6).
5. Read `.m/handoff/kimi-to-claude.md`.
6. Merge each Kimi-surfaced concern into the observation list as a new `[OBSERVATION — kimi]` entry, preserving any `file:line` references verbatim.
7. Apply the Section 7 budget check. Leave the handoff files in place; they are cleaned up at the end of Phase 3.

The merged observation list is the input to Phase 2.

## 9. Pass-2: Final Plan Review (plan, blocking)

Runs after the three exit-gate checks pass, before the plan document is emitted.

1. Build the handoff payload: the Operating-Rules Preamble (Section 4), the full final plan
   draft, plus this verdict-format instruction appended at the end:

   > End your output with a line of exactly `VERDICT: LGTM` if you have no disagreements with this plan, or `VERDICT: CHANGES` if you do. If `VERDICT: CHANGES`, list each disagreement above the verdict line as a numbered item: `Dn. <short description> | claude: <claude-pos + one-line rationale> | kimi: <kimi-pos + one-line rationale>`. Include concrete `file:line` references where applicable.

2. Apply the Secret Redaction Rule. Write it to `.m/handoff/claude-to-kimi.md` (overwriting any Pass-1 content).
3. Set `KIMI_PROMPT="$(cat .m/handoff/claude-to-kimi.md)"`, then run the Metered Kimi Invocation (Section 6).
4. **First check `KIMI_DISABLED`.** If the Section 6 invocation flipped it to `true` (a
   mid-pass CLI failure), do not read or parse a verdict — the `kimi-to-claude.md` on disk
   is stale Pass-1 content. Skip directly to the degradation below. Otherwise, read
   `.m/handoff/kimi-to-claude.md` and parse the final non-empty line:
   - Exactly `VERDICT: LGTM`: finalize tasks, run Handoff Cleanup, emit the plan.
   - Exactly `VERDICT: CHANGES`: parse the numbered disagreements and enter the Disagreement Menu (Section 10).
   - Neither verdict string on the last non-empty line: treat as `VERDICT: CHANGES` (conservative default). If no parseable disagreements are listed, hard-block and re-grill the user on the final plan content.
5. Apply the Section 7 budget check.

If `KIMI_DISABLED=true` (disabled, CLI failure, or budget reached): skip Pass-2, finalize
tasks, run Handoff Cleanup, emit the plan.

## 10. Disagreement Menu (plan)

For each numbered disagreement `Dn` Kimi listed, present a single bounded-menu question to
the user. Source tags are mandatory and must appear verbatim:

```
Decision D{n}: {short description}
  [claude] {claude-pos + rationale}
  [kimi]   {kimi-pos + rationale}

Pick:
  A) claude
  B) kimi
  C) merge — describe
  D) neither — describe
```

Collect the user's picks for all disagreements. Apply them to the plan draft in order.

After all picks are applied, run one final Kimi verification pass if the token budget still
permits it:

- If `KIMI_DISABLED=true` or the meter is at/over budget: skip the final verification, run Handoff Cleanup, emit the plan.
- Otherwise: write the merged plan to `.m/handoff/claude-to-kimi.md` (with redaction and preamble), run the Metered Kimi Invocation, parse the verdict.
  - `VERDICT: LGTM` → run Handoff Cleanup, emit the plan.
  - `VERDICT: CHANGES` → **hard-block**. Convert the remaining disagreements into new Phase 2 grill questions. Do NOT invoke Kimi again for the remainder of this run (set `KIMI_DISABLED=true`). Loop back to Phase 2.

## 11. Research: Parallel Dual-Engine Researcher

Used by `/m:research` (and by `/m:plan`'s worktree research spawn) when the provider is
`kimi`. Kimi researches the same questions independently and in parallel with Claude's
research agent; Claude then reconciles both into one finding set.

1. After forming the 2–5 focused research questions, write them (plus relevant file paths
   and the Operating-Rules Preamble) to `.m/handoff/claude-to-kimi.md`, redacted per
   Section 5, with this instruction appended:

   > Research these questions against the local repository, official documentation, and primary sources. Separate stable local observations, external facts (with versions/dates), and recommendations. For each question, give a decisive recommendation plus the strongest counter-argument. Cite `file:line` for local claims and source URLs for external claims. List remaining unknowns explicitly.

2. Run Claude's research agent and the Metered Kimi Invocation (Section 6) concurrently when possible.
3. Read `.m/handoff/kimi-to-claude.md`. **Reconcile**:
   - Where both engines agree, mark the finding `[CORROBORATED]`.
   - Where they disagree, present both positions side-by-side — `[claude]` vs `[kimi]` — and do not silently pick a winner. The user decides what to adopt (research is advisory).
   - Merge unique findings from each engine, preserving citations verbatim.
4. The reconciled set is the research output. Apply the Section 7 budget check; if the
   budget is reached mid-research, finish reconciliation with whatever Kimi returned and
   continue Claude-only per `KIMI_ON_EXCEED`.

If `KIMI_DISABLED=true`, run Claude-only research exactly as before.

## 12. Review: Mandatory Second Engine

Used by `/m:review` and `/m:review-fanout`. When the provider is `kimi`, the Kimi review
runs on **every** review — no prompt, no high-stakes gating. Kimi has no native review
subcommand and no read-only sandbox (Section 6.1), so **Claude produces the diff and Kimi
reads it from the scratch directory**. Do not instruct Kimi to run `git` itself: that
requires the `Bash` tool, which carries no sensitive-file guard.

1. After Claude's primary review produces its verdict, Claude generates the diff for the
   target and writes it to `.m/handoff/kimi-scratch/changes.diff`:
   - Uncommitted changes: `git diff HEAD` plus the untracked files listed by `git status --short`
   - Against a base branch: `git diff <branch>...HEAD`
   - A specific commit: `git show <SHA>`

   Apply the Secret Redaction Rule (Section 5) to the diff before writing it. A diff is
   source code, so the "never include raw source" scope note in Section 5 does not apply
   here — the diff is the review subject — but every deny-list pattern still does.

2. Build `KIMI_PROMPT` as the Operating-Rules Preamble (Section 4), then:

   > The file `changes.diff` in your working directory is the complete change set under review. Read it. Review only this change set. Focus on: correctness bugs, security threats, regressions, and clean-code violations introduced by these changes. Report each finding with `file:line` (as named in the diff) and a verbatim snippet from the diff. Do not restate unchanged code. Do not run shell commands; everything you need is in `changes.diff`.

3. Run the Metered Kimi Invocation (Section 6).
4. Present Claude's findings and Kimi's findings side-by-side under a dedicated
   **Second Opinion (Kimi)** section. Do not merge silently. For `/m:review-fanout`, place
   this alongside the judge verdict, not inside the judge's merged list.
5. **Disagreement rule:** if Claude says `APPROVED` but Kimi flags criticals that survive
   re-verification (the hallucination filter applied to Kimi's findings too), downgrade to
   `BLOCKED`. The more permissive verdict never wins by default. Kimi findings are leads to
   confirm against real code, never ground truth.
6. Apply the Section 7 budget check. If `KIMI_DISABLED=true` (disabled, CLI failure, or
   budget reached), note it in the Review Metadata and skip the second engine — do not fail
   the review.

## 13. Handoff Cleanup

At the end of every command that used this protocol — on every terminal path — delete the
handoff and meter files if they exist:

```bash
rm -f .m/handoff/claude-to-kimi.md .m/handoff/kimi-to-claude.md \
      .m/handoff/kimi-events.jsonl .m/handoff/kimi-meter.txt
rm -rf .m/handoff/kimi-scratch
```

Removing the scratch directory is what discards anything Kimi wrote during the run
(Section 6.1, Layer 1). Do not persist any dialogue artifact. Only the final
plan/research/review output remains.
