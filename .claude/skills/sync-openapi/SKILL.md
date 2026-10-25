---
name: sync-openapi
description: Update backend/api/openapi.yaml so it matches the inbound HTTP handlers. Use after changing anything under backend/internal/adapters/inbound/httpadapter (routes, request/response structs, status codes, headers), or when the spec may have drifted.
---

# Sync openapi.yaml with the inbound handlers

`backend/api/openapi.yaml` is hand-maintained and is the contract the BFF relies on. The handlers are the truth; make the spec describe them exactly.

## Procedure

1. Read `backend/api/openapi.yaml` and every non-test file under `backend/internal/adapters/inbound/httpadapter/`: `routes.go`, `middleware.go`, `httpx/httpx.go` and each resource package (`collector/`, `persona/`, `health/`, ...).
2. For every `httpx.Route` (its `Pattern` is `"METHOD /path"`; `{name}` segments are path parameters) make sure the spec has:
   - the path and method, a unique `operationId` and the path parameters,
   - the request body schema, taken from the `*Request` struct and its `json` tags (`DisallowUnknownFields` is on, so the schema should not allow extras),
   - one response per status the handler can produce: each `httpx.WriteJSON` status, `400` from `httpx.ParsePathID` and `httpx.Tenant`, and the statuses `httpx.WriteError` maps from domain error kinds (invalid 400, not found 404, conflict 409, anything else 500),
   - response headers the handler sets (`Location`, `X-Request-ID` from the middleware).
3. Build schemas from the response structs: non-pointer fields are `required`, pointer fields are `nullable`, slices are never null when the handler initialises them (check). Use the exact `json` names.
4. Reuse existing `components` and `$ref`s, keep the file's existing style and ordering, and use the error body schema that `httpx.ErrorResponse` defines.
5. Remove spec paths/operations whose route no longer exists.
6. Verify: the YAML parses (`python3 -c "import yaml,sys; yaml.safe_load(open('backend/api/openapi.yaml'))"`), and run `go test ./internal/adapters/inbound/httpadapter/ -run OpenAPI` in `backend/` (compares spec operations with registered routes).
7. Tell the user which operations or schemas changed. Do not touch the handlers to fit the spec unless asked.
