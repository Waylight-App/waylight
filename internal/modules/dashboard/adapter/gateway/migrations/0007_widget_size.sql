-- Web size of a widget (#595): small, medium, large or strip; '' keeps a freely sized tile.
ALTER TABLE dashboard_widgets ADD COLUMN size TEXT NOT NULL DEFAULT '';
