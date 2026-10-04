# tax-return-organizer

>>This repo is still WIP

This repository hold services to help you organize documents for your tax return. It is used by two personas:

1. The collector: defines which documents and information a client must provide
2. The client: uploads the files and enters necessary data

### Personas & tenancy

This is a multi-tenant (SaaS-style) system: there can be many collectors (e.g. different tax advisories), each managing their own set of clients. A client belongs to exactly one collector. Categories and checklist templates are defined by a collector and reused across that collector's own clients; tax years, documents, and checklist item completions each belong to one specific client.

## Usage by collector

- Setup checklist categories and templates describing the necessary data and files (including metadata)
    - Private finances: bank statements, ETFs, savings accounts, ...
    - Work: salary statements, number of homeoffice days, ...
    - Insurance: premium notices, ...
    - Real Estate: bank statements, loan amounts, ...
- Define if a checklist item is required once per tax year or recurring monthly
- Setup deadlines when these files/information need to be provided, per checklist template/item

## Usage by client

- See the files/information necesary to provide for each year
- Upload files and set necessary metadata
- Enter structured data directly on a checklist item (e.g. a count or amount), satisfying that item in place of uploading a file
- get notified about upcoming deadlines
- package/export files and data to send to tax office
    - include files metadata into the export (how needs to be evaluated)


## Architecture

We use a Go backend and a SvelteKit BFF, backed by a PostgreSQL database. They are shipped as three docker containers orchestrated in docker compose. Documents and their metadata as well as standalone data objects are stored in PostgreSQL, not on the filesystem.

### Go backend

- Hexagonal architecture
- storing the collectors models
- Storing clients data and documents (as blobs) and their metadata in PostgreSQL, via a Postgres outbound adapter (kept swappable per hexagonal arch, in case storage moves elsewhere later)
- Exposing REST API for the BFF to call
- Auth is intentionally left out for now — deferred until upload/classify/store/export work end-to-end

### SvelteKit BFF

- collector page
- client page

### Docker compose

- NodeJS container with bff
- Plain container with compiled backend binary
- postgres container, with a named docker volume mounted at `/var/lib/postgresql/data` (Postgres' `PGDATA` dir), so the database (and thus the documents stored as blobs in it) persists across container restarts/recreation — without it, `docker compose down` would wipe all data
- Only the Go backend container talks to postgres directly