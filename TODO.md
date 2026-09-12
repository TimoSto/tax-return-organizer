# TODO

Rough build order for implementing the architecture described in [README.md](README.md), following dependencies bottom-up (data model → backend core → backend API → containerized backend → frontend → full stack → auth last).

## 1. Data & backend core
- [ ] Write Postgres schema as migrations (tax_years, categories, documents, checklist_templates, checklist_items)
- [ ] Go: domain types + repository ports (hexagonal core), no infra yet
- [ ] Go: Postgres adapter implementing the repository ports
- [ ] Go: REST API for document CRUD (upload, list, get, delete) + categories
- [ ] docker compose: backend + postgres only, verify end-to-end via curl/Postman
- [ ] Go: checklist_templates/checklist_items endpoints (create, generate monthly instances, mark done)
- [ ] Go: classification (start rule/keyword-based, behind its own port so it's swappable)
- [ ] Go: export endpoint (bundle documents per category/year)

## 2. Frontend
- [ ] SvelteKit BFF: skeleton + proxy routes to the Go API
- [ ] UI: bar (year select) + rail + overview/checklist view
- [ ] UI: document tree view, per-folder upload, metadata edit pane
- [ ] UI: category list views
- [ ] UI: hook up export

## 3. Full stack & hardening
- [ ] docker compose: add BFF container, run all three together
- [ ] Auth (deferred on purpose until everything above works)
