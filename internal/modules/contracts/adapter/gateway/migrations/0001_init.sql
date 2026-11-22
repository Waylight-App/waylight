-- Contracts with cancellation deadlines (#687). Personal content (ADR 0018):
-- private for members and admins by default, shareable like a note.
CREATE TABLE contracts_items (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    provider     TEXT NOT NULL DEFAULT '',
    cost_cents   INTEGER NOT NULL DEFAULT 0,
    period       TEXT NOT NULL,             -- monthly, quarterly, yearly
    term_end     TEXT NOT NULL,             -- YYYY-MM-DD as entered
    notice_days  INTEGER NOT NULL DEFAULT 0,
    renew_months INTEGER NOT NULL DEFAULT 0,
    remind_days  INTEGER NOT NULL DEFAULT 14,
    note         TEXT NOT NULL DEFAULT '',
    reminded_for TEXT NOT NULL DEFAULT '',  -- deadline of the last reminder
    owner_id     TEXT NOT NULL,
    visibility   TEXT NOT NULL DEFAULT 'private',
    created_at   TEXT NOT NULL
);
CREATE INDEX contracts_items_owner ON contracts_items (owner_id);
