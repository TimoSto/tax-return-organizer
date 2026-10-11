ALTER TABLE collectors DROP CONSTRAINT collectors_id_format;
ALTER TABLE collectors ALTER COLUMN id TYPE uuid USING gen_random_uuid();
