-- Ownership per record (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE focus_sessions ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE focus_sessions ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX focus_sessions_owner ON focus_sessions (owner_id);
