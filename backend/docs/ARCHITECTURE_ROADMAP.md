# Architecture roadmap

Weaknesses found in an architecture review of the backend (October 2026), while nothing is deployed and changes are still cheap. [ARCHITECTURE.md](ARCHITECTURE.md) describes the current state; this file lists what should change, why, and when. Each entry has a status: *decided* (agreed, waiting for the feature that needs it) or *proposed* (to be decided when the feature comes).

What holds up well and is not revisited: the layering enforced by depguard, tenant isolation through `...Store.For`, the inbound ports, one package per use case, no optimistic locking on personas, and narrowing the `Transactor` per use case.

## High: fix before the related feature is built

### 1. Document content stays out of the aggregate

*Status: decided. When: before the upload use case.*

The domain model put `Content bytes` on `Document`, and the README stored documents as blobs in Postgres. A `ClientTaxYear` is loaded and saved as one aggregate, so every read of a tax year would load all its files, and every save would write them again. The HTTP adapter cannot take uploads either: `httpx.DecodeJSON` accepts only JSON, `MaxBodyBytes` is 1 MiB, and the server's `WriteTimeout` is 15 s (`internal/backend/backend.go`).

Decision:

- An outbound port `DocumentContentStore` with streaming `Put(ctx, io.Reader) (ContentRef, size, sha256, error)` and `Open(ctx, ContentRef) (io.ReadCloser, error)`. The `Document` entity holds only `ContentRef`, `SizeBytes`, `SHA256` and `MimeType`; the content never passes through the domain.
- The first adapter is Postgres, in tables of its own that are never joined into the tax year queries. A separate table alone does not make large files safe: pgx reads and writes a `bytea` value as one buffer, so a single value per file would hold the whole file in memory on every upload and download (and is capped at 1 GB). The content is therefore stored in chunks, `document_contents (id, size_bytes, sha256, wrapped_key, created_at)` and `document_content_chunks (content_id, seq, data)` with about 1 MiB per chunk, so `Put` and `Open` stream with bounded memory. Postgres large objects (`lo_*`) would stream too, but need a transaction per access and their own cleanup (`lo_unlink`), and cannot be protected by a foreign key; chunks are simpler.
- Encrypted at rest with a key from the environment (envelope encryption: a data key per content, wrapped by the master key). Each chunk is encrypted on its own (AES-GCM with its own nonce, authenticated together with `content_id` and `seq`), so encryption streams as well and chunks cannot be swapped or reordered unnoticed. Tax documents fall under professional secrecy (§ 203 StGB), see [AUTHENTICATION.md](AUTHENTICATION.md#dsgvo-points).
- Size limits: at most 25–50 MB per file (exact value to be set with the upload use case) and a limit per tax year. Tax documents are mostly PDFs and photos of a few MB. The total size of the documents is what decides when to move to an object store (S3-compatible, EU-hosted) behind the same port, without touching the core: once backups, the write-ahead log or a restore get too slow because every file lives in the database.
- Content that a document refers to cannot be deleted:
  - Postgres: `documents.content_ref REFERENCES document_contents (id) ON DELETE RESTRICT`, so the database refuses to delete referenced content whatever the code does. The app's database role has no `DELETE` on the content tables; only the role of the cleanup job does.
  - Object store, later: the app's credentials may put and get objects but not delete them; only the cleanup job's credentials may. Versioning is on, so a delete or an overwrite keeps the old version, and Object Lock in governance mode adds a retention period (not compliance mode, see erasure below). A periodic check that every `ContentRef` exists, and occasionally that it still matches its SHA-256, reports content that is missing or changed. Replication to a second EU region covers the loss of a region.
- Erasure (Art. 17 DSGVO, deleting a client) works by crypto-shredding: deleting the wrapped data key makes the content unreadable at once, even where the bytes themselves are locked or still in a backup. Retention periods (e.g. § 147 AO) decide how long a key is kept, not the storage.
- Order: the content is written first, the document metadata second. A failed metadata save leaves unreferenced content, which a periodic cleanup deletes. Content is never referenced before it exists. The cleanup decides from the database, never from a listing of the storage: it deletes only content that no document refers to and that is older than a grace period (e.g. 24 h), so an upload whose metadata is still being saved is not removed.
- Upload and download routes get their own limits: a maximum file size, a request deadline per route (`http.ResponseController.SetReadDeadline`/`SetWriteDeadline`) instead of the global timeouts, and a MIME type sniffed from the content (`http.DetectContentType`) and checked against an allow-list, never the client's `Content-Type`.

### 2. The tax year is not saved by replacing it whole

*Status: proposed. When: with the `ClientTaxYear` repository.*

`SavePersona` replaces the whole persona in one batch, and that is fine for at most 100 items with a single editor. A tax year is different. It holds hundreds of `RequiredItem`s (12 periods × items × instances), and the client and the collector write to it at the same time (upload vs. `Close`, see "Transactions" in [ARCHITECTURE.md](ARCHITECTURE.md#transactions)). Copying the persona pattern would rewrite every item on each upload, and the last save would silently win.

Proposal:

- `client_tax_years.version`, checked and increased by **every** write that touches the year, including child writes (upload, remove document, enter value, set metadata). A mismatch is a domain conflict (409). This is the version check that "Transactions" already asks for, applied to every write path.
- The repository offers targeted writes (`AddDocument`, `EnterValue`, ...) that each change the rows they need plus the version, not a full replace. The aggregate still checks the rules (open year, spec, `Done()`) before the write.
- Read models for the views that span aggregates or need only a projection: the client's item list with its state, and the staff overview "outstanding items per client". Add them as query ports with their own result types (as "Inbound and outbound ports" already allows) instead of loading whole aggregates.

### 3. Collectors are created by the operator

*Status: decided. When: with authentication (step 4 in [AUTHENTICATION.md](AUTHENTICATION.md#implementation-order)).*

`POST /collectors` takes no credentials, so anything that can reach the backend can create tenants. Decision: there is no self-service signup. The operator creates a collector with a `docker create-collector` subcommand (next to `docker migrate`). It runs the `createcollector` use case, which also creates the owner's identity, and the owner is invited by email once invites exist. `POST /collectors` is then removed from the API. The BFF only calls `GET /collectors/{id}`, so it is not affected.

## Medium

### 4. Drop the collector existence checks with authentication

*Status: decided. When: auth step 3, when routes lose `{collectorId}`.*

Every tenant-scoped repository returns `domain.ErrCollectorNotFound` for an unknown collector (see [tenant-scoped-stores.md](tenant-scoped-stores.md)). Reads start `FROM collectors c LEFT JOIN ...`, `SavePersona` probes `collectors` in its batch to tell the two causes of an empty upsert apart, and `TestPersonaRepositoryUnknownCollector` pins it. It exists only because the collector id comes from the URL today. With authentication the actor is built from a membership, which has a foreign key to its collector, and its sessions are deleted with it, so the collector of an actor always exists.

Plan: when routes lose `{collectorId}`, remove the existence checks from the repositories and the contract tests. Reads query the resource table directly; a foreign key violation on a write is then a bug (500), not `ErrCollectorNotFound`. `GetCollector` keeps its not-found for the unscoped `CollectorRepository`.

### 5. Repositories may make their own writes atomic

*Status: proposed. When: together with 4, or with the tax year repository.*

Repositories may not open a transaction (`pgdb.DBTX` has no `Begin`), so a multi-statement save has to be one pgx batch. `SavePersona` therefore queues a guarded CTE upsert per item, `EXISTS` guards on every item statement, and a collector probe. That SQL is hard to review, and the tax year repository would copy it.

Proposal: add `Begin(ctx) (pgx.Tx, error)` to `pgdb.DBTX`. On the pool it starts a transaction, on a `pgx.Tx` it creates a savepoint. A repository can then make its own save atomic, and the use case still owns the outer transaction when there is one. Write the saves set-based: one upsert of all items from `unnest` arrays and one delete of the removed ones. The `Transactor` stays the way to make writes through **several** repositories atomic.

### 6. Validation errors say which field

*Status: proposed. When: before the endpoints that create and edit personas.*

`httpx.WriteError` answers with the sentinel's message only, so the details added with `fmt.Errorf("%w: %q", ...)` are lost. Saving a persona with 30 items answers "item title is required" without saying which item, and it reports only the first problem. A form in the BFF cannot mark the field.

Proposal: a domain validation error that carries a field path (`items[3].title`, `items[0].spec.metadataFields[1].name`) and collects several violations. Each violation keeps its sentinel, so `problemcodes.go` keeps working. The adapter maps it to an `errors` member in the problem details (RFC 9457 allows extensions): `[{"pointer": "/items/3/title", "code": "item_title_required"}]`. The paths are domain terms; the adapter translates them to JSON pointers of its request type.

### 7. Audit log

*Status: proposed. When: decided before tax years.*

The authentication design logs security events. Under DSGVO and § 203 StGB, the reads of staff matter as well: who opened which client's documents. Proposal: an outbound port `AuditLog`, called by the use cases (the core decides what is audited). Writes record their entry in the same transaction. Reads append it separately, so they keep running on the pool without a `Transactor` (see "Transactions" in [ARCHITECTURE.md](ARCHITECTURE.md#transactions)). A failed audit write fails the request.

### 8. Row-level security stays deferred

*Status: proposed, deferred.*

Postgres row-level security (`SET LOCAL app.collector_id`) would add a database-level tenant guard to the stores (see "Known limits" in [tenant-scoped-stores.md](tenant-scoped-stores.md#known-limits)). `SET LOCAL` needs a transaction, and reads deliberately run on the pool. Revisit if the audit log (7) or something else puts reads into transactions anyway.

## Low: cleanups

9. **Problem details everywhere.** `ServeMux` answers unknown routes (404) and wrong methods (405) with `text/plain`. Wrap the mux so these become problem+json too. The request log carries no actor; add the user and collector id once authentication exists (never the email).
10. **`IDGenerator` without errors.** Since Go 1.24 `crypto/rand` does not fail (it crashes the program instead), so the error results of `NewCollectorID`, `NewPersonaID` and `NewItemID` and their handling in `createcollector.newPersonaIDs` are dead code.
11. **Consistency.** `ErrTooManyItems` is a conflict (409), the other limits are invalid input (400): make it `invalid`. `createcollector` returns its errors without the `create collector:` prefix that "Request flow" in [ARCHITECTURE.md](ARCHITECTURE.md#request-flow) asks for.
12. **`Clock` port.** Tax years (`ClosedAt`) and sessions (expiry) need the current time. Add `outbound.Clock` with the first of them, as planned in [AUTHENTICATION.md](AUTHENTICATION.md#ports-and-adapters).
13. **Error registry.** `domain.Errors()` is a mutable, package-level registry in production code that only `TestEveryDomainErrorHasAProblemCode` uses. Acceptable while it stays that small; noted so it does not grow into more.
14. **Locale-aware sorting.** `ListPersonas` sorts names bytewise (`COLLATE "C"` in `postgres/persona`), which puts "Zeta" before "alpha" and umlauts after "z", wrong for German users. Sort with an ICU collation (`und-x-icu`, or per collector locale) and pin the order in the contract tests.
