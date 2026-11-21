-- The running or paused focus timer, one per account (#631). Durations are
-- milliseconds; ends_at is NULL while paused.
CREATE TABLE focus_timers (
    owner_id    TEXT PRIMARY KEY,
    phase       TEXT NOT NULL,
    done        INTEGER NOT NULL,
    ends_at     TEXT,
    left_ms     INTEGER NOT NULL,
    total_ms    INTEGER NOT NULL,
    work_min    INTEGER NOT NULL,
    short_min   INTEGER NOT NULL,
    long_min    INTEGER NOT NULL,
    long_every  INTEGER NOT NULL
);
