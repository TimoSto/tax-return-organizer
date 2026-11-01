CREATE TABLE tax_practices (
    id    uuid PRIMARY KEY,
    name  text NOT NULL,
    email text NOT NULL
);

CREATE TABLE personas (
    id           uuid PRIMARY KEY,
    tax_practice_id uuid NOT NULL REFERENCES tax_practices (id),
    name         text NOT NULL,
    description  text NOT NULL DEFAULT '',
    -- The id alone is unique already. This is the target of the composite foreign keys of child tables, which tie
    -- a child to the tax practice of its persona.
    CONSTRAINT personas_id_tax_practice_id_key UNIQUE (id, tax_practice_id)
);

CREATE INDEX personas_tax_practice_id_idx ON personas (tax_practice_id);

-- Item ids are unique per persona, not globally, so an id of another persona (or tenant) can never clash.
CREATE TABLE item_definitions (
    id            uuid NOT NULL,
    persona_id    uuid NOT NULL,
    -- Every tenant table carries its tax practice. The foreign key below makes it the tax practice of the persona,
    -- so the database refuses an item under a persona of another tax practice, whatever the statement says.
    tax_practice_id uuid NOT NULL,
    position      int  NOT NULL,
    title         text NOT NULL,
    -- Which kinds and recurrences exist is domain logic: the repository validates them when it reads a row.
    kind          text NOT NULL,
    recurrence    text NOT NULL,
    document_spec jsonb,
    value_spec    jsonb,
    -- Instance fields make an item repeatable (see ItemDefinition in the domain). An empty array is a single item.
    instance_fields jsonb NOT NULL DEFAULT '[]',
    -- A document item carries a document spec, a data item a value spec, never both.
    CONSTRAINT item_definitions_spec_matches_kind CHECK (
        (kind = 'document' AND document_spec IS NOT NULL AND value_spec IS NULL)
        OR (kind = 'data' AND value_spec IS NOT NULL AND document_spec IS NULL)
    ),
    -- A JSON null literal is not SQL NULL, so it would pass the check above but is no spec.
    CONSTRAINT item_definitions_document_spec_object
        CHECK (document_spec IS NULL OR jsonb_typeof(document_spec) = 'object'),
    CONSTRAINT item_definitions_value_spec_object
        CHECK (value_spec IS NULL OR jsonb_typeof(value_spec) = 'object'),
    CONSTRAINT item_definitions_instance_fields_array
        CHECK (jsonb_typeof(instance_fields) = 'array'),
    CONSTRAINT item_definitions_persona_fkey FOREIGN KEY (persona_id, tax_practice_id)
        REFERENCES personas (id, tax_practice_id) ON DELETE CASCADE,
    PRIMARY KEY (persona_id, id)
);

CREATE INDEX item_definitions_persona_id_idx ON item_definitions (persona_id, position);
