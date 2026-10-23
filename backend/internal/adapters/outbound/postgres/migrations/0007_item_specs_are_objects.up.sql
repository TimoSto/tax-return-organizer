-- A JSON null literal is not SQL NULL, so it would pass the "spec matches kind" check but is no spec.
ALTER TABLE item_definitions
    ADD CONSTRAINT item_definitions_document_spec_object
        CHECK (document_spec IS NULL OR jsonb_typeof(document_spec) = 'object'),
    ADD CONSTRAINT item_definitions_value_spec_object
        CHECK (value_spec IS NULL OR jsonb_typeof(value_spec) = 'object');
