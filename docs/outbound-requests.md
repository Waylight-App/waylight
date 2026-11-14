# Outbound requests

trisse-home sends no telemetry and contacts no server of its own. It only
talks to services you set up or switch on. This page lists every outside
connection, when it happens and how often, so you can check the claim and
plan your firewall.

## Services you configure

These requests go to addresses that only the signed-in owner enters in the
settings, and only once a service is set up. Their passwords and tokens are
stored encrypted.

| Service | When and how often |
| --- | --- |
| Service tiles | A health check of each tile every minute; the TLS certificate is read from the same connection. |
| CalDAV calendars and task lists | Every 15 minutes, and when you add an event or item. |
| ICS subscriptions (calendar, waste, school holidays) | At their refresh interval, hourly by default, with `If-None-Match` and `If-Modified-Since`; school holidays at most weekly. |
| RSS, Atom and JSON feeds | Every 30 minutes. |
| AdGuard Home | Every minute. |
| Pi-hole | `/api/auth` once per session, then statistics and the blocking state every minute. |
| Nextcloud | Every five minutes. |
| FRITZ!Box | Every 15 seconds; the call list only when a calls widget exists. |
| Home Assistant | `/api/states` while a smart home widget is open, at most every ten seconds. |
| Proxmox VE | `/api2/json/cluster/resources` every 30 seconds. |
| Sonarr, Radarr, Prowlarr | While an arr widget is open, at most once a minute. |
| Jellyfin, Plex | While a media widget is open, at most every 15 seconds. |
| Immich | While an Immich widget is open, at most every five minutes. |
| Speedtest Tracker | Every ten minutes. |
| JSON widget endpoints | While a widget shows them, at most once per configured interval. |
| Ollama | Only when you ask for a summary. |
| WebDAV off-site backups | After each backup, and when you restore from it. |
| ntfy, Gotify, webhooks, SMTP | When a notification is sent. |
| Electricity prices (`api.awattar.de` / `api.awattar.at`) | Only while an electricity price widget is shown, at most every 15 minutes per bidding zone ([details](modules/electricity.md)). |
| IMAP mailbox (parcels) | One folder, read-only over TLS, at the interval you choose. Off by default ([details](parcels-mail.md)). |

## Public services

These are contacted only when a module that needs them is in use.

| Service | When and how often |
| --- | --- |
| [Open-Meteo](https://open-meteo.com) | Weather for the places of your weather widgets. |
| Deutscher Wetterdienst | Weather warnings, rain radar, pollen and UV index for places in Germany. The radar image is fetched by the server, so your browser never contacts the DWD. Pollen and UV only when a widget shows them, at most hourly. |
| [Transitous](https://transitous.org) | Departures, cached for one minute per stop. Please respect its fair-use policy, or run your own MOTIS and set `TRANSIT_API_URL`. |
| DHL tracking API | Hourly for parcels not yet delivered, with your API key. |
| Tankerkönig | While a fuel widget is open, at most every five minutes. |
| MusicBrainz and Cover Art Archive | Artist search when you use it; releases of followed artists every six hours, at most one request per second. |
| Steam and Epic Games Store | Every twelve hours once game deals are switched on. |
| Launch Library 2 and NASA APOD | Once switched on, only while a space widget shows them, at most hourly and every three hours. |
| Container registries | Image digests of your running containers, at most every six hours per image, only with the Docker update check on. |
| GitHub releases API | At most once a day, only with the update check under Settings → About on. |

## Pages you save

Bookmarks and the read-later list fetch the pages you give them:

- **Bookmark details** (off by default): the page's title and icon, once when
  you add a bookmark, at most 512 KB per page and 64 KB per icon.
- **Link check** (off by default): each bookmarked page once a week.
- **Read later**: the page and its images, once, when you save it.

These requests time out after eight seconds, follow at most five redirects
and never connect to loopback, link-local (including the cloud metadata
address `169.254.169.254`), unspecified or multicast addresses. Addresses in
your LAN stay allowed, because home-server bookmarks point there on purpose.

## Why static analysis complains

Code scanners report several of these requests as "server-side request
forgery", because the server fetches URLs that come from user input. This is
intended: only the single owner can enter those URLs, behind a session, CSRF
protection and the network allow-list. Integration targets are additionally
validated when saved and again on every connection (see
[integration targets](security-model.md#integration-targets-191)).

Two other findings come up regularly:

- **MD5 in the FRITZ!Box login.** TR-064 uses HTTP digest authentication,
  which the protocol defines with MD5. The password is not stored as an MD5
  hash; it is encrypted with `SECRET_KEY`, and MD5 only answers the box's
  login challenge.
- **Open redirect in the search endpoint.** The endpoint redirects to the
  search engine or to a URL typed into the search field. It requires a
  session, and the `SameSite=Strict` cookie is never sent with requests from
  other sites, so it cannot be abused to redirect strangers.
