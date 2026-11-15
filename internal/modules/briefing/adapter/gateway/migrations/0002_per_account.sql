-- The briefing is per account (ADR 0018, #294): sections and location are
-- personal. The single row becomes the owner's; accounts without their own
-- row use the owner's sections and location. The scheduled push still goes
-- to the instance's channels, so only the owner's row sends one for now.
-- SQLite cannot drop CHECK (id = 1), so the table is rebuilt; no other
-- table references it. account_id defaults to the owner for older exports.
CREATE TABLE briefing_settings_v2 (
    account_id  TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    layout      TEXT NOT NULL DEFAULT '[]',
    push_on     INTEGER NOT NULL DEFAULT 0,
    push_minute INTEGER NOT NULL DEFAULT 420,
    push_days   TEXT NOT NULL DEFAULT '',
    skip_empty  INTEGER NOT NULL DEFAULT 1,
    lat         REAL NOT NULL DEFAULT 0,
    lon         REAL NOT NULL DEFAULT 0,
    last_sent   TEXT
);
INSERT INTO briefing_settings_v2 (account_id, layout, push_on, push_minute, push_days, skip_empty, lat, lon, last_sent)
SELECT 'owner', layout, push_on, push_minute, push_days, skip_empty, lat, lon, last_sent FROM briefing_settings;
DROP TABLE briefing_settings;
ALTER TABLE briefing_settings_v2 RENAME TO briefing_settings;
