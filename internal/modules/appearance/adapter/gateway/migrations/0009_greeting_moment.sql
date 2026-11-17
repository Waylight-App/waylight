-- A short line after the greeting by time of day and weekday (#512); off
-- unless the account switches it on.
ALTER TABLE appearance_settings ADD COLUMN greeting_moment INTEGER NOT NULL DEFAULT 0;
