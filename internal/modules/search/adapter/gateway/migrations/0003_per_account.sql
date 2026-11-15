-- Search settings and shortcuts per account (ADR 0018, #294). The single
-- settings row and all shortcuts become the owner's. SQLite cannot drop
-- CHECK (id = 1) or change a primary key, so both tables are rebuilt; no
-- other table references them. account_id defaults to the owner, so an
-- export from before this change imports as the owner's settings.
CREATE TABLE search_settings_v2 (
    account_id       TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    default_provider TEXT NOT NULL,
    open_in_new_tab  INTEGER NOT NULL DEFAULT 0,
    capture_prefixes TEXT NOT NULL DEFAULT ''
);
INSERT INTO search_settings_v2 (account_id, default_provider, open_in_new_tab, capture_prefixes)
SELECT 'owner', default_provider, open_in_new_tab, capture_prefixes FROM search_settings;
DROP TABLE search_settings;
ALTER TABLE search_settings_v2 RENAME TO search_settings;

CREATE TABLE search_shortcuts_v2 (
    account_id   TEXT NOT NULL DEFAULT 'owner',
    keyword      TEXT NOT NULL,
    name         TEXT NOT NULL,
    url_template TEXT NOT NULL,
    PRIMARY KEY (account_id, keyword)
);
INSERT INTO search_shortcuts_v2 (account_id, keyword, name, url_template)
SELECT 'owner', keyword, name, url_template FROM search_shortcuts;
DROP TABLE search_shortcuts;
ALTER TABLE search_shortcuts_v2 RENAME TO search_shortcuts;

-- Shortcuts without a settings row (possible before) would otherwise be
-- hidden by the "own settings exist" check; give the owner a row.
INSERT INTO search_settings (account_id, default_provider)
SELECT 'owner', '' WHERE NOT EXISTS (SELECT 1 FROM search_settings WHERE account_id = 'owner')
  AND EXISTS (SELECT 1 FROM search_shortcuts);
