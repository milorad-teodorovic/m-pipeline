#!/usr/bin/env python3
"""
PreToolUse hook for Edit, Write, and MultiEdit.

Purpose: enforce the /m:develop phase protocol. When /m:develop is active
(marker file `.m/DEVELOP_ACTIVE` present in the project root), this hook
blocks Edit/Write/MultiEdit outside `.m/` unless the current required phase
has been entered via its corresponding /m:* skill (marker file
`.m/phase-<name>-started` present).

Scope: only the Edit, Write, and MultiEdit tools are gated (see
`ALLOWED_TOOLS`). NotebookEdit (which carries `notebook_path`, not
`file_path`) and Bash-driven writes are intentionally NOT gated — the gate
is a cooperative guardrail for the pipeline's own file tools, not a sandbox
against arbitrary mutation.

Marker protocol
---------------

`.m/DEVELOP_ACTIVE` is written by `/m:develop` on pipeline entry and
deleted on pipeline exit. It is a single-line YAML-ish file:

    current_phase: <refine|plan|implement|review|iterate>

Each `/m:*` phase skill is required to:

- On entry: touch `.m/phase-<name>-started`
- On successful completion: touch `.m/phase-<name>-done`

`/m:develop` updates `current_phase` in `.m/DEVELOP_ACTIVE` at every stage
transition, and verifies the prior phase's `-done` marker before moving on.

Allow list
----------

Writes to the following paths always pass, because the protocol itself
must be able to write them:

- Anything under `.m/` (PLAN.md, REFINE.md, PROGRESS.md, phase markers,
  handoff dirs, etc.)
- The `DEVELOP_ACTIVE` marker itself

Everything else under the project root is blocked until the active phase
is entered via its skill.

Denial signals
--------------

Every denial appends one JSON record to
`~/.claude/m-learning/signals/gate-denials.jsonl`, which `/m:learn` reads
alongside the other signal files. A denial is the only record of pipeline
discipline that the agent being gated does not author, so it is kept even
though the hook is otherwise stateless.

The decision is always emitted and flushed before any signal work begins,
so neither a raising nor a blocking signal write can suppress or delay it —
see `emit_deny` and `record_denial`.
"""

import json
import os
import sys
from datetime import datetime, timezone


ALLOWED_TOOLS = {"Edit", "Write", "MultiEdit"}

KNOWN_PHASES = {"refine", "plan", "implement", "review", "iterate"}

SIGNAL_FILE = "gate-denials.jsonl"
MAX_RECORD_BYTES = 4096
MAX_PATH_CHARS = 200


def find_project_root(start_dir: str):
    """Walk up from start_dir looking for `.m/DEVELOP_ACTIVE`.

    Returns the directory that contains `.m/DEVELOP_ACTIVE`, or None if no
    such directory is found before reaching the filesystem root. The walk
    terminates at the root (parent == current); there is no fixed depth cap,
    so a deeply nested cwd cannot silently slip past the gate.
    """
    current = os.path.abspath(start_dir)
    while True:
        candidate = os.path.join(current, ".m", "DEVELOP_ACTIVE")
        if os.path.isfile(candidate):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return None
        current = parent


def parse_current_phase(active_path: str):
    """Parse `current_phase:` out of the DEVELOP_ACTIVE marker."""
    try:
        with open(active_path, "r", encoding="utf-8") as handle:
            for raw in handle:
                line = raw.strip()
                if line.startswith("current_phase:"):
                    return line.split(":", 1)[1].strip()
    except OSError:
        return None
    return None


def is_allowlisted(target_path: str, project_root: str) -> bool:
    """True when the write target is within the pipeline bookkeeping area.

    Uses realpath (not abspath) so a symlink planted inside `.m/` cannot
    point outside the bookkeeping area and smuggle a write past the gate.
    """
    if not target_path:
        return True
    abs_target = os.path.realpath(target_path)
    m_dir = os.path.realpath(os.path.join(project_root, ".m"))
    return abs_target == m_dir or abs_target.startswith(m_dir + os.sep)


def relative_target(target_path: str, project_root: str):
    """Repository-relative form of a write target, capped at MAX_PATH_CHARS.

    Returns None for an empty target, and falls back to the bare basename
    when the target resolves outside the project root. An absolute path is
    never returned, because the signal log is global and spans every
    repository the user works in.
    """
    if not target_path:
        return None
    try:
        relative = os.path.relpath(
            os.path.realpath(target_path), os.path.realpath(project_root)
        )
    except ValueError:
        return os.path.basename(target_path)[:MAX_PATH_CHARS]
    if relative.startswith(".."):
        return os.path.basename(target_path)[:MAX_PATH_CHARS]
    return relative[:MAX_PATH_CHARS]


def recorded_phase(phase):
    """The phase name safe to write into the global signal log.

    Returns None when no phase was parsed, the phase itself when it names a
    known pipeline phase, and the literal `unrecognized` otherwise. The
    marker file this value comes from is writable by the agent the gate
    denies, and the log is read back into `/m:learn`, so arbitrary marker
    text must never reach it.
    """
    if phase is None:
        return None
    return phase if phase in KNOWN_PHASES else "unrecognized"


def record_denial(project_root, phase, tool_name, target_path, reason) -> None:
    """Append one gate-denial record to the global learning-signal log.

    `reason` is `missing_phase_marker` when the active phase was never
    entered through its skill, or `unreadable_marker` when DEVELOP_ACTIVE
    is present but carries no parseable `current_phase:` line, in which
    case `phase` is None.

    Must only be called after `emit_deny` has already flushed the decision.
    Nothing here can suppress a denial, but it can still fail or stall, so
    ordering is the guarantee rather than the exception handling: a failure
    returns after one stderr line, and the log is opened non-blocking so a
    FIFO or similar planted at the signal path fails fast instead of hanging
    the hook.

    `phase` is recorded only when it names a known pipeline phase. Anything
    else becomes the literal `unrecognized`, because the marker file is
    writable by the very agent this gate denies and the log is read back
    into `/m:learn`.

    The record is built in full and written with a single append so
    concurrent sessions cannot interleave partial lines; a record that would
    exceed MAX_RECORD_BYTES is dropped rather than split.
    """
    try:
        home = os.path.expanduser("~")
        if home == "~" or not os.path.isdir(home):
            return
        signals_dir = os.path.join(home, ".claude", "m-learning", "signals")
        os.makedirs(signals_dir, exist_ok=True)
        record = {
            "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "type": "gate_denial",
            "reason": reason,
            "project": os.path.basename(project_root),
            "phase": recorded_phase(phase),
            "tool": tool_name,
            "path": relative_target(target_path, project_root),
        }
        line = json.dumps(record, separators=(",", ":")) + "\n"
        if len(line.encode("utf-8")) > MAX_RECORD_BYTES:
            return
        flags = os.O_WRONLY | os.O_APPEND | os.O_CREAT | getattr(os, "O_NONBLOCK", 0)
        handle = os.open(os.path.join(signals_dir, SIGNAL_FILE), flags, 0o600)
        try:
            os.write(handle, line.encode("utf-8"))
        finally:
            os.close(handle)
    except Exception as exc:
        try:
            print(
                f"[enforce-develop-phase] denial signal not recorded: "
                f"{type(exc).__name__}",
                file=sys.stderr,
            )
        except Exception:
            pass
        return


def emit_deny(reason: str) -> None:
    """Print the deny decision on stdout and flush it.

    Deliberately does not exit. Callers emit the decision first, then record
    the signal, then exit — so no signal-path failure or stall can suppress
    or delay a denial that has already been delivered.
    """
    print(
        json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": reason,
                }
            }
        )
    )
    sys.stdout.flush()


def main() -> None:
    """Gate one Edit/Write/MultiEdit call against the active phase marker.

    Malformed or empty stdin fails open by design: the hook exits 0 (allow)
    rather than denying, because a parser hiccup would otherwise brick every
    edit in the session. Do not harden that path to fail closed.
    """
    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError:
        sys.exit(0)

    tool_name = payload.get("tool_name", "")
    if tool_name not in ALLOWED_TOOLS:
        sys.exit(0)

    tool_input = payload.get("tool_input", {}) or {}
    target_path = tool_input.get("file_path", "") or ""

    cwd = payload.get("cwd") or os.getcwd()
    project_root = find_project_root(cwd)
    if not project_root:
        sys.exit(0)

    if is_allowlisted(target_path, project_root):
        sys.exit(0)

    active_path = os.path.join(project_root, ".m", "DEVELOP_ACTIVE")
    current_phase = parse_current_phase(active_path)
    if not current_phase:
        emit_deny(
            "`.m/DEVELOP_ACTIVE` is present but `current_phase:` is missing "
            "or unreadable. Either fix the marker or delete it if /m:develop "
            "is not actually running."
        )
        record_denial(
            project_root, None, tool_name, target_path, "unreadable_marker"
        )
        sys.exit(0)

    started_marker = os.path.join(
        project_root, ".m", f"phase-{current_phase}-started"
    )
    if os.path.isfile(started_marker):
        sys.exit(0)

    emit_deny(
        "/m:develop pipeline is active and the current phase "
        f"({current_phase!r}) has not been entered via its skill.\n\n"
        f"Before any Edit/Write/MultiEdit outside `.m/`, you must invoke "
        f"the corresponding skill (for example: Skill(skill=\"m:{current_phase}\")). "
        f"The skill is required to touch `.m/phase-{current_phase}-started` "
        f"on entry. That marker is missing, so this tool call is blocked.\n\n"
        f"If /m:develop is not actually running, remove "
        f"`.m/DEVELOP_ACTIVE` to disable this gate."
    )
    record_denial(
        project_root, current_phase, tool_name, target_path,
        "missing_phase_marker",
    )
    sys.exit(0)


if __name__ == "__main__":
    main()
