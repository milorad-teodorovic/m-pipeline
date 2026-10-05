---
max_turns: 120
timeout_seconds: 2400
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline]
---
/seamark:develop Customers want their daily job to run at the scheduled wall-clock time in their own time zone, not in UTC. Add a Location field of type *time.Location to Daily in package schedule and make NextRun use it. A nil Location keeps the current UTC behavior. Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
