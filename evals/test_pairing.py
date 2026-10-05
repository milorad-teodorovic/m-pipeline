"""Tests for the plain arm and the paired pipeline comparison."""

import io
import json
from contextlib import redirect_stdout
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest

from checks import InvalidRun
from pairing import paired, report_arms, strip_plugin
from run import ROOT, copy_plugin


def report(trials, fixtures, **metadata):
    base = {"model": "pinned", "judge_model": "j", "runs": 3, "checker_hash": "c", "oracle_hash": "o", "fixture_hashes": fixtures}
    return {"complete": True, "metadata": {**base, **metadata}, "trials": trials}


def trial(case, arm, status, cost=1.0):
    return {"case": case, "arm": arm, "status": status, "cost_usd": cost, "duration_seconds": 10}


SUITE = {"cases": {"develop--x": {"pair": "develop--neutral-x"}, "develop--neutral-x": {}}}


class PairingTest(unittest.TestCase):
    def write(self, directory, name, value):
        path = Path(directory) / name
        path.write_text(json.dumps(value))
        return path

    def test_report_arms(self):
        self.assertEqual(report_arms(SimpleNamespace(plain=True, ablation="none")), (("with", "without"),))
        self.assertEqual(report_arms(SimpleNamespace(plain=False, ablation="with-without")), (("with", "with"), ("without", "without")))
        self.assertEqual(report_arms(SimpleNamespace(plain=False, ablation="none")), (("with", "with"),))

    def test_strip_plugin_removes_every_component(self):
        with tempfile.TemporaryDirectory() as temp:
            snapshot = Path(temp) / "plugin"
            copy_plugin(ROOT, snapshot)
            strip_plugin(snapshot)
            for part in ("commands", "skills", "hooks", "references", "rules"):
                self.assertFalse((snapshot / part).exists(), part)
            manifest = json.loads((snapshot / ".claude-plugin/plugin.json").read_text())
            self.assertFalse({"commands", "skills", "hooks"} & manifest.keys())
            self.assertTrue((snapshot / "evals/suite.json").is_file())
            self.assertFalse((snapshot / "evals/pairing.py").exists())

    def test_paired_reports_delta(self):
        with tempfile.TemporaryDirectory() as temp:
            left = self.write(temp, "a.json", report([trial("develop--x", "with", s) for s in ("PASS", "PASS", "FAIL")], {"develop--x": "h"}))
            right = self.write(temp, "b.json", report([trial("develop--neutral-x", "without", s, 0.5) for s in ("PASS", "FAIL", "FAIL")] + [trial("develop--neutral-x", "with", "PASS")], {"develop--neutral-x": "h"}))
            out = io.StringIO()
            with redirect_stdout(out):
                self.assertEqual(paired(left, right, SUITE), 0)
            result = json.loads(out.getvalue())
            self.assertAlmostEqual(result["delta_pass_rate"], 1 / 3)
            self.assertEqual(result["summary"]["plain"]["trials"], 3)
            self.assertEqual(result["summary"]["plain"]["mean_cost_usd"], 0.5)

    def test_paired_rejects_mismatched_inputs(self):
        with tempfile.TemporaryDirectory() as temp:
            left = self.write(temp, "a.json", report([trial("develop--x", "with", "PASS")], {"develop--x": "h"}))
            other = self.write(temp, "b.json", report([trial("develop--neutral-x", "without", "PASS")], {"develop--neutral-x": "different"}))
            with self.assertRaises(InvalidRun):
                paired(left, other, SUITE)
            model = self.write(temp, "c.json", report([trial("develop--neutral-x", "without", "PASS")], {"develop--neutral-x": "h"}, model="other"))
            with self.assertRaises(InvalidRun):
                paired(left, model, SUITE)
            unpaired = self.write(temp, "d.json", report([trial("develop--neutral-x", "with", "PASS")], {"develop--neutral-x": "h"}))
            with self.assertRaises(InvalidRun):
                paired(unpaired, left, SUITE)


if __name__ == "__main__":
    unittest.main()
