-- The connected shopping list is a personal setting (ADR 0018, #294); the
-- single row becomes the owner's. Accounts without their own list use the
-- owner's, so a household shares one list by default. SQLite cannot drop
-- CHECK (id = 1), so the table is rebuilt; no other table references it.
-- account_id defaults to the owner for exports from before this change.
CREATE TABLE shopping_settings_v2 (
    account_id         TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    list_url           TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
INSERT INTO shopping_settings_v2 (account_id, list_url, username, password_encrypted)
SELECT 'owner', list_url, username, password_encrypted FROM shopping_settings;
DROP TABLE shopping_settings;
ALTER TABLE shopping_settings_v2 RENAME TO shopping_settings;
