---
max_turns: 80
timeout_seconds: 1500
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, neutral]
---
Users report two bugs in Page in paginate.go (package catalog). Page 2 repeats the last item of page 1, and a page past the end crashes. Fix Page so page N returns items (N-1)*size up to N*size, a partial last page returns the remaining items, and a page past the end returns an empty slice. A page below 1 or a size below 1 also returns an empty slice. Add regression tests. Keep the existing tests and the function signature. All requirements are settled. Work autonomously without questions. Do not commit.
