-- Device token bindings of the push relay (ADR 0025): the SHA-256 hash of a
-- token, the instance key that sent to it first and the last use.
CREATE TABLE bindings (
    token_hash BLOB PRIMARY KEY,
    instance   TEXT NOT NULL,
    last_used  INTEGER NOT NULL
) STRICT;

CREATE INDEX bindings_last_used ON bindings (last_used);
