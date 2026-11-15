-- The tour state is per account (ADR 0018, #294): a new member gets the
-- tour even when the owner finished it. The single row becomes the owner's.
-- SQLite cannot drop CHECK (id = 1), so the table is rebuilt; no other
-- table references it. account_id defaults to the owner for older exports.
CREATE TABLE onboarding_state_v2 (
    account_id   TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    completed_at TEXT
);
INSERT INTO onboarding_state_v2 (account_id, completed_at) SELECT 'owner', completed_at FROM onboarding_state;
DROP TABLE onboarding_state;
ALTER TABLE onboarding_state_v2 RENAME TO onboarding_state;
