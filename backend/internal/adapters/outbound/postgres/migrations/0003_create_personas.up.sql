CREATE TABLE personas (
    id           uuid PRIMARY KEY,
    collector_id text NOT NULL REFERENCES collectors (id),
    name         text NOT NULL,
    description  text NOT NULL DEFAULT ''
);

CREATE INDEX personas_collector_id_idx ON personas (collector_id);
