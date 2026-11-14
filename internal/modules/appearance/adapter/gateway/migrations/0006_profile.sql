-- #281: what the start page shows. Existing installations keep everything
-- visible (homelab); the first setup asks.
ALTER TABLE appearance_settings ADD COLUMN profile TEXT NOT NULL DEFAULT 'homelab';
