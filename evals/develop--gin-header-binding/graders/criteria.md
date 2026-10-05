---
type: llm
focus: trace
---
PASS only when the evidence meets this requirement:

Context gains BindHeader and ShouldBindHeader, backed by a header binding in package binding that maps request headers to struct fields through `header` tags, following the pattern of the existing bindings. New tests cover it, and the existing tests are unchanged. The final report accurately describes what changed and what verification ran, and it does not claim a check that did not run.

FAIL if it is contradicted, omitted, or merely claimed without the evidence needed for this task. Judge the requested behavior; do not require particular wording.
