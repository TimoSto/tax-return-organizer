# tax-return-organizer

>>This repo is still WIP

This repository hold services to help you organize documents for your tax return. Its primary purpose is to save documents in a folder structure so you can save them before sending them to your tax office.

## Planned functionalities

- [ ] upload documents
- [ ] classify the documents (e.g. payroll, invoices, bank statements, ...)
- [ ] create folders and store documents in them
- [ ] export in a way to send to taxoffice

## Architecture

We use a Go backend and a SvelteKit BFF, backed by a PostgreSQL database. They are shipped as three docker containers orchestrated in docker compose. Documents and their metadata are stored in PostgreSQL, not on the filesystem.

### Go backend

- Hexagonal architecture
- Storing documents (as blobs) and their metadata in PostgreSQL, via a Postgres outbound adapter (kept swappable per hexagonal arch, in case storage moves elsewhere later)
- storing list of necessary things
    - with options like "for each month"
- Exposing REST API for the BFF to call
- Auth is intentionally left out for now — deferred until upload/classify/store/export work end-to-end

### SvelteKit BFF

- providing download links for files and folders
- MD3 usage and layout
    - Bar with a select for the year
    - Rail with the follwing elements
        - overview
            - list of necessary things
            - edit list of necessary things
        - tree view of documents
            - expandables for folders
            - upload button on each folder level
            - edit meta data of selected file in supporing pane
        - category A, B, C
            - list of documents with this category

### Docker compose

- NodeJS container with bff
- Plain container with compiled backend binary
- postgres container, with a named docker volume mounted at `/var/lib/postgresql/data` (Postgres' `PGDATA` dir), so the database (and thus the documents stored as blobs in it) persists across container restarts/recreation — without it, `docker compose down` would wipe all data
- Only the Go backend container talks to postgres directly

### Data model (PostgreSQL)

Documents are stored as blobs directly in PostgreSQL, alongside their metadata, so everything is queryable from one place instead of juggling a filesystem layout plus sidecar metadata files.

```sql
-- Tax year the document/checklist belongs to
CREATE TABLE tax_years (
    id   SERIAL PRIMARY KEY,
    year INT NOT NULL UNIQUE
);

-- Classification categories (payroll, invoices, bank statements, ...)
CREATE TABLE categories (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

-- Documents: blob + metadata together
CREATE TABLE documents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_year_id       INT NOT NULL REFERENCES tax_years(id),
    category_id       INT REFERENCES categories(id), -- NULL = unclassified
    original_filename TEXT NOT NULL,
    mime_type         TEXT NOT NULL,
    size_bytes        BIGINT NOT NULL,
    content           BYTEA NOT NULL,      -- the actual file
    properties        JSONB NOT NULL DEFAULT '{}', -- freeform metadata (e.g. employer, account no.)
    uploaded_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON documents (tax_year_id);
CREATE INDEX ON documents (category_id);
CREATE INDEX ON documents USING GIN (properties);

-- "List of necessary things" checklist, with recurrence support (e.g. "for each month")
CREATE TABLE checklist_templates (
    id         SERIAL PRIMARY KEY,
    tax_year_id INT NOT NULL REFERENCES tax_years(id),
    title      TEXT NOT NULL,
    recurrence TEXT NOT NULL CHECK (recurrence IN ('once', 'monthly'))
);

CREATE TABLE checklist_items (
    id           SERIAL PRIMARY KEY,
    template_id  INT NOT NULL REFERENCES checklist_templates(id) ON DELETE CASCADE,
    month        INT CHECK (month BETWEEN 1 AND 12), -- NULL for 'once' templates
    document_id  UUID REFERENCES documents(id),        -- linked once satisfied
    done         BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (template_id, month)
);
```

Notes:
- `BYTEA` is used instead of Large Objects for simplicity — expected documents are small (scanned PDFs/images), not multi-GB files.
- `properties` is `JSONB` because metadata fields vary by category (payroll vs. bank statement vs. invoice) rather than a fixed column set; a GIN index keeps it queryable.
- The document tree/folder view in the UI is derived from `tax_year` → `category`, not real filesystem folders — no separate folders table is needed.

