# Sports

**What it does:** Follows a football team: its last result, the next fixture and its place in the table. Without a team the widget shows the top of the league table (#285).

## Configuration

Add the **Sports** widget, choose the league (1. Bundesliga, 2. Bundesliga, 3. Liga) and, optionally, a team in its settings. No account or API key is needed.

## Outbound requests

The community-run [OpenLigaDB](https://www.openligadb.de) (`api.openligadb.de`): the teams, the matches and the table of the chosen league, only while a widget shows them and at most every 15 minutes per league. When it does not answer, the last known data stays visible. Club logos are not shown, so the browser loads nothing from elsewhere. `SPORTS_API_URL` points the module elsewhere (for tests).

## Limits

The first provider covers the German football leagues. Other sports and countries can be added as further providers behind the same port (`usecase.Provider`).

## API

`GET /api/sports/leagues`, `GET /api/sports/teams?league=bl1`, `GET /api/sports/overview?league=bl1&team=40` – see `api/modules/sports.yaml`.
