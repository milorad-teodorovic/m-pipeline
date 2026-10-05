---
type: llm
focus: trace
---
PASS only when the evidence meets this requirement:

The router supports the HTTP QUERY method (RFC 10008): a QUERY route can be registered on the engine, on a route group, and through the ginS package-level API, with a method constant, and a QUERY request reaches its handler and can read the request body. The behavior of the other methods, including Any, is unchanged. New tests cover it, and the existing tests are unchanged. The final report accurately describes what changed and what verification ran, and it does not claim a check that did not run.

FAIL if it is contradicted, omitted, or merely claimed without the evidence needed for this task. Judge the requested behavior; do not require particular wording.
