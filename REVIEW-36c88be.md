# Review of 36c88be: get persona use case with item definitions

Hexagonal slice is clean; main risks are in `SavePersona` and fake/Postgres drift.

## Findings (most important first)

1. **Fake and Postgres diverge on item ids.**
   - Postgres rejects an item id that belongs to another persona (`saveItems`, `RowsAffected()==0`); the fake accepts it (one object per persona).
   - The `porttest` contract has no case for it, so tests can't catch it.
   - The Postgres error is a plain `fmt.Errorf` → HTTP 500. Should be a domain sentinel (e.g. `ErrItemIDTaken`), documented on the port, covered by a contract test, and enforced by the fake via a global item-id index.

2. **`GetPersona` reads in two unsynchronized queries** (`postgres/persona/persona.go`).
   - A concurrent `SavePersona` can commit in between → old name with new items.
   - Fix: one `REPEATABLE READ` read-only transaction, or fetch items via `jsonb_agg` in a single query.

3. **Missing collector surfaces as a 500.**
   - Saving a persona for a nonexistent collector hits FK violation `23503`, not mapped to `ErrCollectorNotFound`.
   - The fake accepts it silently (second fake/Postgres divergence). Map the error and add a contract case.

4. **Lost updates (design risk).**
   - `SavePersona` replaces the whole aggregate; read-modify-write use cases overwrite each other silently.
   - Add the planned version column + 409 before the first write use case, so the `SavePersona` signature/contract doesn't change later.

5. **`id::text <> ALL($2)` in the delete** can't use the PK index.
   - Use `[]uuid.UUID` and `id <> ALL($2::uuid[])`. Per-item `Exec` loop = one round trip each; consider `pgx.Batch`.

6. **Aggregate encapsulation is leaky.**
   - `Persona.Items` is exported (hard to avoid; needed for reconstitution).
   - `AddItem` returns a pointer to a local copy, not the slice element; mutating it silently does nothing. Return a value or the id.
   - `UpdateItem` generates a throwaway UUID and overwrites it; consider an internal constructor taking the id.

7. **Smaller points**
   - `toDomain` turns empty `metadataFields: []` into `nil` (nil-vs-empty subtlety).
   - API naming: `spec` (metadata field) vs `valueSpec` (item); pick one before clients depend on it.
   - `wire(log, collectors, personas)` grows a parameter per repository; use a `repositories` struct.
   - No authorization yet; collector ids are guessable (known WIP).
   - `SavePersona` has no production caller; endpoint only 404s outside tests (documented).
