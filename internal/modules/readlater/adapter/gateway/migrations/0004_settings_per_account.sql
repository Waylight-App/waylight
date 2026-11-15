-- Auto-archiving is a personal setting (ADR 0018, #294); the single row
-- becomes the owner's. Accounts without their own setting use the owner's.
-- SQLite cannot drop CHECK (id = 1), so the table is rebuilt; no other
-- table references it. account_id defaults to the owner for older exports.
CREATE TABLE readlater_settings_v2 (
    account_id   TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    archive_days INTEGER NOT NULL DEFAULT 0
);
INSERT INTO readlater_settings_v2 (account_id, archive_days) SELECT 'owner', archive_days FROM readlater_settings;
DROP TABLE readlater_settings;
ALTER TABLE readlater_settings_v2 RENAME TO readlater_settings;
