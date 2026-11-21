-- Signing in with Home Assistant (#654) stores a refresh token next to the
-- short-lived access token in token_encrypted. A pasted long-lived token
-- leaves these empty.
ALTER TABLE smarthome_settings ADD COLUMN refresh_encrypted TEXT NOT NULL DEFAULT '';
ALTER TABLE smarthome_settings ADD COLUMN client_id TEXT NOT NULL DEFAULT '';
ALTER TABLE smarthome_settings ADD COLUMN token_expiry INTEGER NOT NULL DEFAULT 0;
