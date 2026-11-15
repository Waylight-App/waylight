-- Connected to-do lists belong to one account (ADR 0018, #294). Existing
-- ones are the owner's. Members never see the owner's external lists;
-- guests and kiosk displays keep seeing the owner's, as before.
ALTER TABLE todos_sources ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE todos_sources ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private';
CREATE INDEX todos_sources_owner ON todos_sources (owner_id);
