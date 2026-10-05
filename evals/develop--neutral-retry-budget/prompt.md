---
max_turns: 80
timeout_seconds: 1500
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, neutral]
---
Calls to the partner API fail sometimes, and our jobs give up on the first failure. Add retries to Client.Do in package partner so that transient partner failures no longer break our jobs. Use the Sleep and Now fields that Client already has for waiting and for the current time. Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
