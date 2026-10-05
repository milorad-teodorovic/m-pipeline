"""Deterministic checks over complete Claude traces and resulting workspaces."""

from dataclasses import dataclass
import fnmatch
import hashlib
import json
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import tempfile


class InvalidRun(ValueError):
    """The harness cannot establish an outcome from the available evidence."""


@dataclass
class Call:
    name: str
    input: dict
    position: int
    response: dict | None = None
    response_position: int = -1

    @property
    def succeeded(self):
        return self.response is not None and not self.response.get("is_error", False)


class Trace:
    def __init__(self, events):
        self.events = events
        initial = [e for e in events if e.get("type") == "system" and e.get("subtype") == "init"]
        terminal = [e for e in events if e.get("type") == "result"]
        if not initial or not terminal:
            raise InvalidRun("trace needs initialization and a terminal result")
        self.cwd = Path(initial[0]["cwd"])
        self.plugin_errors = initial[0].get("plugin_errors", [])
        self.result = terminal[-1]
        if self.result.get("is_error") or str(self.result.get("subtype", "")).startswith("error"):
            raise InvalidRun("agent did not complete: " + str(self.result.get("subtype")))
        self.final = str(self.result.get("result", ""))
        if re.match(r"\s*Unknown command:", self.final):
            raise InvalidRun("requested command is unavailable")
        by_id = {}
        responses = {}
        self.calls = []
        self.permission_denials = []
        for position, event in enumerate(events):
            message = event.get("message", {})
            content = message.get("content", []) if isinstance(message, dict) else []
            if not isinstance(content, list):
                continue
            for block in content:
                if not isinstance(block, dict):
                    continue
                if event.get("type") == "assistant" and block.get("type") == "tool_use":
                    if block["id"] not in by_id:
                        call = Call(block["name"], block.get("input", {}), position)
                        by_id[block["id"]] = call
                        self.calls.append(call)
                elif event.get("type") == "user" and block.get("type") == "tool_result":
                    responses[block["tool_use_id"]] = (block, position)
        for ident, call in by_id.items():
            if ident not in responses:
                raise InvalidRun(f"missing tool result: {ident}")
            call.response, call.response_position = responses[ident]
            output = str(call.response.get("content", ""))
            denied = "Permission to use" in output and "denied" in output
            if denied and ".claude/seamark-learning/" not in json.dumps(call.input):
                self.permission_denials.append(call)

    @classmethod
    def read(cls, path):
        try:
            events = [json.loads(line) for line in Path(path).read_text().splitlines() if line.strip()]
            if any(not isinstance(event, dict) for event in events):
                raise ValueError("event is not an object")
            return cls(events)
        except (OSError, KeyError, json.JSONDecodeError, ValueError) as error:
            if isinstance(error, InvalidRun):
                raise
            raise InvalidRun(f"unreadable or incomplete trace: {error}") from error


def file_hashes(root):
    """Hash content independently of the Git index the agent can change."""
    result = {}
    for path in sorted(Path(root).rglob("*")):
        relative = path.relative_to(root)
        if any(part in {".git", "__pycache__"} for part in relative.parts):
            continue
        if path.is_symlink():
            result[str(relative)] = "symlink:" + str(path.readlink())
        elif path.is_file():
            result[str(relative)] = hashlib.sha256(path.read_bytes()).hexdigest()
    return result


def commands(command):
    """Extract direct simple commands, preserving quoted strings as one token."""
    lexer = shlex.shlex(command, posix=True, punctuation_chars=";&|()<>\n")
    lexer.whitespace = " \t\r"
    lexer.whitespace_split = True
    try:
        tokens = list(lexer)
    except ValueError:
        return []
    segments, current = [], []
    for token in tokens + [";"]:
        if token and all(char in ";&|()\n" for char in token):
            current = strip_prefixes(current)
            if current:
                segments.append(current)
                current = []
        else:
            current.append(token)
    return segments


SHELL_KEYWORDS = {"do", "then", "else", "elif", "!", "{", "time"}
ASSIGNMENT = re.compile(r"[A-Za-z_][A-Za-z0-9_]*=")


def strip_prefixes(args):
    """Drop leading shell keywords and NAME=value assignments before the command word."""
    start = 0
    while start < len(args) and (args[start] in SHELL_KEYWORDS or ASSIGNMENT.match(args[start])):
        start += 1
    return args[start:]


def runs_command(call, wanted, require_success=False):
    if call.name != "Bash" or call.response is None or (require_success and not call.succeeded):
        return False
    return any(
        len(args) >= len(wanted)
        and Path(args[0]).name == wanted[0]
        and args[1:len(wanted)] == wanted[1:]
        for args in commands(call.input.get("command", ""))
    )


def touches(call, path):
    if not call.succeeded:
        return False
    if call.name == "Write":
        return call.input.get("file_path", "").endswith("/" + path) or call.input.get("file_path") == path
    if call.name == "Bash":
        return any(args[0] == "touch" and path in args[1:] for args in commands(call.input.get("command", "")))
    return False


def verdict(text):
    # Select the final explicit verdict, ignoring mentions inside quoted code.
    text = re.sub(r"```[\s\S]*?```", "", text)
    text = re.sub(r"`([^`\n]*)`", r"\1", text)
    explicit = re.findall(r"(?im)^#{1,6}\s+Verdict\s*\n\s*(?:\*\*)?(APPROVED WITH WARNINGS|APPROVED|BLOCKED|PASSED|N/A)\b", text)
    if explicit:
        return explicit[-1].upper()
    # A pipeline outcome takes precedence over a nested review verdict listed
    # among the completed checks in the same delivery report.
    sections = re.findall(r"(?ims)^### Pipeline Stages Run\s*\n(.*?)(?=^### |\Z)", text)
    pipeline = [outcome for section in sections for outcome in re.findall(r"(?i)\bverify\s*\([^\n)]*\b(PASSED|BLOCKED)\b", section)]
    if pipeline:
        return pipeline[-1].upper()
    matches = re.findall(r"(?i)\b(?:Verdict|Status)(?:\*\*)?\s*:?\s*(?:\*\*)?\s*\n?\s*(?:\*\*)?(APPROVED WITH WARNINGS|APPROVED|BLOCKED|PASSED|N/A)\b", text)
    if not matches:
        matches = re.findall(r"(?im)^\s*(?:\*\*)?(APPROVED WITH WARNINGS|APPROVED|BLOCKED|PASSED|N/A)(?:\*\*)?(?:\s|$|[.:])", text)
    return matches[-1].upper() if matches else None


def safe_path(root, relative):
    path = Path(root) / relative
    if not path.resolve().is_relative_to(Path(root).resolve()):
        raise InvalidRun(f"artifact escapes workspace: {relative}")
    return path


def denial_blocks_contract(call, trace, case):
    """Required fixture access differs from an optional out-of-scope search."""
    if call.name == "Bash":
        return any(check["kind"] == "command" and runs_command(call, check["argv"]) for check in case.get("checks", []))
    target = call.input.get("file_path") or call.input.get("path")
    if not target or call.name not in {"Read", "Write", "Edit", "Grep", "Glob"}:
        return False
    path = Path(target)
    path = path if path.is_absolute() else trace.cwd / path
    try:
        relative = str(path.resolve().relative_to(trace.cwd.resolve()))
    except ValueError:
        return False
    required = set(case.get("required_files", []))
    for check in case.get("checks", []):
        if check["kind"] in {"changed", "exists"}:
            required.update(check["paths"])
    return relative in required


def phase_order(trace, phases):
    indices = {f"seamark:{name}": index for index, name in enumerate(phases)}
    if "review" in phases:
        indices["seamark:review-fanout"] = phases.index("review")
    current, previous = -1, -1
    for entry in trace.calls:
        name = entry.input.get("skill")
        if entry.name != "Skill" or not entry.succeeded or name not in indices:
            continue
        index = indices[name]
        if index not in {current, current + 1}:
            return False
        if index != current and current >= 0:
            marker = f".seamark/phase-{phases[current]}-done"
            if not any(previous < c.position and c.response_position < entry.position and touches(c, marker) for c in trace.calls):
                return False
        current, previous = index, entry.position
    return current == len(phases) - 1


def oracle(root, source, directory=".", timeout=90):
    """Run the hidden oracle tests against a copy of root.

    directory is the package directory, relative to root, that receives the
    oracle file; timeout bounds the full go test run in seconds.
    """
    if not shutil.which("go"):
        raise InvalidRun("Go is unavailable for independent verification")
    with tempfile.TemporaryDirectory(prefix="m-eval-oracle-") as scratch:
        workspace = Path(scratch) / "workspace"
        shutil.copytree(root, workspace, symlinks=True, ignore=shutil.ignore_patterns(".git", ".checks"))
        for path in workspace.rglob("*"):
            if path.is_symlink():
                raise InvalidRun("oracle workspace contains a symlink")
        target = safe_path(workspace, directory)
        if not target.is_dir():
            raise InvalidRun(f"oracle package directory is missing: {directory}")
        shutil.copyfile(source, target / "zz_eval_oracle_test.go")
        try:
            completed = subprocess.run(["go", "test", "-json", "./..."], cwd=workspace, capture_output=True, text=True, timeout=timeout)
        except subprocess.TimeoutExpired as error:
            raise InvalidRun("independent tests timed out") from error
        expected = set(re.findall(r"(?m)^func (Test\w+)\(", Path(source).read_text()))
        if not expected:
            raise InvalidRun("oracle source declares no test functions")
        try:
            events = [json.loads(line) for line in completed.stdout.splitlines() if line.strip()]
        except ValueError as error:
            raise InvalidRun("Go did not emit a valid test event stream") from error
        passed = {event.get("Test") for event in events if event.get("Action") == "pass"}
        failures = any(event.get("Action") == "fail" for event in events)
        success = completed.returncode == 0 and not failures and expected <= passed
        evidence = (completed.stdout + completed.stderr)[-8000:]
        if expected - passed:
            evidence += "\nOracle tests did not pass: " + ", ".join(sorted(expected - passed))
        return success, evidence


KINDS = {"skill", "no_skill", "exists", "absent", "contains", "glob_contains", "unchanged", "changes", "changed", "verdict", "command", "history", "phases", "oracle"}


def evaluate(check, trace, workspace, initial, oracle_root):
    kind = check["kind"]
    paths = check.get("paths", [])
    final = file_hashes(workspace)
    changed = {p for p in initial.keys() | final.keys() if initial.get(p) != final.get(p)}
    if kind == "skill":
        passed = any(c.name == "Skill" and c.succeeded and c.input.get("skill") == check["name"] for c in trace.calls)
    elif kind == "no_skill":
        passed = not any(c.name == "Skill" and c.input.get("skill", "").startswith(check["prefix"]) for c in trace.calls)
    elif kind == "exists":
        passed = all(safe_path(workspace, p).is_file() for p in paths)
    elif kind == "absent":
        passed = all(not list(Path(workspace).glob(p)) for p in paths)
    elif kind == "contains":
        path = safe_path(workspace, check["path"])
        passed = path.is_file() and check["text"] in path.read_text()
    elif kind == "glob_contains":
        matches = list(Path(workspace).glob(check["glob"]))
        passed = bool(matches) and all(check["text"] in safe_path(workspace, str(p.relative_to(workspace))).read_text() for p in matches)
    elif kind == "unchanged":
        passed = all(initial.get(p) == final.get(p) for p in initial if any(fnmatch.fnmatchcase(p, pattern) for pattern in paths))
    elif kind == "changes":
        passed = all(any(fnmatch.fnmatchcase(p, pattern) for pattern in paths) for p in changed)
    elif kind == "changed":
        passed = all(p in changed for p in paths)
    elif kind == "verdict":
        passed = verdict(trace.final) in check["values"]
    elif kind == "command":
        count = sum(runs_command(c, check["argv"], check.get("success", False)) for c in trace.calls)
        # The controlled verification fixture records individual executions;
        # a single Bash call can run it repeatedly inside a shell loop.
        if count and check["argv"] == ["python3", "check.py"] and check.get("min", 1) > 1:
            path = safe_path(workspace, ".checks/runs.jsonl")
            intact = all(p in initial and initial[p] == final.get(p) for p in ("check.py", "verification.json"))
            if intact and path.is_file():
                records = [json.loads(line) for line in path.read_text().splitlines()]
                if all(r.get("run") == i + 1 and r.get("exit_code") in (0, 1) for i, r in enumerate(records)):
                    count = sum(not check.get("success", False) or r["exit_code"] == 0 for r in records)
        passed = count >= check.get("min", 1)
    elif kind == "history":
        path = safe_path(workspace, ".checks/runs.jsonl")
        records = [json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []
        passed = len(records) >= check.get("min", 1) and all(any(r.get("exit_code") == code for r in records) for code in check.get("exit_codes", []))
    elif kind == "phases":
        passed = phase_order(trace, check["names"])
    elif kind == "oracle":
        return oracle(workspace, Path(oracle_root) / check["file"], check.get("dir", "."), check.get("timeout", 90))
    else:
        raise InvalidRun(f"unknown deterministic check: {kind}")
    return passed, "satisfied" if passed else json.dumps(check, sort_keys=True)


def grade_trial(raw, case, trace_path, workspace, initial, oracle_root, arm="with"):
    report = {
        "status": "INVALID", "checks": [], "semantic_score": None,
        "cost_usd": raw.get("costUsd", 0), "judge_cost_usd": raw.get("judgeCostUsd", 0),
        "duration_seconds": raw.get("durationSeconds"), "turns": raw.get("turns"),
    }
    try:
        if raw.get("error") or raw.get("skippedPaidGraders"):
            raise InvalidRun(str(raw.get("error") or "paid graders were skipped"))
        trace = Trace.read(trace_path)
        report["plugin_errors"] = trace.plugin_errors
        report["permission_denials"] = [{"tool": call.name, "input": call.input} for call in trace.permission_denials]
        blocked = [call for call in trace.permission_denials if denial_blocks_contract(call, trace, case)]
        if blocked:
            raise InvalidRun(f"permission denial prevents required fixture access or verification: {blocked[0].name}")
        if not Path(workspace).is_dir():
            raise InvalidRun("final workspace is missing")
        # Trace presence is mandatory even for purely conversational checks.
        for check in case.get("checks", []):
            if check.get("arm") == "with" and arm == "without":
                continue
            passed, evidence = evaluate(check, trace, workspace, initial, oracle_root)
            report["checks"].append({"id": check["id"], "passed": passed, "evidence": evidence})
        diagnostics = {"plugin-fired", "help-skill-fired", "no-seamark-skill-fired"}
        semantic = [g for g in raw.get("graders", []) if g["name"] not in diagnostics]
        expected = set(case["graders"]) - diagnostics
        if {g["name"] for g in semantic} != expected or not semantic:
            raise InvalidRun("missing or unexpected native graders")
        report["semantic_score"] = sum(bool(g["passed"]) for g in semantic) / len(semantic)
        report["semantic_evidence_truncated"] = any("messages elided" in g.get("evidence", "") for g in semantic)
        report["graders"] = [{k: g[k] for k in ("name", "passed", "explanation", "judgeVotes") if k in g} for g in semantic]
        report["judge_disagreements"] = sum(len(set(g.get("judgeVotes", []))) > 1 for g in semantic)
        report["activation"] = [c.input.get("skill") for c in trace.calls if c.name == "Skill" and c.succeeded]
        passed = all(c["passed"] for c in report["checks"]) and all(g["passed"] for g in semantic)
        report["status"] = "PASS" if passed else "FAIL"
    except (InvalidRun, OSError, ValueError, KeyError) as error:
        report["error"] = str(error)
    return report
