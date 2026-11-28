-- #866: the reminder the evening before a collection; off by default at
-- 19:00 (minute of the day). sent_for is the collection day (YYYY-MM-DD) the
-- last reminder was sent for, so each day is reminded once.
CREATE TABLE waste_reminder (
    id       INTEGER PRIMARY KEY CHECK (id = 1),
    enabled  INTEGER NOT NULL DEFAULT 0,
    minute   INTEGER NOT NULL DEFAULT 1140,
    sent_for TEXT NOT NULL DEFAULT ''
);
