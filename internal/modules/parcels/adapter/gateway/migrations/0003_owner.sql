-- Ownership per parcel (ADR 0018, #294). Existing parcels belong to the
-- owner and stay visible to everyone, so guests see what they saw before.
-- Parcels found in the mailbox are created by the system for the owner;
-- mail settings are personal settings and follow later.
ALTER TABLE parcels ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE parcels ADD COLUMN visibility TEXT NOT NULL DEFAULT 'everyone';
CREATE INDEX parcels_owner ON parcels (owner_id);
