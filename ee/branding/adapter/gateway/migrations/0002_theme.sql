-- The organisation's default theme and whether members are kept on it (#305).
ALTER TABLE branding_settings ADD COLUMN default_theme TEXT NOT NULL DEFAULT '';
ALTER TABLE branding_settings ADD COLUMN lock_theme INTEGER NOT NULL DEFAULT 0;
