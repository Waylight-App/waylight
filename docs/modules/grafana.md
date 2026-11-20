# Grafana

**What it does:** panels and alerts of your Grafana on the dashboard. The *Grafana panel* widget shows one panel as an image, rendered by Grafana; the *Grafana alerts* widget lists the firing and pending alerts of Grafana Alerting. Waylight talks to Grafana, so the token never reaches the browser or the app. The smart search answers "grafana alarme" / "grafana alerts".

## Configuration

Settings → Grafana: the address of Grafana (e.g. `http://192.168.1.2:3000`) and a service account token.

1. In Grafana, open **Administration → Users and access → Service accounts** and add a service account with the role **Viewer**.
2. Add a token to it and copy it (it starts with `glsa_`).
3. Paste address and token into Waylight. Waylight tests them with a dashboard search before it saves them; the token is stored encrypted.

Panels as images need the [Grafana image renderer](https://grafana.com/grafana/plugins/grafana-image-renderer/) (plugin or the `grafana/grafana-image-renderer` container). Without it the panel widget says so and links to the panel in Grafana.

## Widgets

- **Grafana panel:** pick a dashboard and a panel, a time range (e.g. `6h`, `7d`) and how often the image is refreshed. The image follows the light or dark look of Waylight.
- **Grafana alerts:** firing alerts first, then pending ones, with the summary and since when they are active. The heading links to the alert list in Grafana.

Guests see a panel or the alerts only when the widget is on one of their tabs; they cannot browse the dashboards.

## Outbound requests

Only the configured Grafana, while a Grafana widget is open or the search asks: `/api/search` (dashboards), `/api/dashboards/uid/{uid}` (panels), `/render/d-solo/{uid}` (images), `/api/prometheus/grafana/api/v1/alerts` (alerts, cached for 30 seconds). The token is sent as `Authorization: Bearer`.

## API

`GET/PUT /api/grafana/settings`, `GET /api/grafana/alerts`, `GET /api/grafana/dashboards`, `GET /api/grafana/dashboards/{uid}/panels`, `GET /api/grafana/render/{uid}/{panel}` – see `api/modules/grafana.yaml`.
