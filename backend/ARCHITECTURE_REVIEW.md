# Architecture review of `backend/` (hexagonal Go): open items

Only the findings that are still open are listed. The rest of the original review (tenant-scoped stores, depguard rules, id generation, problem+json errors, route tenant metadata, domain invariants, graceful shutdown, hygiene) has been resolved.

## High
1. **The 4-char collector id is the only access control, and it is enumerable.**
   `domain/collector_id.go` (4 chars × 31-symbol alphabet ≈ 923k ids) + `tenantFromPath` in `httpadapter/routes.go` (tenant = URL segment). `GET /collectors/{id}` returns the collector's **email** (PII) and 404-vs-200 is an existence oracle, so the whole tenant space can be scraped in ~1M requests. `POST /collectors` is unauthenticated and unthrottled and writes 1 + 2 personas + 5 items per call (write amplification, id-space exhaustion; birthday collisions get likely at ~1k tenants and `maxIDAttempts = 5` then returns a 500). "No auth yet" is documented, but this is deployed via `cmd/docker`, so treat the short id as public. Fix: do not ship publicly before auth; meanwhile make the id longer/non-guessable, or at least rate-limit + drop email from `GET`.

## Low / hygiene
2. **No pagination / read model.** Pass-through use cases (`getcollector`, `getpersona`) are anemic; fine for hexagonal consistency, but the read path returns full aggregates (all items) for list endpoints with no pagination. Plan a read-model/query port before data grows.
3. **Test code in the production tree.** Fakes (`adapters/outbound/fake`) and `postgrestest` live in the production tree as non-`_test` packages; harmless to the binary but conceptually test code (`internal/testutil/…` would be clearer). The fake transactor doesn't roll back, so atomicity is only covered by Postgres tests (documented).
4. **No metrics or tracing.** Migrations, pool settings and log level are configurable by now; observability beyond logs is still missing.
5. **Response writing edge cases.**
   - `httpx.WriteJSON` logs with `slog.Error` (no ctx), so the request id is lost on write failures.
   - `recoverPanic` writes a 500 unconditionally, even if the handler already started the response.
6. **Mux-generated 404/405 are plain text.** They bypass `application/problem+json`, so clients see two error formats.
