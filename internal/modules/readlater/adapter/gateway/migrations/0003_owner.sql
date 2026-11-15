-- Ownership per record (ADR 0018, #294). Existing rows belong to the owner
-- and stay visible to everyone, so guests see what they saw before.
ALTER TABLE readlater_articles ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE readlater_articles ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX readlater_articles_owner ON readlater_articles (owner_id);
