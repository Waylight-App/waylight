-- The saved place whose sunset and sunrise switch the automatic colour
-- scheme (#863). An empty id follows the system, as before.
ALTER TABLE appearance_settings ADD COLUMN sun_place_id TEXT NOT NULL DEFAULT '';
ALTER TABLE appearance_settings ADD COLUMN sun_lat REAL NOT NULL DEFAULT 0;
ALTER TABLE appearance_settings ADD COLUMN sun_lon REAL NOT NULL DEFAULT 0;
