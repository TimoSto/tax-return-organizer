# Tenant-scoped stores

How the backend keeps one collector (tenant) away from the data of another, and the recipe for adding a new tenant-owned resource. Reference implementation: personas (`outbound.PersonaStore`).

## The problem

A *collector* is the tenant. Everything a collector owns (personas today, clients and uploads later) must only be reachable by that collector. If repository methods take a `collectorID` argument, every use case has to remember to pass the right one, and a forgotten or wrong filter silently leaks data across tenants. Authentication does not remove this: it tells us who is calling, not that every query filters by it.

## The pattern

Make the safe thing the only possible thing: a repository that is **bound to one collector**, so none of its methods has a collector parameter and none can reach another tenant's rows.

```go
// internal/core/port/outbound/repositories.go
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

A use case takes the caller as an explicit `domain.Actor` argument (the inbound adapter supplies it, see `httpx.Actor`), checks what it may do, and asks the store for a repository bound to the actor's collector:

```go
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor, id domain.PersonaID) (*domain.Persona, error) {
    if err := actor.RequireStaff(); err != nil {
        return nil, fmt.Errorf("get persona: %w", err)
    }

    p, err := u.personas.For(actor.CollectorID()).GetPersona(ctx, id)
```

### Rules

- **Tenant-owned resources are stores.** A new field in `outbound.Repositories` for something a collector owns is a `...Store`, never a plain repository. Its repository methods take no collector id.
- **`CollectorRepository` stays unscoped.** The collector is the tenant itself; it is looked up and created by id.
- **Reading another tenant's id looks like "not found".** `GetPersona` with an id of another collector returns `domain.ErrPersonaNotFound`, same as a missing id (no existence oracle).
- **Saving a foreign entity is a programming error.** `SavePersona` of a persona whose `CollectorID()` differs from the bound one returns `outbound.ErrTenantMismatch`. It is not a domain error, so it surfaces as a 500 and is never shown to the client.
- **Stores check that the collector exists, use cases do not.** Referential integrity is the storage's job: every repository method returns `domain.ErrCollectorNotFound` if the bound collector does not exist, and a rejected save writes nothing. The collector takes precedence over every other outcome, so an unknown collector never shows up as `ErrPersonaNotFound`. Postgres gets this from the foreign key on writes (map the violation of `<table>_collector_id_fkey`, SQLSTATE `23503`, to the sentinel; where an upsert can affect no row without inserting, as for a persona id of another collector, check `collectors` in the same batch to tell the causes apart) and by reading through `collectors` on reads (`FROM collectors c LEFT JOIN <table> ...`: no row means no collector, so an empty list is never ambiguous). Fakes get an `outbound.CollectorRepository` and look the collector up. The contract tests (`porttest.TestPersonaRepositoryUnknownCollector`) pin the behaviour for every adapter, so it does not depend on one database happening to raise an error. Use cases call `store.For(actor.CollectorID())` directly (`getpersona`, `listpersonas`).
- **A repository checks only what it is bound to.** The collector existence check covers the collector of `For` and nothing deeper. Resources nested inside a tenant (later: client → tax year → upload) do not report missing ancestors with their own sentinels. A child whose parent is missing is simply not found (`ErrUploadNotFound`, not `ErrClientNotFound` or `ErrTaxYearNotFound`), and the parent's not-found comes from asking the parent's own repository or aggregate. Without this rule, every query would need a LEFT JOIN chain over all ancestors, a precedence order per method, and fakes that emulate each foreign key. If a use case must tell "no such client" from "client has no uploads", it reads the client first, through its own repository. One extra read in the one use case that needs it is cheaper than ancestor checks in every adapter.
- **Deleting a collector cascades.** Tenant tables reference `collectors` with `ON DELETE CASCADE`, and deleting a collector relies on that instead of deleting its resources one by one; fakes must emulate it. Not implemented yet: there is no delete use case, and `personas.collector_id` still has the default `NO ACTION`. Add the cascade in a migration together with the first delete use case.
- **Keep the SQL guard too.** `ON CONFLICT ... WHERE personas.collector_id = EXCLUDED.collector_id` stays as a second line of defence for ids that belong to another collector.

### Cost

Nothing is stored per tenant. There is one store per process, wired once on the shared pool (or the current transaction). `For` returns a small value `{db, collectorID}` (a few words) that is usually created once per request and garbage-collected afterwards. No map, cache, connection or schema per collector.

## Adding a new tenant-owned resource (checklist)

1. **Port** (`internal/core/port/outbound/repositories.go`): add `XStore` with `For(collectorID) XRepository` and the `XRepository` interface without collector parameters. Document the not-found and `ErrTenantMismatch` behaviour on the methods. Add `X XStore` to `Repositories` (`transactor.go`).
2. **Postgres** (`internal/adapters/outbound/postgres/<x>`): `Store{db pgdb.DBTX}` with `NewStore(db)`, `For(id)` returning `&Repository{db, collectorID}`. Every query filters on `r.collectorID`, reads start from `collectors` and writes map the foreign key violation to `domain.ErrCollectorNotFound` (see the rules above). Reject a mismatching entity in the save method with `outbound.ErrTenantMismatch`. Add `var _ outbound.XStore = (*Store)(nil)` and the same for the repository.
3. **Fake** (`internal/adapters/outbound/fake/<x>`): a `Store` owning the data and mutex, built with `NewStore(collectors outbound.CollectorRepository)`; `For` returns a view that filters on the bound id, and every method first checks the collector there.
4. **Wiring**: add `X: pgx.NewStore(db)` to `postgresRepositories` in `internal/backend/backend.go`, and to the repository literal in `cmd/e2e/main.go`.
5. **Use cases**: depend on `outbound.XStore` (not on `CollectorRepository`), take the `domain.Actor` as an argument of `Execute`, check what it may do (see "Actors and authorization" in `ARCHITECTURE.md`) and call `u.xs.For(actor.CollectorID())`; an unknown collector comes back from the store as `domain.ErrCollectorNotFound`. Handlers get the actor only from `httpx.Actor`, never from the request body or other input.
6. **Contract tests** (`internal/testutil/porttest`): take a store factory (`func(t, collectors ...domain.CollectorID) outbound.XStore`), save each entity through `store.For(entity.CollectorID())`, and include an isolation test and an unknown-collector test. Copy `TestPersonaRepositoryIsolation`: another collector cannot get by id, does not list, and a foreign save fails with `outbound.ErrTenantMismatch` and changes nothing. Copy `TestPersonaRepositoryUnknownCollector`: every method returns `domain.ErrCollectorNotFound` for a collector the factory did not create, and the rejected save leaves nothing behind. Run both from the fake and the postgres test packages.
7. **Test doubles in use case tests**: a store stub with `For` returning a repository stub (see `failingStore` in `usecases/getpersona/usecase_test.go`).

## Known limits

- Ids that are unique across all tenants (item definition ids) still produce `domain.ErrItemIDTaken` when a save reuses another tenant's id, which tells the caller that the id exists elsewhere. Prefer ids that cannot clash (random UUIDs) and consider scoping uniqueness per parent.
- The pattern is enforced by types, not by the database. Postgres row-level security (`SET LOCAL app.collector_id` per transaction in `pgdb.Transactor.InTx`) would add a database-level layer; it is not implemented.
