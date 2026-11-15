-- The credits goal is personal (ADR 0018, #294); the single row becomes the
-- owner's. SQLite cannot drop CHECK (id = 1), so the table is rebuilt; no
-- other table references it. account_id defaults to the owner, so an export
-- from before this change imports as the owner's row.
CREATE TABLE study_settings_v2 (
    account_id TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    total_ects REAL NOT NULL
);
INSERT INTO study_settings_v2 (account_id, total_ects) SELECT 'owner', total_ects FROM study_settings;
DROP TABLE study_settings;
ALTER TABLE study_settings_v2 RENAME TO study_settings;
