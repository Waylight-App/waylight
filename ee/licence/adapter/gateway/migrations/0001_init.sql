-- The licence key entered in the settings (#307). One row at most; a key set
-- with LICENCE_KEY is never stored.
CREATE TABLE licence_key (
    id         INTEGER PRIMARY KEY CHECK (id = 1),
    key        TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
