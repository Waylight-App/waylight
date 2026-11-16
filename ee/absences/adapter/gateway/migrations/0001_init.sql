-- Absences (#299): one person away on whole days, from_day..to_day included.
-- Days are 'YYYY-MM-DD' text, so comparing them as strings is comparing days.
CREATE TABLE absences (
    id         TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    from_day   TEXT NOT NULL,
    to_day     TEXT NOT NULL,
    kind       TEXT NOT NULL,
    note       TEXT NOT NULL DEFAULT '',
    private    INTEGER NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX absences_days ON absences (to_day, from_day);
CREATE INDEX absences_account ON absences (account_id);
