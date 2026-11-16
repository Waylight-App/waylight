-- Where an announcement was last posted (#332): the Slack or Teams channel
-- and when; empty when it was never posted.
ALTER TABLE announcements ADD COLUMN posted_channel TEXT NOT NULL DEFAULT '';
ALTER TABLE announcements ADD COLUMN posted_at TEXT NOT NULL DEFAULT '';
