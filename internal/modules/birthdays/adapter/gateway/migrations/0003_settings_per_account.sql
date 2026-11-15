-- Birthday settings per account (ADR 0018, #294): the address book and its
-- app password are personal, with no fallback to the owner's. The single
-- row becomes the owner's. Reminders go to the instance's channels and stay
-- the owner's. SQLite cannot drop CHECK (id = 1), so the table is rebuilt;
-- no other table references it. account_id defaults to the owner for
-- exports from before this change.
CREATE TABLE birthday_settings_v2 (
    account_id         TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    remind_enabled     INTEGER NOT NULL DEFAULT 0,
    remind_at          TEXT NOT NULL DEFAULT '08:00',
    last_reminded      TEXT NOT NULL DEFAULT '',
    carddav_url        TEXT NOT NULL DEFAULT '',
    username           TEXT NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL DEFAULT ''
);
INSERT INTO birthday_settings_v2 (account_id, remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted)
SELECT 'owner', remind_enabled, remind_at, last_reminded, carddav_url, username, password_encrypted FROM birthday_settings;
DROP TABLE birthday_settings;
ALTER TABLE birthday_settings_v2 RENAME TO birthday_settings;
