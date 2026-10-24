# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-tenant (SaaS-style) tool for organizing tax return documents. A *Collector* (e.g. a tax advisory) defines Personas → required documents/data; *Clients* upload files and enter data per tax year. The README holds the use cases and the full domain model (mermaid diagram). The project is WIP: only the Collector create/get slice exists so far, and there is no authentication.

Monorepo with two independent projects (each has its own CI workflow, triggered by path):

- `backend/` — Go REST API (Postgres). Contract: `backend/api/openapi.yaml` (hand-maintained; update it when endpoints change, using the `sync-openapi` skill — a hook reminds Claude after handler edits, and `TestOpenAPI_Routes` fails if routes and spec diverge).
- `bff/` — SvelteKit (Svelte, adapter-node) backend-for-frontend with Paraglide i18n (`messages/{de,en}.json`). Still mostly scaffold/demo code.
- `docker-compose.yml` — local Postgres only (db/user/password all `tax`, port 5432).

## Backend commands (run in `backend/`)

```sh
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0 run ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0 run --build-tags e2e ./cmd/...   # lint e2e entrypoint too
go test -race -count=1 ./...
go test ./internal/core/usecases/createcollector -run TestName     # single test
go build -o bin/docker ./cmd/docker
```

- Postgres adapter tests start their own container via testcontainers (`internal/testutil/pgtest`), so **Docker must be running**.
- `cmd/docker` is the production entrypoint (env: `DATABASE_URL` required, `HTTP_ADDR` default `:8080`). `cmd/e2e` (build tag `e2e`: `go run -tags e2e ./cmd/e2e`) starts a throwaway Postgres container, seeds a collector with id `aaaa`, and serves the API for end-to-end tests against the real backend (to live in a separate e2e package; the BFF's own Playwright tests mock the backend).

## Backend architecture (hexagonal)

- `internal/core/domain` — entities and validation (`Collector`, `CollectorID`, `Email`), plus sentinel errors (`ErrCollectorNotFound`, `ErrCollectorIDTaken`). Collector IDs are short strings, not UUIDs (see the baseline migration), unlike the README diagram.
- `internal/core/port` — outbound ports (repository interfaces, `Transactor`) owned by the core. **Tenant-owned resources (everything a collector owns, e.g. personas) are exposed as a `...Store` whose `For(collectorID)` returns a repository bound to that collector, so repository methods never take a collector id.** Only `CollectorRepository` (the tenant root) is unscoped. Follow the checklist in `backend/docs/tenant-scoped-stores.md` when adding a resource. `port/porttest` holds shared contract tests that every repository implementation (fake and postgres) runs. There are no inbound port types: `httpadapter` resource packages declare the small interfaces (e.g. `collector.Creator`) the use cases must satisfy, and `backend.wire` assigns the use cases to them.
- `internal/core/usecases/<name>` — one package per use case (`New(repo).Execute`).
- `internal/adapters/inbound/httpadapter` — routes, middleware, per-resource subpackages; `httpx` has shared HTTP helpers. Handlers depend on use cases via the interfaces in `Deps`.
- `internal/adapters/outbound/{postgres,fake}/<resource>` — repository implementations in per-resource subpackages. Postgres migrations are embedded SQL in `postgres/migrations` (numbered `NNNN_name.up|down.sql`) and applied by `postgres.Migrate` on startup.
- `internal/backend` — composition root: `wire()` builds use cases on repositories and returns the HTTP handler; `Run` migrates, connects, serves, and shuts down gracefully. New use cases must be wired here and added to `httpadapter.Deps`.

## BFF commands (run in `bff/`)

```sh
npm ci
npm run dev
npm run lint          # prettier --check + eslint
npm run format
npm run check         # paraglide compile + svelte-check
npm run build
npm run test:unit -- --run
npx playwright test   # e2e (npm run test:e2e also runs `playwright install`); mocks the backend inside the test files (server-side fetches, so not via page.route)
npx vitest --run src/lib/vitest-examples/greet.spec.ts   # single unit test file
```

CI (Node 25) runs lint, check, build, unit tests and Playwright e2e as separate jobs.
