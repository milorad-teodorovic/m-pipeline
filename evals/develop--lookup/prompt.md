---
max_turns: 120
timeout_seconds: 2400
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline]
---
/seamark:develop Store.Find in package users returns a zero User for an unknown ID, and callers treat that as a real user. Change Find to return (User, error), with an exported ErrNotFound for an unknown ID, and update every caller. No caller may treat an unknown ID as a user, and the profile endpoint must answer 404 Not Found for an unknown ID. Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
