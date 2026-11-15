-- Ownership per record (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE photos ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE photos ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX photos_owner ON photos (owner_id);
