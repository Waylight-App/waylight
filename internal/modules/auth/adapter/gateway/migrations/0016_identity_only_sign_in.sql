-- An account created from an invitation through the provider (#528) has a
-- password nobody knows: its linked sign-in is the only way in, and it
-- cannot be disconnected until a password is set.
ALTER TABLE auth_identities ADD COLUMN only_sign_in INTEGER NOT NULL DEFAULT 0;
