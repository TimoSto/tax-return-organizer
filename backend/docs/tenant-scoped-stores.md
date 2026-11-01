# Tenant-scoped stores

How the backend keeps one tax practice (tenant) away from the data of another, and the recipe for adding a new tenant-owned resource. Reference implementation: personas (`outbound.PersonaStore`).

## The problem

A *tax practice* is the tenant. Everything a tax practice owns (personas today, clients and uploads later) must only be reachable by that tax practice. If repository methods take a `taxPracticeID` argument, every use case has to remember to pass the right one, and a forgotten or wrong filter silently leaks data across tenants. Authentication does not remove this: it tells us who is calling, not that every query filters by it.

## The pattern

Make the safe thing the only possible thing: a repository that is **bound to one tax practice**, so none of its methods has a tax practice parameter and none can reach another tenant's rows.

```go
// internal/core/port/outbound/repositories.go
type PersonaStore interface {
    For(taxPracticeID domain.TaxPracticeID) PersonaRepository
}

type PersonaRepository interface {            // acts on the bound tax practice only
    GetPersona(ctx, id domain.PersonaID) (*domain.Persona, error)
    ListPersonas(ctx) ([]*domain.Persona, error)
    SavePersona(ctx, p *domain.Persona) error
}

// internal/backend: the full set, only in the composition root
type Repositories struct {
    TaxPractices outbound.TaxPracticeRepository // the tenant root, deliberately unscoped
    Personas     outbound.PersonaStore          // tenant-owned resources are stores
}
```

A use case takes the caller as an explicit `domain.Actor` argument (the inbound adapter supplies it, see `httpx.Actor`), checks what it may do, and asks the store for a repository bound to the actor's tax practice:

```go
func (u *Usecase) Execute(ctx context.Context, actor domain.Actor, id domain.PersonaID) (*domain.Persona, error) {
    if err := actor.RequireStaff(); err != nil {
        return nil, fmt.Errorf("get persona: %w", err)
    }

    p, err := u.personas.For(actor.TaxPracticeID()).GetPersona(ctx, id)
```

### Rules

- **Tenant-owned resources are stores.** A new field in `backend.Repositories` (or in a use case's own repository set) for something a tax practice owns is a `...Store`, never a plain repository. Its repository methods take no tax practice id.
- **`TaxPracticeRepository` stays unscoped.** The tax practice is the tenant itself; it is looked up and created by id.
- **Reading another tenant's id looks like "not found".** `GetPersona` with an id of another tax practice returns `domain.ErrPersonaNotFound`, same as a missing id (no existence oracle).
- **Saving a foreign entity is a programming error.** `SavePersona` of a persona whose `TaxPracticeID()` differs from the bound one returns `outbound.ErrTenantMismatch`. It is not a domain error, so it surfaces as a 500 and is never shown to the client.
- **Stores check that the tax practice exists, use cases do not.** Referential integrity is the storage's job: every repository method returns `domain.ErrTaxPracticeNotFound` if the bound tax practice does not exist, and a rejected save writes nothing. The tax practice takes precedence over every other outcome, so an unknown tax practice never shows up as `ErrPersonaNotFound`. Postgres gets this from the foreign key on writes (map the violation of `<table>_tax_practice_id_fkey`, SQLSTATE `23503`, to the sentinel; where an upsert can affect no row without inserting, as for a persona id of another tax practice, check `tax practices` in the same batch to tell the causes apart) and by reading through `tax practices` on reads (`FROM tax_practices c LEFT JOIN <table> ...`: no row means no tax practice, so an empty list is never ambiguous). The contract tests (`porttest.TestPersonaRepositoryUnknownTaxPractice`) pin the behaviour, so it does not depend on one database happening to raise an error. Use cases call `store.For(actor.TaxPracticeID())` directly (`getpersona`, `listpersonas`).
- **A repository checks only what it is bound to.** The tax practice existence check covers the tax practice of `For` and nothing deeper. Resources nested inside a tenant (later: client → tax year → upload) do not report missing ancestors with their own sentinels. A child whose parent is missing is simply not found (`ErrUploadNotFound`, not `ErrClientNotFound` or `ErrTaxYearNotFound`), and the parent's not-found comes from asking the parent's own repository or aggregate. Without this rule, every query would need a LEFT JOIN chain over all ancestors, and a precedence order per method. If a use case must tell "no such client" from "client has no uploads", it reads the client first, through its own repository. One extra read in the one use case that needs it is cheaper than ancestor checks in every adapter.
- **Deleting a tax practice cascades.** Tenant tables reference `tax practices` with `ON DELETE CASCADE`, and deleting a tax practice relies on that instead of deleting its resources one by one. Not implemented yet: there is no delete use case, and `personas.tax_practice_id` still has the default `NO ACTION`. Add the cascade in a migration together with the first delete use case.
- **Every tenant table carries `tax_practice_id`.** A table that only reaches its tenant through a parent (`item_definitions` through `persona_id`) still gets its own `tax_practice_id NOT NULL`. It references the parent through `(parent_id, tax_practice_id)`, with a `UNIQUE (id, tax_practice_id)` on the parent as the target: `item_definitions` has `FOREIGN KEY (persona_id, tax_practice_id) REFERENCES personas (id, tax_practice_id)`. The database then refuses a child whose tax practice differs from its parent's, whatever a statement says, so a whole chain (later client → tax year → item → document) belongs to one tenant. Writes filter on `tax_practice_id` directly instead of guarding every statement with `EXISTS (... parent ...)`; an insert under another tenant's parent fails on the foreign key. Reads join on both columns (`i.persona_id = p.id AND i.tax_practice_id = p.tax_practice_id`). It is also what row-level security would need, a direct policy on every table (see [roadmap #8](ARCHITECTURE_ROADMAP.md#8-row-level-security-as-a-second-tenant-guard)).
- **Keep the SQL guard too.** `ON CONFLICT ... WHERE personas.tax_practice_id = EXCLUDED.tax_practice_id` stays as a second line of defence for ids that belong to another tax practice.

### Cost

Nothing is stored per tenant. There is one store per process, wired once on the shared pool (or the current transaction). `For` returns a small value `{db, taxPracticeID}` (a few words) that is usually created once per request and garbage-collected afterwards. No map, cache, connection or schema per tax practice.

## Adding a new tenant-owned resource (checklist)

1. **Port** (`internal/core/port/outbound/repositories.go`): add `XStore` with `For(taxPracticeID) XRepository` and the `XRepository` interface without tax practice parameters. Document the not-found and `ErrTenantMismatch` behaviour on the methods.
2. **Postgres** (`internal/adapters/outbound/postgres/<x>`): `Store{db pgdb.DBTX}` with `NewStore(db)`, `For(id)` returning `&Repository{db, taxPracticeID}`. The migration gives the table a `tax_practice_id NOT NULL` and, for a child table, the composite foreign key to its parent (see the rules above). Every query filters on `r.taxPracticeID`, reads start from `tax practices` and writes map the foreign key violation to `domain.ErrTaxPracticeNotFound` (see the rules above). Reject a mismatching entity in the save method with `outbound.ErrTenantMismatch`. Add `var _ outbound.XStore = (*Store)(nil)` and the same for the repository.
3. **Wiring**: add `X outbound.XStore` to `Repositories` and `X: pgx.NewStore(db)` to `postgresRepositories` in `internal/backend/backend.go`. A transactional use case that writes X adds it to its own repository set and to its pick in `internal/backend`.
4. **Use cases**: depend on `outbound.XStore` (not on `TaxPracticeRepository`), take the `domain.Actor` as an argument of `Execute`, check what it may do (see "Actors and authorization" in `ARCHITECTURE.md`) and call `u.xs.For(actor.TaxPracticeID())`; an unknown tax practice comes back from the store as `domain.ErrTaxPracticeNotFound`. Handlers get the actor only from `httpx.Actor`, never from the request body or other input.
5. **Contract tests** (`internal/testutil/porttest`): take a store factory (`func(t, tax practices ...domain.TaxPracticeID) outbound.XStore`), save each entity through `store.For(entity.TaxPracticeID())`, and include an isolation test and an unknown-tax-practice test. Copy `TestPersonaRepositoryIsolation`: another tax practice cannot get by id, does not list, and a foreign save fails with `outbound.ErrTenantMismatch` and changes nothing. Copy `TestPersonaRepositoryUnknownTaxPractice`: every method returns `domain.ErrTaxPracticeNotFound` for a tax practice the factory did not create, and the rejected save leaves nothing behind. Run them from the postgres test package.
6. **Test doubles in use case tests**: a store stub with `For` returning a repository stub that hands out canned results and applies no storage rules (see `personaStub` and `failingStore` in `usecases/getpersona/usecase_test.go`). There is no in-memory implementation of the store.

## Known limits

- Child ids are unique per parent, not globally: `item_definitions` has the primary key `(persona_id, id)`, so an id that another persona or tenant uses never clashes and a save cannot reveal that it exists. Follow the same rule for new child tables. A clash within one parent is a bug and a plain error (500), like `domain.ErrItemIDTaken`.
- Tenant isolation is enforced by types, not by the database: the database ensures that the rows of a chain belong to one tenant (composite foreign keys), but not that a query only reaches the caller's tenant. Postgres row-level security would add that layer; it is not implemented. It would not need a `Transactor` in the use cases: each repository method sets its bound tax practice with `set_config('app.tax_practice_id', $1, true)` in the same batch or transaction as its own statements. The `tax_practice_id` it needs is already on every tenant table (see the rules above). The rest of the design and its consequences (the policies, a separate app role, tests that do not connect as superuser) are in [roadmap #8](ARCHITECTURE_ROADMAP.md#8-row-level-security-as-a-second-tenant-guard).
