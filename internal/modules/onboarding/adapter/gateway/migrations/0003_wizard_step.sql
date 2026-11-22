-- The step the onboarding wizard stopped at, so it can be resumed (#739).
-- Empty means the wizard has not started; the ids belong to the frontend.
ALTER TABLE onboarding_state ADD COLUMN wizard_step TEXT NOT NULL DEFAULT '';
