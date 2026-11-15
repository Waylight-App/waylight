-- Ownership per birthday (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE birthdays ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE birthdays ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX birthdays_owner ON birthdays (owner_id);
