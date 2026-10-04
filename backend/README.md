# Backend

Go backend for tax-return-organizer. See the [repo-level README](../README.md) for the overall architecture, and [TODO.md](../TODO.md) for the build order.

## Layout

```
internal/
  core/
    domain/   entities + business rules, no external/infra dependencies (except google/uuid for IDs)
    port/     outbound repository interfaces the domain/use-case layer depends on;
              adapters (e.g. Postgres) implement these later
```

`internal/core` is hexagonal architecture's core: `domain` and `port` know nothing about Postgres, HTTP, or any other adapter. Adapters (Postgres repository implementations, the REST API) are added in later steps, live outside `core`, and depend inward on `core/port`/`core/domain` — never the other way around.

### Domain types

The system is multi-tenant: there can be many collectors, each with their own clients.

- `Collector`, `Client` — tenant roots, identified by app-generated UUIDs (`NewCollector`, `NewClient`). A `Client` belongs to exactly one `Collector`.
- `ClientTaxYear` — one client's tax year, scoped to a `Client`.
- `Category`, `ChecklistTemplate` — defined by a `Collector`, reused across that collector's clients. Every `ChecklistTemplate` belongs to exactly one `Category`, has a recurrence (`once`/`monthly`), and a `Deadline` rule (days after the covered period ends — the tax year itself for `once`, each covered month for `monthly`).
- `ChecklistItem` — one client's instance of a template requirement for a given tax year (and month, if monthly), scoped to a `ClientTaxYear`. `ChecklistTemplate.GenerateItems()` expands a template into its items for one client's tax year, computing each item's absolute due date from the template's `Deadline` rule — this is business logic, so it lives in the domain rather than in a repository or handler. An item is satisfied via `SatisfyWithDocument` or `SatisfyWithValue` (a structured value entered directly, e.g. a homeoffice-days count) — never both; `IsOverdue` checks completion against the due date.
- `Document` — a file a client uploaded (blob + metadata), scoped to a `ClientTaxYear`, optionally classified under a `Category` — constructed via `NewDocument`, which enforces non-empty filename/mime type/content.

### Ports

`internal/core/port/repository.go` declares one repository interface per aggregate (`CollectorRepository`, `ClientRepository`, `ClientTaxYearRepository`, `CategoryRepository`, `ChecklistTemplateRepository`, `ChecklistItemRepository`, `DocumentRepository`). `port.ErrNotFound` is the sentinel adapters should wrap so callers can check with `errors.Is`.

No implementation exists yet — that's a later TODO step (Postgres adapter, including the schema/migrations, derived from what these interfaces need).

## Database migrations

Not written yet. Per [TODO.md](../TODO.md), migrations are adapter-owned infrastructure and are designed together with the Postgres adapter so the schema follows from what the domain/ports actually need, rather than the domain being reverse-engineered from a schema written first.

## Running tests

```sh
cd backend
go test ./...
```
