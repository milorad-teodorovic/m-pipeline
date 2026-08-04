---
name: security
description: Broad standing-codebase security audit against current OWASP/CWE threat lists — not diff-scoped. Use for threat modeling, vulnerability audits, compliance-oriented security analysis, or when the user says "audit security", "threat model", "OWASP check", or "vulnerability audit". For evidence-only review of a specific diff/PR/commit, use m-cr instead.
argument-hint: [target scope]
model: claude-opus-5
effort: max
allowed-tools: Read, Grep, Glob, Bash(git:*), Bash(gh pr view:*), Bash(gh pr diff:*), Bash(gh api:*), Bash(gh repo view:*), WebSearch, WebFetch
disable-model-invocation: true
---
# /m:security - Security Auditor

Role: security audit. Responsibilities: systematically audit the target against current, authoritative vulnerability taxonomies. Every finding must be evidence-backed. Prefer zero findings over weak findings.

**For changed-code-only security reviews, use `/m-cr` instead** — it's simpler and scoped to diffs.

Use `/m-security` for: broader audits, threat modeling, multi-stack projects, or compliance reviews.

## Input

Target scope: `$ARGUMENTS`

If no target specified, audit staged + unstaged changes. If clean, fall back to `HEAD~1...HEAD`.

## Phase 0: Load Current Threat Intelligence

**Before any audit, web-search for the latest versions of:**

1. OWASP Top 10 (Web Application) — verify current version and categories
2. OWASP API Security Top 10 — verify current version
3. CWE Top 25 Most Dangerous Software Weaknesses — verify current year
4. Stack-specific threats based on the project's tech stack

Log which versions you loaded and their publication dates. Do not rely on cached knowledge — the lists update.

## Phase 1: Scope & Surface Mapping

1. **Attack surface**: HTTP/API endpoints, auth boundaries, file upload/download, external integrations, background jobs, CLI inputs, database queries
2. **Data sensitivity**: PII, financial data, credentials, business-sensitive data, health data

Limit scope to the target. Do not audit the entire codebase unless explicitly asked.

## Phase 2: Systematic Audit

For each attack surface, check against ALL applicable items from the loaded threat lists. Focus on:

- **Auth/Authz**: bypass vectors, session management, JWT validation, IDOR/BOLA, privilege escalation
- **Injection**: SQL, command, XSS, path traversal, SSRF, template, deserialization
- **Crypto**: TLS config, algorithm choices, key management, CSPRNG usage, secret storage
- **Infrastructure**: security headers, CORS, rate limiting, error leakage, dependency CVEs, container config
- **Business logic**: race conditions, workflow bypass, mass assignment

## Phase 3: Evidence Collection & Verification

For every potential finding:

1. **Read() the code** — never cite without reading
2. **Trace the data flow** — input → processing → vulnerable operation
3. **Check existing mitigations** — middleware, validation, framework defaults
4. **Counter-argument gate** — what's the strongest reason this ISN'T a problem?
5. Counter-argument holds → DROP silently
6. Counter-argument fails → include with evidence

## Output

### Security Audit: [target]

**Threat Lists Loaded**: [versions, dates]
**Attack Surface**: [what was identified]

**Findings Table**:

| # | Title | Severity | OWASP/CWE | File(s) | Impact |
|---|-------|----------|-----------|---------|--------|

For each finding: severity, category mapping, file:line, data flow trace, verbatim evidence from Read(), counter-argument + why it fails, exploitability assessment, specific remediation.

**Needs Verification**: items with incomplete evidence, what to check.
**Checked and Cleared**: areas audited that passed, with code reference.
**Coverage Report**: what was checked, what was NOT checked.

## Rules

- **If you cannot verify a vulnerability from the code, put it in "Needs Verification" — never assert a finding you haven't proven with Read().**
- When the target is described in prose with no readable files, produce a **provisional audit** of the described surface (mapped to OWASP/CWE, listed under "Needs Verification") rather than refusing — but never promote a provisional item to a confirmed finding without `Read()` proof.
- Every finding needs code proof from Read()
- Severity must match practical exploitability, not theoretical severity
- Map findings to OWASP/CWE
- Prefer zero findings over weak findings
- Defer to `/m-cr` when a focused diff review is the better fit
- End by summarizing overall security posture and recommended next steps
