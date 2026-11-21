# Integrations

**What it does:** an admin saves the address and the credential of a
service once, as a named connection such as "Home" or "Office", and widgets
of that service choose it. A service can have several connections, for
example two Proxmox clusters.

## Status

This is the store and its API. Widgets and the settings pages of the
integrations do not use it yet; their own settings stay as they are until
they move to connections (#610).

## Services

Each integration module declares the services it reaches and the
credential fields they need. `GET /api/integrations/services` lists them:

| Service | Product | Fields (secret in bold, optional in italics) |
| --- | --- | --- |
| `adguard` | AdGuard Home | `username`, **`password`** |
| `fritzbox` | FRITZ!Box | *`username`*, ***`password`*** |
| `grafana` | Grafana | **`token`** |
| `homeassistant` | Home Assistant | **`token`** |
| `immich` | Immich | **`apiKey`** |
| `jellyfin`, `plex` | Jellyfin, Plex | **`token`** |
| `nextcloud` | Nextcloud | `username`, **`appPassword`** |
| `pihole` | Pi-hole | *`username`*, **`password`** |
| `proxmox` | Proxmox VE | `tokenId`, **`secret`**, *`fingerprint`* |
| `sonarr`, `radarr`, `prowlarr` | Sonarr, Radarr, Prowlarr | **`apiKey`** |
| `speedtest` | Speedtest Tracker | **`token`** |
| `tankerkoenig` | Tankerkönig | **`apiKey`**; the address is fixed |

A new connection must name one of these services and fill its required
fields; fields the service does not know are refused. Connections of a
service whose module is switched off stay in the list and editable.

## Rules

- Only admins see and change connections; they belong to the instance.
- A name is unique per service, compared without case. An instance keeps
  at most 100 connections.
- The credential is a set of named fields (for example `apiKey`, or
  `username` and `password`), at most four. It is stored encrypted with
  `SECRET_KEY`. Secret fields are never returned: the API says which fields
  are set and shows the values of the others (`values`), such as a user name.
- Moving a connection to another address needs the credential again, so a
  stored one is never sent to a new host.
- Removing a connection that widgets name in their settings
  (`connectionId`) is refused until it is confirmed; the list shows which
  widgets these are.
- If `SECRET_KEY` changes, the credentials can no longer be read. The list
  still opens and shows no fields; enter the credentials again.

## Checking a connection

`POST /api/integrations/connections/{id}/check` sends the request the
service's settings page makes before saving, with the stored address and
credential. It answers 204 when the service accepts it and otherwise the
same message the settings page shows.

## Config as code

The `integrations` section of [waylight.yaml](../config-as-code.md) lists
service, name and address. It never contains a credential. While the file
manages the section, connections cannot be added, renamed or removed in the
app, but credentials can still be entered.

## API

`GET /api/integrations/connections` (optionally `?service=immich`)
lists the connections with `fields`, `values` and `usedBy`;
`GET /api/integrations/services` lists the services; `POST /api/integrations/connections`,
`PUT /api/integrations/connections/{id}`,
`PUT /api/integrations/connections/{id}/credential` and
`DELETE /api/integrations/connections/{id}?confirm=true` change them. See
`api/modules/integrations.yaml`.

## Outbound requests

Only the check: one request to the connection's address, the same one the
service's settings page makes.
