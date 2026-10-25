# Tenant-scoped stores

How the backend keeps one collector (tenant) away from the data of another, and the recipe for adding a new tenant-owned resource. Reference implementation: personas (`port.PersonaStore`).

## The problem

A *collector* is the tenant. Everything a collector owns (personas today, clients and uploads later) must only be reachable by that collector. If repository methods take a `collectorID` argument, every use case has to remember to pass the right one, and a forgotten or wrong filter silently leaks data across tenants. Authentication does not remove this: it tells us who is calling, not that every query filters by it.

## The pattern

Make the safe thing the only possible thing: a repository that is **bound to one collector**, so none of its methods has a collector parameter and none can reach another tenant's rows.

```go
// internal/core/port/repositories.go
type PersonaStore interface {
    For(collectorID domain.CollectorID) PersonaRepository
}

type PersonaRepository interface {            // acts on the bound collector only
    GetPersona(ctx, id domain.PersonaID) (*domain.Persona, error)
    ListPersonas(ctx) ([]*domain.Persona, error)
    SavePersona(ctx, p *domain.Persona) error
}

type Repositories struct {
    Collectors CollectorRepository            // the tenant root, deliberately unscoped
    Personas   PersonaStore                   // tenant-owned resources are stores
}
```

A use case reads the tenant from the context and asks the store for a bound repository:

```go
collectorID, err := tenant.MustCollectorFrom(ctx)
...
p, err := u.personas.For(collectorID).GetPersona(ctx, id)
```

### Rules

- **Tenant-owned resources are stores.** A new field in `port.Repositories` for something a collector owns is a `...Store`, never a plain repository. Its repository methods take no collector id.
- **`CollectorRepository` stays unscoped.** The collector is the tenant itself; it is looked up and created by id.
- **Reading another tenant's id looks like "not found".** `GetPersona` with an id of another collector returns `domain.ErrPersonaNotFound`, same as a missing id (no existence oracle).
- **Saving a foreign entity is a programming error.** `SavePersona` of a persona whose `CollectorID()` differs from the bound one returns `port.ErrTenantMismatch`. It is not a domain error, so it surfaces as a 500 and is never shown to the client.
- **Keep the SQL guard too.** `ON CONFLICT ... WHERE personas.collector_id = EXCLUDED.collector_id` stays as a second line of defence for ids that belong to another collector.

### Cost

Nothing is stored per tenant. There is one store per process, wired once on the shared pool (or the current transaction). `For` returns a small value `{db, collectorID}` (a few words) that is usually created once per request and garbage-collected afterwards. No map, cache, connection or schema per collector.

## Adding a new tenant-owned resource (checklist)

1. **Port** (`internal/core/port/repositories.go`): add `XStore` with `For(collectorID) XRepository` and the `XRepository` interface without collector parameters. Document the not-found and `ErrTenantMismatch` behaviour on the methods. Add `X XStore` to `Repositories` (`transactor.go`).
2. **Postgres** (`internal/adapters/outbound/postgres/<x>`): `Store{db pgdb.DBTX}` with `NewStore(db)`, `For(id)` returning `&Repository{db, collectorID}`. Every query filters on `r.collectorID` (`WHERE <table>.collector_id = $n`). Reject a mismatching entity in the save method with `port.ErrTenantMismatch`. Add `var _ port.XStore = (*Store)(nil)` and the same for the repository.
3. **Fake** (`internal/adapters/outbound/fake/<x>`): a `Store` owning the data and mutex; `For` returns a view that filters on the bound id.
4. **Wiring**: add `X: pgx.NewStore(db)` to `postgresRepositories` in `internal/backend/backend.go`, and to the repository literal in `cmd/e2e/main.go`.
5. **Use cases**: depend on `port.XStore` and call `u.xs.For(collectorID)` after `tenant.MustCollectorFrom(ctx)`. Never pass a collector id from request input other than the one in the tenant context.
6. **Contract tests** (`internal/core/port/porttest`): take a store factory (`func(t, collectors ...domain.CollectorID) port.XStore`), save each entity through `store.For(entity.CollectorID())`, and include an isolation test. Copy `TestPersonaRepositoryIsolation`: another collector cannot get by id, does not list, and a foreign save fails with `port.ErrTenantMismatch` and changes nothing. Run it from both the fake and the postgres test packages.
7. **Test doubles in use case tests**: a store stub with `For` returning a repository stub (see `failingStore` in `usecases/getpersona/usecase_test.go`).

## Known limits

- Ids that are unique across all tenants (item definition ids) still produce `domain.ErrItemIDTaken` when a save reuses another tenant's id, which tells the caller that the id exists elsewhere. Prefer ids that cannot clash (random UUIDs) and consider scoping uniqueness per parent.
- The pattern is enforced by types, not by the database. Postgres row-level security (`SET LOCAL app.collector_id` per transaction in `pgdb.Transactor.InTx`) would add a database-level layer; it is not implemented.
