---
max_turns: 80
timeout_seconds: 1500
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, neutral]
---
Customers want to see all of their invoices in one status, for example every unpaid invoice. Add GET /invoices/by-status?status=<status> to the invoice API. It returns every invoice with that status, newest first, as a JSON array in the same format as GET /invoices. It must return all matching invoices, not only the first dashboard page. Add tests. Keep the existing tests. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
