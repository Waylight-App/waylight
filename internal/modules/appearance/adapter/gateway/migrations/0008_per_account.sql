-- Appearance per account (ADR 0018, #294): theme, language, greeting and
-- launcher are personal settings. The single row becomes the owner's.
-- SQLite cannot drop the CHECK (id = 1), so both tables are rebuilt; no
-- other table references them. account_id defaults to the owner, so an
-- export from before this change imports as the owner's settings.
CREATE TABLE appearance_settings_v2 (
    account_id       TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    theme            TEXT NOT NULL,
    accent_color     TEXT NOT NULL DEFAULT '',
    background_image TEXT NOT NULL DEFAULT '',
    custom_css       TEXT NOT NULL DEFAULT '',
    language         TEXT NOT NULL DEFAULT '',
    greeting_name    TEXT NOT NULL DEFAULT '',
    launcher         TEXT NOT NULL DEFAULT 'both',
    profile          TEXT NOT NULL DEFAULT 'homelab'
);
INSERT INTO appearance_settings_v2 (account_id, theme, accent_color, background_image, custom_css, language, greeting_name, launcher, profile)
SELECT 'owner', theme, accent_color, background_image, custom_css, language, greeting_name, launcher, profile FROM appearance_settings;
DROP TABLE appearance_settings;
ALTER TABLE appearance_settings_v2 RENAME TO appearance_settings;

CREATE TABLE appearance_css_backup_v2 (
    account_id   TEXT NOT NULL DEFAULT 'owner' PRIMARY KEY,
    theme        TEXT NOT NULL DEFAULT '',
    css          TEXT NOT NULL,
    accent_color TEXT NOT NULL DEFAULT ''
);
INSERT INTO appearance_css_backup_v2 (account_id, theme, css, accent_color)
SELECT 'owner', theme, css, accent_color FROM appearance_css_backup;
DROP TABLE appearance_css_backup;
ALTER TABLE appearance_css_backup_v2 RENAME TO appearance_css_backup;
