-- Ownership per to-do (ADR 0018, #294). Existing to-dos belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
-- Connected sources (todos_sources) are personal settings and follow later.
ALTER TABLE todos ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE todos ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX todos_owner ON todos (owner_id);
