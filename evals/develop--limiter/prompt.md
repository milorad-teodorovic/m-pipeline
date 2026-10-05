---
max_turns: 120
timeout_seconds: 2400
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline]
---
/seamark:develop Implement Allow in limiter.go (package ratelimit). The Limiter is a fixed-window rate limiter: each key may make at most limit allowed requests per window, windows are tracked per key, a new window starts once window has elapsed since the key's current window began, and time comes only from the injected now function. Allow must be safe for concurrent use. Keep NewLimiter's signature and the existing tests. Add tests, including a concurrency test. Use the standard library only. All requirements are settled. Work autonomously without questions. Do not commit.
