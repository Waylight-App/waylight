-- Todoist lists (#170) name a project instead of a URL.
ALTER TABLE todos_sources ADD COLUMN list_id TEXT NOT NULL DEFAULT '';
