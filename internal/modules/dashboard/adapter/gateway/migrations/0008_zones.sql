-- Zones (#619): a tab can be laid out in Today, My things and System instead
-- of the free grid; a widget's zone '' lets the client pick one by kind.
ALTER TABLE dashboard_tabs ADD COLUMN zones INTEGER NOT NULL DEFAULT 0;
ALTER TABLE dashboard_widgets ADD COLUMN zone TEXT NOT NULL DEFAULT '';
