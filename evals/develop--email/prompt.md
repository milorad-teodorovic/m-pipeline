---
max_turns: 120
timeout_seconds: 2400
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline]
---
/seamark:develop In package email, NormalizeEmail in email.go returns its input unchanged. Make it trim surrounding whitespace, including spaces, tabs, and newlines, and convert the address to lower case. Add table-driven tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
