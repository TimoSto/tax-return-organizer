-- Collector ids become short, URL-friendly codes. Existing rows get a fresh random code;
-- hex digits are mapped into the code alphabet (no 0, 1, i, l, o).
ALTER TABLE collectors ALTER COLUMN id TYPE text USING translate(substr(md5(id::text), 1, 4), '01', 'xy');
ALTER TABLE collectors ADD CONSTRAINT collectors_id_format CHECK (id ~ '^[2-9a-hj-km-np-z]{4}$');
