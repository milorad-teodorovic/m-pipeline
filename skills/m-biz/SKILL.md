---
name: m-biz
description: Analyze business logic to align code structure with business intent. Use when features involve business rules, pricing, workflows, compliance, authorization policies, domain modeling, or stakeholder-driven requirements; or when a `.business/` directory exists.
argument-hint: [feature or domain area]
model: opus
effort: max
user-invocable: false
paths:
  - ".business/**"
  - "**/BUSINESS.md"
---
# /m-biz - Business Logic Overview & Alignment

Role: business analysis and software architecture. Responsibilities: understand **why** a business decision was made, map domain logic, and recommend how to structure code and features so they serve business intent cleanly and sustainably.

Activates on features with significant business rules, domain logic, pricing, workflows, compliance, or authorization policies.

## Input

Feature or domain area: `$ARGUMENTS`

If no arguments provided, ask the user which business domain to analyze.

## Business Context Sources

Before analyzing, check for a `.business/` directory in the project root:

1. **`.business/BUSINESS.md`** — project-level business logic documentation
2. **`.business/*.pdf`**, **`.business/*.xlsx`**, **`.business/*.docx`** — specification documents, contracts, regulatory requirements
3. Read any available specs before asking the user to explain the business context. These documents are the source of truth for business rules.

Also read `.m/INDEX.md`, `PROJECT_INDEX.md`, and relevant domain code in the repository.

## Workflow

### Phase 1: Domain Discovery

1. **Check `.business/` for specs** — read BUSINESS.md and any available spec documents first
2. **Identify the business domain** — what part of the business does this feature serve?
3. **Map existing implementation** — read the codebase for related domain logic, trace the flow
4. **Research the domain** — use web search for industry-standard approaches, regulatory requirements, common pitfalls, how similar products handle this

### Phase 2: Business Intent Analysis

For the feature or domain area, answer:

1. **Why does this exist?** — business problem, who benefits, what happens if we build it wrong?
2. **What are the business rules?**
   - Explicit rules (documented in `.business/`, regulated, contractual)
   - Implicit rules (conventions, expectations, edge cases)
   - Rules that might change (policy-driven vs. structural)
3. **Who are the stakeholders?** — end users, operators, compliance/legal, engineering
4. **What are the business metrics?** — success measurement, failure costs, observability needs

If the user can't answer business questions and no `.business/` specs exist, recommend they consult stakeholders before proceeding. Do not invent business rules.

### Phase 3: Domain Modeling

Map business logic into software concepts:

1. **Entities and Value Objects** — core domain objects and their invariants
2. **Business Rules as Code** — categorize each rule:
   - **Structural** (hardcode) — unlikely to change, fundamental to the domain
   - **Policy** (configurable) — changes with business needs
   - **External** (fetched) — comes from another system or database
3. **State Machines** — valid states, transitions, who/what triggers them
4. **Workflows** — multi-step processes, ordering constraints, compensation logic
5. **Boundaries** — where this domain ends and another begins, API contracts

### Phase 4: Code Structure Recommendation

1. Package/module structure that mirrors the domain (not technical layers)
2. Where business rules live — centralized vs. distributed
3. What to make configurable vs. hardcode
4. Testing strategy — which rules are critical enough for comprehensive coverage
5. Extensibility points — where the business is likely to want changes

## Output

Produce:
- Business Context (why, problem, who)
- Business Rules Map (structural / policy / external, with source and changeability)
- Domain Model (entities, state machines, workflows)
- Stakeholder Impact (who's affected, risk per stakeholder)
- Recommended Code Structure (packages, services, rule placement)
- Configuration vs. Hardcode Decisions
- Risk Assessment (what if rules are wrong? blast radius?)
- Open Business Questions (needs stakeholder input, not engineering judgment)

## Rules

- **If business rules are unclear from code and specs alone, say "I don't know — this needs stakeholder input" rather than inventing rules.**
- Check `.business/` directory first — specs are the source of truth
- Prefer domain-driven boundaries over technical-layer boundaries
- Flag when code structure fights the business model — these are expensive long-term
- Distinguish between "the business wants X" and "the code currently does X" — they may differ
- Things that change together should live together
- End by asking whether to proceed to `/m:refine` or `/m:plan`
