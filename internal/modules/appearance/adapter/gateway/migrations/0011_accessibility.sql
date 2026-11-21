-- Accessibility options of an account (#681): text size, high contrast and
-- the simple view. Defaults keep the look of existing accounts.
ALTER TABLE appearance_settings ADD COLUMN text_size TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE appearance_settings ADD COLUMN high_contrast INTEGER NOT NULL DEFAULT 0;
ALTER TABLE appearance_settings ADD COLUMN simple_view INTEGER NOT NULL DEFAULT 0;
