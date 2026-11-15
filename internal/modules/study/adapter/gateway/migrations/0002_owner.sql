-- Ownership per study module (ADR 0018, #294). Existing rows belong to the
-- owner and stay visible to everyone, so guests see what they saw before.
ALTER TABLE study_modules ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE study_modules ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX study_modules_owner ON study_modules (owner_id);
