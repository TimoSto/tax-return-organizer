# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-tenant (SaaS-style) tool for organizing tax return documents. A *Collector* (e.g. a tax advisory) defines Personas → required documents/data; *Clients* upload files and enter data per tax year. The README holds the use cases and the full domain model (mermaid diagram). The project is WIP: only the Collector create/get slice exists so far, and there is no authentication.

Monorepo with two independent projects (each has its own CI workflow, triggered by path):

- `backend/` — Go REST API (Postgres). Contract: `backend/api/openapi.yaml` (hand-maintained; update it when endpoints change).
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
- `cmd/docker` is the production entrypoint (env: `DATABASE_URL` required, `HTTP_ADDR` default `:8080`). `cmd/e2e` (build tag `e2e`: `go run -tags e2e ./cmd/e2e`) starts a throwaway Postgres container, seeds a collector with id `aaaa`, and serves the API for the BFF's Playwright e2e tests.

## Backend architecture (hexagonal)

- `internal/core/domain` — entities and validation (`Collector`, `CollectorID`, `Email`), plus sentinel errors (`ErrCollectorNotFound`, `ErrCollectorIDTaken`). Collector IDs are short strings, not UUIDs (see migration `0002`), unlike the README diagram.
- `internal/ports` — `inbound.go` (use-case interfaces, e.g. `CreateCollectorPort`) and `outbound.go` (repository interfaces). `ports/porttest` holds shared contract tests that every repository implementation (fake and postgres) runs.
- `internal/core/usecases/<name>` — one package per use case (`New(repo).Execute`).
- `internal/adapters/inbound/httpadapter` — routes, middleware, per-resource subpackages; `httpx` has shared HTTP helpers. Handlers depend on use-case function values via `Deps`.
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
npx playwright test   # e2e (npm run test:e2e also runs `playwright install`)
npx vitest --run src/lib/vitest-examples/greet.spec.ts   # single unit test file
```

CI (Node 25) runs lint, check, build, unit tests and Playwright e2e as separate jobs.
