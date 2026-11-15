-- Calendars per account (ADR 0018, #294): CalDAV settings, subscriptions
-- and connected accounts are personal. Everything existing becomes the
-- owner's. There is no fallback to the owner's calendars for members (they
-- are private); guests and kiosk displays keep seeing the owner's.
--
-- calendar_settings: SQLite cannot drop CHECK (id = 1), so it is rebuilt;
-- no other table references it. account_id defaults to the owner for
-- exports from before this change.
CREATE TABLE calendar_settings_v2 (
    account_id         TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    calendar_urls      TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
INSERT INTO calendar_settings_v2 (account_id, calendar_urls, username, password_encrypted)
SELECT 'owner', calendar_urls, username, password_encrypted FROM calendar_settings;
DROP TABLE calendar_settings;
ALTER TABLE calendar_settings_v2 RENAME TO calendar_settings;

ALTER TABLE calendar_subscriptions ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE calendar_subscriptions ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private';
CREATE INDEX calendar_subscriptions_owner ON calendar_subscriptions (owner_id);
ALTER TABLE calendar_accounts ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE calendar_accounts ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private';
CREATE INDEX calendar_accounts_owner ON calendar_accounts (owner_id);
