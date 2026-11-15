-- Personal tabs (ADR 0018, #294). Every existing tab belongs to the owner,
-- who becomes the first admin. Tabs are private by default: a member sees
-- only their own; guests and kiosk displays keep seeing the tabs shared with
-- them by ID. Team tabs (#295) will use group visibility.
ALTER TABLE dashboard_tabs ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE dashboard_tabs ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private';
CREATE INDEX dashboard_tabs_owner ON dashboard_tabs (owner_id);
