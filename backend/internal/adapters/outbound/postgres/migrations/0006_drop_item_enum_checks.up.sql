-- Which kinds and recurrences exist is domain logic. The repository validates them when it reads a row.
ALTER TABLE item_definitions
    DROP CONSTRAINT item_definitions_kind_check,
    DROP CONSTRAINT item_definitions_recurrence_check;
