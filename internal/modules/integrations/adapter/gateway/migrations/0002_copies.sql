-- Module settings copied into connections (#610): one row per source, e.g.
-- "immich" or "arr:<instance ID>", written in the same transaction as the
-- copy once it has been read back and compared. A source with a row is
-- never copied again, also when its connection was removed later. The
-- module settings themselves stay untouched.
CREATE TABLE integrations_copies (
    source        TEXT PRIMARY KEY,
    connection_id TEXT NOT NULL,
    copied_at     TEXT NOT NULL
);
