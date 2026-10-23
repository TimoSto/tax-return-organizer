CREATE TABLE item_definitions (
    id            uuid PRIMARY KEY,
    persona_id    uuid NOT NULL REFERENCES personas (id) ON DELETE CASCADE,
    position      int  NOT NULL,
    title         text NOT NULL,
    kind          text NOT NULL CHECK (kind IN ('document', 'data')),
    recurrence    text NOT NULL CHECK (recurrence IN ('monthly', 'quarterly', 'semiannually', 'annually')),
    document_spec jsonb,
    value_spec    jsonb,
    -- A document item carries a document spec, a data item a value spec, never both.
    CONSTRAINT item_definitions_spec_matches_kind CHECK (
        (kind = 'document' AND document_spec IS NOT NULL AND value_spec IS NULL)
        OR (kind = 'data' AND value_spec IS NOT NULL AND document_spec IS NULL)
    )
);

CREATE INDEX item_definitions_persona_id_idx ON item_definitions (persona_id, position);
