# Architecture review of `backend/`

The structure is good. The layering is enforced by depguard, tenant isolation is built into the types, and the domain is well encapsulated. The weaknesses below are mostly places where the design's own rules are bent, or where the tenant model is not ready for authentication. They are ordered by how much they matter.

This review is from reading the code. The tests and the linter were not run.

## 1. Tenancy is a hidden dependency (`tenant.Require` from `context`)

> **Done.** Use cases take the collector id as an argument, `httpx.Tenant` is the single place that resolves it, and `core/tenant`, `Route.Tenant` and `checkPathTenants` are gone.

**Where:** `core/tenant`, every tenant-scoped use case (`getpersona`, `listpersonas`).

**Problem:** The tenant is a context value, so a use case's signature (`Execute(ctx, id)`) doesn't say it needs one.
- A forgotten `withTenant` or a `Tenant: false` flag only fails at runtime (`ErrNoTenant`, a 500).
- `checkPathTenants` is a startup check that exists only to compensate for this.
- The use-case contract becomes implicit, and each new use case has to remember to call `Require`.

**Fix:** Make the tenant part of the input.
- Use `Execute(ctx, collectorID, personaID)` or `Execute(ctx, Query{Collector: ..., ID: ...})`.
- The HTTP adapter resolves the principal and passes it explicitly. Later, an authenticated `Principal` or `Actor` type can take its place.
- If you want to keep the ergonomics, keep `tenant` only as an adapter-side convention. Don't let the core read from `ctx`.

The inbound port becomes explicit and the compiler enforces it. The tenant check moves from a runtime to a compile-time guarantee.

## 2. The "collector exists" rule lived in the callers

> **Done.** Use cases check the collector (`getpersona` and `listpersonas` call `CollectorRepository.GetCollector` first), the adapters stay dumb, and the foreign key is only a safety net.

**Where:** `listpersonas/usecase.go`, `getpersona/usecase.go`, `port/outbound/repositories.go`.

**Problem:** Only `listpersonas` verified the collector. For writes, Postgres would have raised an FK error (a 500) while the in-memory fake would have accepted the persona, so the result depended on the adapter.

**Decision:** Whether a collector exists is business logic. Mapping the FK error in the Postgres adapter was rejected: it would tie the rule to the schema, cover writes but not lists, and force the fakes to emulate foreign keys.
- Every tenant use case checks the collector through `CollectorRepository` before it uses a store.
- The behaviour of a repository for an unknown collector is not part of its contract.
- When a third use case needs the check, extract it into a shared helper.

## 3. The domain is not free of infrastructure types, and ID types are inconsistent

> **Done (consistency only).** `ItemID` got `ParseItemID` and `ErrInvalidItemID`, and the three generator ports are now one `outbound.IDGenerator`. Deliberately not done: `CollectorID` stays a short, URL-friendly string (a product choice, see the baseline migration), and `uuid.UUID` stays the representation of persona and item ids in the domain, because a hand-written UUID type would cost more than the dependency it removes.

**Where:** `domain/ids.go`, `domain/collector_id.go`, depguard `domain-allowlist`.

**Problem:**
- `PersonaID` and `ItemID` are `uuid.UUID` aliases, so the domain imports `github.com/google/uuid`. It's a small leak, but the allow-list normalises it.
- `CollectorID` is a validated string, `PersonaID` and `ItemID` are UUIDs, and `ItemID` has no `Parse` function.
- Ports have three separate generator interfaces, which `createcollector` re-bundles into `IDGenerator`.

**Fix:**
- Define all ids as `[16]byte`-backed or string-backed types in the domain with `Parse`/`String`. Keep `uuid` in the adapters (idgen, pg).
- Collapse the three generator ports into one `port.IDGenerator`. Use-case-local interfaces stay only where a use case needs a narrower subset.

## 4. `createcollector` holds orchestration that is really domain logic

> **Done (catalog part).** The catalog is code, so the `PersonaTemplateCatalog` port and the `adapters/outbound/templates` adapter are gone. The content is `personatemplate.Predefined()`, and `createcollector.New` takes the templates as plain data. Deliberately not in `domain`, which stays free of `Locale`/`Localized`. `Instantiate` is pure (the caller passes the ids). The collision retry stays in the use case but is isolated in `withFreeCollectorID`, and the persona and item ids are drawn once outside it. Collector ids stay short (a product choice) and 4 characters collide for real, so a retry cannot be dropped.

**Where:** `createcollector/usecase.go`, `personatemplate`, `port/templates.go`.

**Problem:**
- The use case mixes four concerns: id retry loop, template instantiation, a transaction, and persistence order.
- `personatemplate` is a core package that is neither domain nor use case. `PersonaTemplate.Instantiate` takes the `IDs` port, so a domain-ish object does I/O-flavoured work (id generation).
- The retry-on-`ErrCollectorIDTaken` loop is a collision-handling strategy for persistence, sitting in business code.
- The catalog is hard-wired to `templates.New()` in `wire`. The "catalog" is static data, but it sits behind a port.

**Fix:**
- Make `Instantiate` pure: it takes already-generated ids, or returns a persona draft that the use case turns into entities. Alternatively, move it into `domain` as a factory (`domain.NewPersonaFromTemplate`).
- Move the id-collision retry into the use case's helper or into an `IDGenerator` that guarantees uniqueness. Or drop retries by using a UUID/ULID for the collector id, which makes collisions a non-issue.
- Decide whether the template catalog is data (embed it as JSON or YAML, keeping the port) or code. If it stays code, a port is over-engineering, and the catalog can live in `personatemplate` itself.

## 5. HTTP handlers know about domain types and the use-case command

> **Done (ports part).** Inbound ports live in `core/port/inbound`, the former `core/port` is now `core/port/outbound`, and adapters no longer import use-case packages (depguard). The query-side read models are deliberately deferred.

**Where:** `httpadapter/collector/create.go` (imports `usecases/createcollector`), `persona/get.go` (builds responses from `*domain.Persona`).

**Problem:**
- The handler's `Creator` interface is declared in the adapter but refers to `createcollector.Command`, so the adapter depends on a concrete use-case package. That weakens the claim that "there are no inbound port types".
- The response mapping walks the full aggregate. Each accessor (`Items()`, `Spec()`) defensively clones, so serialising a persona does O(n) allocations per access.
- Reads return domain aggregates straight to the adapter. That's fine today, but any read that needs a projection (counts, joins) will have to load full aggregates.

**Fix:**
- Define inbound ports in the core, for example `core/port/inbound` with `CreateCollector`, `GetPersona` and their Command/Result types. Adapters depend on those, use cases implement them, and `wire` assigns them with a compile-time assertion.
- For read-heavy endpoints, add query-side read models, with a separate `Query` interface returning DTOs. Don't force every read through an aggregate (a light CQRS split).

## 6. Error classification is mixed into the domain, and ports leak adapter concerns

> **Done (codes part).** The domain keeps `Kind` and the message. The stable codes and the status mapping moved to `httpx/problemcodes.go`, with tests that every sentinel has a code and that the OpenAPI enum matches. Not done: the explicit three-class `Classify(err)` function and the cleanup of the `CorruptData`/`ErrTenantMismatch` helpers in `port`.

**Where:** `domain/errors.go`, `port.ErrCorruptData`, `ErrTenantMismatch`.

**Problem:**
- `domain.Error` carries `Code` and `ErrorKind`, which are API-contract concepts (stable code strings) inside the domain. The domain now knows what clients switch on.
- Sentinels, `*Error` types and bare `errors.New` values coexist. `ErrDuplicateItemID` is deliberately not an `*Error`, and `CorruptData` deliberately discards the error chain with `err.Error()`. The classification works, but it is implicit.
- Behaviour depends on `errors.As` against a type in the domain, so any adapter or use case that wraps a domain error with `%w` leaks a 4xx.

**Fix:**
- Keep `Kind` (it's semantic), but move `Code` strings and the HTTP status mapping to the inbound adapter, as a table keyed on the sentinel. A new error then gets an HTTP mapping in one place.
- Make the three classes explicit: a validation error (domain), a "not found / conflict" (domain), and a system error (never exposed). A single `Classify(err) Kind` function in the core would be clearer than relying on type assertions.

## 7. The transaction boundary is leaky and not used consistently

> **Done (ownership).** The use case owns the transaction: `DBTX` has no `Begin`, and `SavePersona` is one atomic batch (a plain item insert makes a taken item id a hard error that aborts the batch). Deliberately kept: `Repositories` stays one struct (`postgresRepositories` already builds the pool and the transactional set from the same list), and read-only use cases do not take the `Transactor`.

**Where:** `port.Transactor`, `port.Repositories`, `postgres.Transactor`, `createcollector`.

**Problem:**
- `Repositories` is a struct of every repository, so each new resource widens it, and every use case in a transaction sees everything.
- Only `createcollector` uses transactions. The other use cases receive repos directly, so a use case that later needs atomicity must be rewritten to take the Transactor.
- `SavePersona` opens its own transaction (a savepoint when nested). So the unit-of-work boundary is decided in both the adapter and the use case.

**Fix:**
- Decide on one rule: the use case owns the transaction, and repositories never open one. Write each multi-statement save in the repo as a batch on the given `DBTX`, as `saveItems` already does.
- Generate the transactional `Repositories` from the same wiring list that builds the non-transactional repos. At the moment `postgresRepositories(pool)` and the `repos func(DBTX)` closure risk drifting apart.

## 8. Concurrency and aggregate integrity

**Where:** `SavePersona` doc comment (explicitly acknowledged).

**Problem:**
- Replace-the-whole-aggregate with last-write-wins is a known gap. The comment correctly names the risk (silent deletion of items another editor added).
- The aggregate has no version field, so adding optimistic locking later changes the `Persona` constructor, `RehydratePersona`, the port and the HTTP contract all at once.

**Fix:**
- Add `Version` to the aggregate now. It's cheap while only two read endpoints exist.
- `SavePersona` then does `UPDATE ... WHERE version = $n` and returns `ErrConcurrentModification`. The HTTP side maps it to 409 plus ETag/If-Match.

## 9. Composition root and lifecycle

**Where:** `internal/backend/backend.go`, `Deps.validate`.

**Problem:**
- `wire` builds every use case from `repos` and constructs `templates.New()` and `idgen.Random{}` inline, so tests of the wiring can't substitute them.
- `Deps.validate` uses a hand-maintained map of nil checks (the comment says "a new field needs a line here"). Forgetting it recreates the exact bug it guards against.
- `logging` is imported from `httpx`, so the shared HTTP helper depends on a global slog (`slog.Error` in `WriteJSON`, without a context).

**Fix:**
- Use reflection once (`validate` walks the struct fields) or build `Deps` through a constructor with required parameters, so the compiler catches omissions.
- Inject the logger (or take it from `ctx`) into `httpx`, and use the context-aware logging call everywhere.

## 10. Test architecture

**Where:** `internal/testutil/*`, `fake/*`.

**Problem:** This is mostly a strength: shared contract tests (`porttest`) run against both the fake and Postgres. The weak spots are that the fake implements its own tenant isolation, ordering and error semantics, so the contract tests have to keep catching drift. And `fake/transactor` can't reproduce the rollback semantics of a real DB unless the fake snapshots state.

**Fix:** Check that the fake `Transactor` rolls back on error (snapshot-and-restore) and that `porttest` has a case that fails if it doesn't.

---

## Suggested order

1. **Quick, low risk:** #9 (struct validation, injected logger), #3 (one `IDGenerator`, no `uuid` in the domain), #2 (FK mapping in the repo).
2. **Before authentication lands:** #1 (explicit tenant/principal in use-case input) and #5 (inbound ports). These two change every signature, so they are cheapest now, while only 4 use cases exist.
3. **Before the next write use case:** #7 (transaction ownership), #8 (version field) and #4 (simplify `createcollector`).
4. **Opportunistic:** #6 (move error codes to the adapter) and #10.
