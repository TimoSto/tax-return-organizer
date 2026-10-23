ALTER TABLE item_definitions
    ADD CONSTRAINT item_definitions_kind_check CHECK (kind IN ('document', 'data')),
    ADD CONSTRAINT item_definitions_recurrence_check
        CHECK (recurrence IN ('monthly', 'quarterly', 'semiannually', 'annually'));
