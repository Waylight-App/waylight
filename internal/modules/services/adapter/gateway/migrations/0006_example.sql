-- Example tiles of the sample data (#854): they stay out of the status
-- summaries and go away with the sample data.
ALTER TABLE services ADD COLUMN example INTEGER NOT NULL DEFAULT 0;

-- Tiles seeded before the flag existed are recognised by their unchanged
-- seed values.
UPDATE services SET example = 1
WHERE (name = 'DuckDuckGo' AND url = 'https://duckduckgo.com/' AND description = 'Example tile – replace with your services')
   OR (name = 'OpenStreetMap' AND url = 'https://www.openstreetmap.org/' AND description = 'Example tile')
   OR (name = 'Open-Meteo' AND url = 'https://open-meteo.com/' AND description = 'Weather data source');
