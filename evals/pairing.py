"""Run a plain no-plugin arm and pair /seamark:develop cases with their plain twins."""

import json
from pathlib import Path
import shutil

from checks import InvalidRun
from selection import noise_floor


PLAIN_MANIFEST = {
    "name": "seamark",
    "description": "Empty plugin for the plain arm: no commands, skills, hooks, references, or rules.",
    "version": "0.0.0",
}


def strip_plugin(snapshot):
    """Remove every plugin component from snapshot and keep only its eval suite.

    The native runner needs a plugin target, so the plain arm loads this empty
    plugin. Claude then runs with the built-in tools and prompts only.
    """
    for part in ("commands", "skills", "hooks", "references", "rules"):
        shutil.rmtree(snapshot / part, ignore_errors=True)
    (snapshot / ".claude-plugin/plugin.json").write_text(json.dumps(PLAIN_MANIFEST, indent=2) + "\n")


def report_arms(args):
    """Return (native arm, report arm) pairs for one native run.

    A plain run loads the empty plugin in the native with-arm and reports it as
    the without-arm, so it lines up with the without-arm of an ablation run.
    """
    if args.plain:
        return (("with", "without"),)
    if args.ablation == "with-without":
        return (("with", "with"), ("without", "without"))
    return (("with", "with"),)


def rates(trials):
    """Return pass count, valid count, mean oracle coverage, and mean cost and duration of trials."""
    valid = [t for t in trials if t["status"] in ("PASS", "FAIL")]
    cost = sum(t.get("cost_usd", 0) for t in valid)
    seconds = sum(t.get("duration_seconds") or 0 for t in valid)
    shares = [t["oracle_coverage"] for t in valid if t.get("oracle_coverage") is not None]
    return {
        "passed": sum(t["status"] == "PASS" for t in valid), "trials": len(valid),
        "mean_coverage": round(sum(shares) / len(shares), 3) if shares else None,
        "invalid": len(trials) - len(valid),
        "mean_cost_usd": round(cost / len(valid), 3) if valid else None,
        "mean_seconds": round(seconds / len(valid)) if valid else None,
    }


def paired(pipeline_path, plain_path, suite):
    """Compare each paired case's pipeline arm with its plain twin's no-plugin arm.

    pipeline_path is a report whose cases carry a "pair" key in suite.json and
    ran with the plugin. plain_path is a report that holds the no-plugin arm of
    each pair, from a --plain run or the without-arm of an ablation run.
    Returns 0 when every pair has trials in both reports, else 2.
    """
    pipeline, plain = (json.loads(Path(p).read_text()) for p in (pipeline_path, plain_path))
    if not pipeline.get("complete") or not plain.get("complete"):
        raise InvalidRun("paired comparison requires complete reports")
    for key in ("model", "judge_model", "runs", "checker_hash", "oracle_hash"):
        if pipeline["metadata"].get(key) != plain["metadata"].get(key):
            raise InvalidRun(f"paired reports differ in {key}")
    rows, totals = [], {"pipeline": [], "plain": []}
    for name in sorted({t["case"] for t in pipeline["trials"]}):
        twin = suite["cases"].get(name, {}).get("pair")
        if not twin:
            raise InvalidRun(f"{name} has no pair in suite.json")
        if pipeline["metadata"]["fixture_hashes"].get(name) != plain["metadata"]["fixture_hashes"].get(twin):
            raise InvalidRun(f"{name} and {twin} start from different fixtures")
        left = [t for t in pipeline["trials"] if t["case"] == name and t["arm"] == "with"]
        right = [t for t in plain["trials"] if t["case"] == twin and t["arm"] == "without"]
        if not left or not right:
            raise InvalidRun(f"{name}: a side of the pair has no trials")
        totals["pipeline"] += left
        totals["plain"] += right
        rows.append({"case": name, "pair": twin, "pipeline": rates(left), "plain": rates(right)})
    summary = {side: rates(trials) for side, trials in totals.items()}
    n = min(summary["pipeline"]["trials"], summary["plain"]["trials"])
    delta = (summary["pipeline"]["passed"] / summary["pipeline"]["trials"] - summary["plain"]["passed"] / summary["plain"]["trials"]) if n else None
    print(json.dumps({"pairs": rows, "summary": summary, "delta_pass_rate": delta, "noise_floor": noise_floor(n)}, indent=2))
    return 0 if all(r["pipeline"]["invalid"] == 0 and r["plain"]["invalid"] == 0 for r in rows) else 2
