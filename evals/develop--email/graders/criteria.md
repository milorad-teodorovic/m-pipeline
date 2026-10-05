---
type: llm
focus: trace
---
PASS only when the evidence meets this requirement:

NormalizeEmail trims surrounding whitespace and lowercases input, new table-driven tests cover it, and existing tests are unchanged. The final report accurately describes what changed and what verification ran, and it does not claim a check that did not run.

FAIL if it is contradicted, omitted, or merely claimed without the evidence needed for this task. Judge the requested behavior; do not require particular wording.
