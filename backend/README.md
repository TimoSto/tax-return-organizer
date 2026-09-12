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

- `TaxYear`, `Category`, `Document`, `ChecklistTemplate`, `ChecklistItem` — constructed via `New*` functions that enforce invariants (e.g. a document needs non-empty content, a checklist item's month is only set for a `monthly` template).
- `ChecklistTemplate.GenerateItems()` expands a template into its items (one for `once`, twelve for `monthly`) — this is business logic, so it lives in the domain rather than in a repository or handler.

### Ports

`internal/core/port/repository.go` declares one repository interface per aggregate (`TaxYearRepository`, `CategoryRepository`, `DocumentRepository`, `ChecklistTemplateRepository`, `ChecklistItemRepository`). `port.ErrNotFound` is the sentinel adapters should wrap so callers can check with `errors.Is`.

No implementation exists yet — that's the next TODO step (Postgres adapter, including the schema/migrations, derived from what these interfaces need).

## Database migrations

Not written yet. Per [TODO.md](../TODO.md), migrations are adapter-owned infrastructure and are designed together with the Postgres adapter (next step) so the schema follows from what the domain/ports actually need, rather than the domain being reverse-engineered from a schema written first.

## Running tests

```sh
cd backend
go test ./...
```
