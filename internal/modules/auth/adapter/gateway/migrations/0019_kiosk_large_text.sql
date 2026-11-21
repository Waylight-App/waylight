-- Larger text on a kiosk display (#680); existing displays keep the normal size.
ALTER TABLE auth_kiosks ADD COLUMN large_text INTEGER NOT NULL DEFAULT 0;
