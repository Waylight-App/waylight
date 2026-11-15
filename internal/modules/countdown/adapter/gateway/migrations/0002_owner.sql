-- Ownership per record (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE countdown_items ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE countdown_items ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX countdown_items_owner ON countdown_items (owner_id);
