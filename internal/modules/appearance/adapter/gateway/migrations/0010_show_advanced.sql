-- Settings for developers and power users stay hidden until the account
-- switches them on (#602).
ALTER TABLE appearance_settings ADD COLUMN show_advanced INTEGER NOT NULL DEFAULT 0;
