-- Ownership per bookmark and folder (ADR 0018, #294). Existing rows belong
-- to the owner and stay visible to everyone, so guests see what they saw
-- before. Tags, link checks and icons belong to their bookmark.
ALTER TABLE bookmarks ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE bookmarks ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX bookmarks_owner ON bookmarks (owner_id);
ALTER TABLE bookmark_folders ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE bookmark_folders ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX bookmark_folders_owner ON bookmark_folders (owner_id);
