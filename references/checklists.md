# Review Checklists Reference (73 checks)

## Clean Code (CC-1 → CC-9)
1. SRP — single responsibility per function/class
2. OCP — open for extension, closed for modification
3. DIP — depend on abstractions, not concretions
4. Cognitive complexity — no deeply nested logic
5. Naming — descriptive, consistent with codebase conventions
6. Magic values — use named constants
7. Dead code — remove unused functions, variables, imports
8. DRY — no duplicated logic
9. KISS/YAGNI — simplest solution, no speculative features

## Go (GO-1 → GO-9)
1. Error handling — all errors checked, wrapped with context
2. Context propagation — `ctx context.Context` through all chains
3. Goroutine safety — no data races, proper synchronization
4. Defer — correct usage for cleanup
5. Interfaces — small, consumer-defined
6. Constructors — `NewXxx(deps, ...Option)` with functional options
7. RFC 7807 — error responses at handler boundary
8. Logging — structured (slog/zerolog), PII masking
9. Test patterns — table-driven, testify, testutil helpers

## TypeScript/React (TS-1 → TS-8)
1. Strict TypeScript — no `any`, proper generics
2. Hook rules — correct dependencies, no conditional hooks
3. Key props — stable keys in lists
4. Error boundaries — present for critical UI sections
5. Shared packages — use the project's existing shared libraries before creating new
6. Zod validation — at API boundaries
7. React Query — for server state (not local state)
8. Component patterns — functional, typed Props interface

## Python (PY-1 → PY-5)
1. Async correctness — all async functions awaited, no blocking in async
2. Pydantic — request/response models with field validators
3. Type hints — full signatures, `|` union syntax
4. Exceptions — specific types, no bare `except:`
5. Logging — `logging.getLogger(__name__)`

## Security (SEC-1 → SEC-9)
1. Input validation — all user input validated before use
2. Authentication — endpoints require auth where needed
3. Secrets — no hardcoded credentials, use env vars
4. SQL injection — parameterized queries only
5. XSS — output encoding, CSP headers
6. CSRF — token validation on state-changing requests
7. Sensitive data — no PII in logs or error responses
8. Rate limiting — on public/expensive endpoints
9. Dependencies — no known vulnerabilities

## Resilience (RES-1 → RES-5)
1. Idempotency — mutation endpoints safely retriable
2. Payload limits — request size limits configured
3. Panic recovery — middleware catches panics
4. Graceful degradation — external service failures handled
5. Timeout budgets — end-to-end timeouts for multi-step ops

## Observability (OBS-1 → OBS-3)
1. Correlation IDs — request trace IDs propagated
2. Health checks — verify dependencies
3. Error rate visibility — metrics on critical paths

## Accessibility (A11Y-1 → A11Y-4)
1. Semantic HTML — proper elements (nav, main, section)
2. ARIA — labels, roles where needed
3. Keyboard navigation — all interactive elements reachable
4. Contrast — meets WCAG guidelines

## Database (DB-1 → DB-7)
1. New migrations only — never modify existing
2. Reversible — migrations can be rolled back
3. Indexes — on frequently queried columns
4. N+1 — no N+1 query patterns
5. Pagination — cursor or offset on list endpoints
6. Transaction boundaries — multi-table ops wrapped
7. Financial precision — decimal types, never float

## Testing (TST-1 → TST-9)
1. Unit tests — business logic covered
2. Integration tests — API endpoints tested end-to-end
3. Edge cases — boundary values, empty inputs
4. Coverage — critical paths tested
5. Naming — descriptive test names
6. Independence — tests don't depend on execution order
7. Mock boundaries — mock at system boundaries only
8. Critical paths — auth, payments, data mutations tested
9. Contract tests — at service boundaries

## Performance (PERF-1 → PERF-5)
1. N+1 queries — batch or join instead
2. Indexes — queries use indexed columns
3. Pagination — no unbounded result sets
4. Caching — hot paths cached appropriately
5. Async — non-blocking I/O where applicable
