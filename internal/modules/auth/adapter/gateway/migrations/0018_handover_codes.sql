-- Browser handover (#560): a pairing code created after a sign-in in the
-- browser is bound to the PKCE challenge of the app that asked for it.
ALTER TABLE auth_pairing_codes ADD COLUMN challenge TEXT NOT NULL DEFAULT '';
