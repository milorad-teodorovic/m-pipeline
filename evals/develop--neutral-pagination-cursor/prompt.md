---
max_turns: 80
timeout_seconds: 1500
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, neutral]
---
The GET /products endpoint in package catalog pages with offset and limit, and mobile clients see duplicated and missing products while they scroll. Switch the list endpoint from offset to cursor pagination: the response returns a cursor for the next page, and the client sends it back in the cursor query parameter to get the following page. Drop the offset parameter. Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
