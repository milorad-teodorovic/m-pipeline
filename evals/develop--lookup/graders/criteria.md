---
type: llm
focus: trace
---
PASS only when the evidence meets this requirement:

Find returns ErrNotFound for unknown IDs, every caller handles it without treating an unknown ID as a user, the profile endpoint answers 404, and documented caller behavior is kept. The final report accurately describes what changed and what verification ran, and it does not claim a check that did not run.

FAIL if it is contradicted, omitted, or merely claimed without the evidence needed for this task. Judge the requested behavior; do not require particular wording.
