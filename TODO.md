# TODO

Rough build order for implementing the architecture described in [README.md](README.md), following dependencies bottom-up (backend core → backend adapters/API → containerized backend → frontend → full stack → auth last).

Domain-first: the domain types and repository ports are designed from what the use cases need, *then* the Postgres schema/migrations are derived to satisfy them — not the other way around. Schema/migrations are adapter-owned infrastructure detail, so they're written as part of the Postgres adapter step, not before the domain exists.

Multi-tenant from the start: the system is multi-tenant (many collectors, each with their own clients), so `Collector` and `Client` are real domain entities with persistent IDs from the first step, and every other entity is scoped to one of them. Auth is still deferred — for now collectors/clients are created via API/seed data with no login, but the scoping itself has to exist from day one so it doesn't get retrofitted later.

Use-cases-before-adapters: application services depend only on `port` interfaces, so they're built and unit-tested against mocks/in-memory fakes before any concrete adapter exists. This exercises the ports for real and catches interface design mistakes before writing SQL against them.

## 1. Data & backend core
- [ ] Go: domain types + repository ports (hexagonal core), no infra yet
    - `Collector`, `Client` (a `Client` belongs to exactly one `Collector`)
    - `ClientTaxYear` — scoped to a `Client`
    - `Document` - Scoped to a `ClientTaxYear` 
    - `Category`, `ChecklistTemplate` — scoped to a `Collector`, reused across its clients; carries a recurrence of `once`/`monthly`. No deadline tracking in the first MVP (deferred — see [README.md](README.md))
    - `ChecklistItem` — one per `ClientTaxYear` per template item (generated from a template: one for `once`, twelve for `monthly`), satisfied by a typed `Value` (a linked `Document` or a text/number/year/bool entered directly), as declared by the template's `Requirement`
- [ ] Go: use cases/application services for collector setup — manage categories and checklist templates for a collector, tested against ports with mocks/in-memory fakes — no adapter yet
- [ ] Go: use cases/application services for client management — create/list clients under a collector, create/list tax years for a client
- [ ] Go: use cases/application services for document operations scoped to a client (upload, list, get, delete; set metadata)
- [ ] Go: checklist use cases — generate items from a template, satisfy an item via document or structured value, mark undone, list items with completion status (missing/done) for in-app display
- [ ] Go: Postgres adapter — schema as migrations (collectors, clients, client_tax_years, categories, documents, checklist_templates, checklist_items) + repository ports implemented against it
- [ ] Go: REST API — collector-side endpoints (categories, checklist templates, client management)
- [ ] Go: REST API — client-side endpoints (tax years, documents, checklist items incl. completion status)
- [ ] docker compose: backend + postgres only, verify end-to-end via curl/Postman
- [ ] Go: export use case (bundle a client's documents + structured data per category/year) + endpoint

## 2. Frontend
- [ ] SvelteKit BFF: skeleton + proxy routes to the Go API
- [ ] UI: collector page — manage categories/checklist templates, manage clients
- [ ] UI: client page — bar (year select) + rail + overview/checklist view, with outstanding (not yet satisfied) items highlighted in-app
- [ ] UI: document tree view, per-item upload, metadata edit pane, structured-value entry (in place of upload)
- [ ] UI: hook up export

## 3. Full stack & hardening
- [ ] docker compose: add BFF container, run all three together
- [ ] Auth (deferred on purpose until everything above works) — introduces collector/client login, tying the existing collector/client scoping to authenticated identity
