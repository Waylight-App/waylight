-- Live presence (#346): at most one status per account, for one day.
-- Statuses of earlier days are deleted by a daily job; no history is kept.
CREATE TABLE presence_status (
    account_id TEXT PRIMARY KEY,
    status     TEXT NOT NULL,
    day        TEXT NOT NULL,
    return_on  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX presence_status_day ON presence_status (day);

-- The organisation-wide switch; no row means on.
CREATE TABLE presence_settings (
    id      INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER NOT NULL
);
