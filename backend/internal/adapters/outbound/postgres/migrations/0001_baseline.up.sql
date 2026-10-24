-- Collector ids are short, URL-friendly codes (no 0, 1, i, l, o).
CREATE TABLE collectors (
    id    text PRIMARY KEY,
    name  text NOT NULL,
    email text NOT NULL,
    CONSTRAINT collectors_id_format CHECK (id ~ '^[2-9a-hj-km-np-z]{4}$')
);

CREATE TABLE personas (
    id           uuid PRIMARY KEY,
    collector_id text NOT NULL REFERENCES collectors (id),
    name         text NOT NULL,
    description  text NOT NULL DEFAULT ''
);

CREATE INDEX personas_collector_id_idx ON personas (collector_id);

CREATE TABLE item_definitions (
    id            uuid PRIMARY KEY,
    persona_id    uuid NOT NULL REFERENCES personas (id) ON DELETE CASCADE,
    position      int  NOT NULL,
    title         text NOT NULL,
    -- Which kinds and recurrences exist is domain logic: the repository validates them when it reads a row.
    kind          text NOT NULL,
    recurrence    text NOT NULL,
    document_spec jsonb,
    value_spec    jsonb,
    -- A document item carries a document spec, a data item a value spec, never both.
    CONSTRAINT item_definitions_spec_matches_kind CHECK (
        (kind = 'document' AND document_spec IS NOT NULL AND value_spec IS NULL)
        OR (kind = 'data' AND value_spec IS NOT NULL AND document_spec IS NULL)
    ),
    -- A JSON null literal is not SQL NULL, so it would pass the check above but is no spec.
    CONSTRAINT item_definitions_document_spec_object
        CHECK (document_spec IS NULL OR jsonb_typeof(document_spec) = 'object'),
    CONSTRAINT item_definitions_value_spec_object
        CHECK (value_spec IS NULL OR jsonb_typeof(value_spec) = 'object')
);

CREATE INDEX item_definitions_persona_id_idx ON item_definitions (persona_id, position);
