-- People of the phone list widget (#300), maintained by admins.
CREATE TABLE phonelist_entries (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    role       TEXT NOT NULL DEFAULT '',
    phone      TEXT NOT NULL DEFAULT '',
    email      TEXT NOT NULL DEFAULT '',
    team       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
