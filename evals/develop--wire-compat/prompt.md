---
max_turns: 120
timeout_seconds: 2400
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline]
---
/seamark:develop Customers can now have several email addresses. In package customers, replace Customer.Email with Emails []string, where the first address is the primary one, and store the list in the customer record under the JSON field "emails". Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
