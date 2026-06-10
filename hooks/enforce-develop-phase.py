#!/usr/bin/env python3
"""
PreToolUse hook for Edit, Write, and MultiEdit.

Purpose: enforce the /m:develop phase protocol. When /m:develop is active
(marker file `.m/DEVELOP_ACTIVE` present in the project root), this hook
blocks code mutation unless the current required phase has been entered
via its corresponding /m:* skill (marker file `.m/phase-<name>-started`
present).

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
"""

import json
import os
import sys


ALLOWED_TOOLS = {"Edit", "Write", "MultiEdit"}


def find_project_root(start_dir: str):
    """Walk up from start_dir looking for `.m/DEVELOP_ACTIVE`.

    Returns the directory that contains `.m/DEVELOP_ACTIVE`, or None if
    no such directory is found within 10 levels.
    """
    current = os.path.abspath(start_dir)
    for _ in range(10):
        candidate = os.path.join(current, ".m", "DEVELOP_ACTIVE")
        if os.path.isfile(candidate):
            return current
        parent = os.path.dirname(current)
        if parent == current:
            return None
        current = parent
    return None


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
    """True when the write target is within the pipeline bookkeeping area."""
    if not target_path:
        return True
    abs_target = os.path.abspath(target_path)
    m_dir = os.path.abspath(os.path.join(project_root, ".m"))
    return abs_target == m_dir or abs_target.startswith(m_dir + os.sep)


def deny(reason: str) -> None:
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
    sys.exit(0)


def main() -> None:
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
        deny(
            "`.m/DEVELOP_ACTIVE` is present but `current_phase:` is missing "
            "or unreadable. Either fix the marker or delete it if /m:develop "
            "is not actually running."
        )

    started_marker = os.path.join(
        project_root, ".m", f"phase-{current_phase}-started"
    )
    if os.path.isfile(started_marker):
        sys.exit(0)

    deny(
        "/m:develop pipeline is active and the current phase "
        f"({current_phase!r}) has not been entered via its skill.\n\n"
        f"Before any Edit/Write/MultiEdit outside `.m/`, you must invoke "
        f"the corresponding skill (for example: Skill(skill=\"m:{current_phase}\")). "
        f"The skill is required to touch `.m/phase-{current_phase}-started` "
        f"on entry. That marker is missing, so this tool call is blocked.\n\n"
        f"If /m:develop is not actually running, remove "
        f"`.m/DEVELOP_ACTIVE` to disable this gate."
    )


if __name__ == "__main__":
    main()
