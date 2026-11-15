-- Ownership per note (ADR 0018, #294). Existing notes belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE notes ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE notes ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX notes_owner ON notes (owner_id);
