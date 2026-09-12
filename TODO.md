# TODO

Rough build order for implementing the architecture described in [README.md](README.md), following dependencies bottom-up (backend core → backend adapters/API → containerized backend → frontend → full stack → auth last).

Domain-first: the domain types and repository ports are designed from what the use cases need, *then* the Postgres schema/migrations are derived to satisfy them — not the other way around. Schema/migrations are adapter-owned infrastructure detail, so they're written as part of the Postgres adapter step, not before the domain exists.

Use-cases-before-adapters: application services depend only on `port` interfaces, so they're built and unit-tested against mocks/in-memory fakes before any concrete adapter exists. This exercises the ports for real and catches interface design mistakes before writing SQL against them.

## 1. Data & backend core
- [x] Go: domain types + repository ports (hexagonal core), no infra yet
- [ ] Go: use cases/application services for core document + category operations (upload, list, get, delete; category CRUD), tested against ports with mocks/in-memory fakes — no adapter yet
- [ ] Go: Postgres adapter — schema as migrations (tax_years, categories, documents, checklist_templates, checklist_items) + repository ports implemented against it
- [ ] Go: REST API wiring the use cases (document CRUD + categories)
- [ ] docker compose: backend + postgres only, verify end-to-end via curl/Postman
- [ ] Go: checklist use cases (create template, generate monthly instances, mark done) + endpoints
- [ ] Go: classification use case (start rule/keyword-based, behind its own port so it's swappable) + wiring
- [ ] Go: export use case (bundle documents per category/year) + endpoint

## 2. Frontend
- [ ] SvelteKit BFF: skeleton + proxy routes to the Go API
- [ ] UI: bar (year select) + rail + overview/checklist view
- [ ] UI: document tree view, per-folder upload, metadata edit pane
- [ ] UI: category list views
- [ ] UI: hook up export

## 3. Full stack & hardening
- [ ] docker compose: add BFF container, run all three together
- [ ] Auth (deferred on purpose until everything above works)
