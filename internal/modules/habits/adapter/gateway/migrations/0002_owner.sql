-- Ownership per record (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE habits ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE habits ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX habits_owner ON habits (owner_id);
