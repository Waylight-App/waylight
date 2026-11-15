-- Game deal settings per account (ADR 0018, #294): everyone has their own
-- Steam wishlist. The single row becomes the owner's; accounts without
-- their own settings use the owner's. SQLite cannot drop CHECK (id = 1), so
-- the table is rebuilt; no other table references it. account_id defaults
-- to the owner for exports from before this change.
CREATE TABLE gamedeals_settings_v2 (
    account_id   TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    steam_id     TEXT NOT NULL DEFAULT '',
    country      TEXT NOT NULL,
    min_discount INTEGER NOT NULL,
    epic         INTEGER NOT NULL DEFAULT 0
);
INSERT INTO gamedeals_settings_v2 (account_id, steam_id, country, min_discount, epic)
SELECT 'owner', steam_id, country, min_discount, epic FROM gamedeals_settings;
DROP TABLE gamedeals_settings;
ALTER TABLE gamedeals_settings_v2 RENAME TO gamedeals_settings;
