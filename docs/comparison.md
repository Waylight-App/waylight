# How Waylight compares

*Last reviewed: 2026-11-14 for Waylight, 2026-11-11 for the others (Homarr v2.12.0, Homepage v2.8.0, Glance v0.8.8,
Dashy 4.9.0). Reviewed with every minor release of Waylight — corrections
are very welcome as an issue or pull request.*

All five projects are good, actively maintained and free. This page tries to
help you pick the right one, including when that is **not** Waylight.
Claims about the other projects link to their own documentation; "not
stated" means we did not find the information there.

## Overview

| | **Waylight** | [Homarr](https://github.com/homarr-labs/homarr) | [Homepage](https://gethomepage.dev/) | [Glance](https://github.com/glanceapp/glance) | [Dashy](https://github.com/Lissy93/dashy) |
| --- | --- | --- | --- | --- | --- |
| Setup | Docker Compose, one container (+ socket proxy); `.deb` package; ready-to-flash Raspberry Pi image ([guide](raspberry-pi.md)) | Docker Compose, Helm, NAS guides ([docs](https://homarr.dev/docs/getting-started/)) | Docker, source, Kubernetes ([docs](https://gethomepage.dev/installation/)) | Single binary or Docker ([README](https://github.com/glanceapp/glance)) | Docker, bare metal, static cloud ([README](https://github.com/Lissy93/dashy)) |
| Configuration | Web UI (drag & drop), Docker labels for service tiles; YAML planned ([#88](https://github.com/itadori26/waylight/issues/88)) | Web UI, "no YAML" | YAML files and Docker labels | YAML files, auto-reload | YAML file plus UI editor |
| Integrations | About 50 modules (AdGuard Home/Pi-hole, Docker, Proxmox, Home Assistant, Nextcloud, FRITZ!Box, WireGuard, Sonarr/Radarr, Jellyfin/Plex, Immich, CalDAV, DHL, transit, …) plus a generic JSON widget ([features](features.md)) | "80+ integrations" | "over 100 services" | Built-in widgets plus `custom-api`, `iframe`, `html`, `extension` | "several pre-built widgets" plus status checks |
| Personal widgets | Notes, to-dos (CalDAV), calendar with write access, shopping list, waste calendar, parcels, transit, weather warnings and radar | Mostly service-oriented | Mostly service-oriented | RSS, Reddit, Hacker News, YouTube, Twitch, markets, weather | Links, status checks, widgets |
| Authentication | Built in, always on: password (Argon2id), TOTP two-factor, passkeys, session overview, IP allow-list; OIDC and forward-auth SSO ([accounts](accounts.md)) | Built in, OIDC and LDAP | Optional password or OIDC gate since v2.0, no rate limiting ([docs](https://gethomepage.dev/installation/)) | Optional username/password ([docs](https://github.com/glanceapp/glance/blob/main/docs/configuration.md)) | Optional basic auth (SHA-256 hashes), Keycloak/OIDC |
| Multi-user | One owner plus up to 20 read-only guests with shared tabs, and kiosk displays ([accounts](accounts.md)) | Users, groups, permissions | No | No | Users with privileges, guest access |
| Resources | Under 9 MB image, ~75 MB RAM idle, enforced in CI ([ADR 0011](adr/0011-lean-by-default.md)) | 500 MB RAM, 600 MB image ([docs](https://homarr.dev/docs/getting-started/)) | not stated | "<20mb binary", "low memory usage" | "256MB of memory should be more than enough" |
| Security model | Read-only socket proxy, restart-only second proxy, keys never leave the host, secrets AES-encrypted ([security model](security-model.md)) | not stated | "fully proxied" API requests, host validation; recommends reverse proxy or VPN | not stated | not stated |
| Languages | English, German | Crowdin, many languages | "multiple languages" | not stated | "30+ human-translated languages" |
| Accessibility | Aims for WCAG 2.2 AA: keyboard-operable grid, screen reader support, axe-core checks in CI; no manual audit yet ([accessibility](accessibility.md)) | not stated | not stated | not stated | not stated |
| Licence | MIT | Apache-2.0 | GPL-3.0 | AGPL-3.0 | MIT |
| Maturity | **Young, one maintainer** | Established, large community | Very large community | Very large community | Large community |

## Choose …

**Homarr** if several people share the dashboard and each needs to edit it
with their own permissions, you want LDAP, or you need many *arr and media
integrations, and you have half a gigabyte of RAM to spare.

**Homepage** if you like configuration as code, want the widest range of
service integrations and automatic tiles from Docker labels, and are fine
editing YAML.

**Glance** if you want a fast, read-only "morning feed" (RSS, Reddit, Hacker
News, YouTube, markets) in a tiny binary, configured in YAML.

**Dashy** if you want a highly themeable link dashboard with status checks,
many languages, a UI editor *and* a YAML file, or a static deployment.

**Waylight** if you want one private start page for your everyday life
*and* your home server, configured entirely in the browser: notes, to-dos and
calendar that sync via CalDAV, shopping list, parcels, transit, weather
warnings, next to Docker, AdGuard Home, WireGuard, Fritz!Box and backups —
with two-factor login by default, a strict least-privilege design and a very
small footprint.

## Where Waylight is weaker (today)

- **Young project with one maintainer.** The roadmap and release pipeline are
  public ([ROADMAP.md](https://github.com/itadori26/waylight/blob/main/ROADMAP.md)),
  but there is no long track record yet.
- **One owner.** Guests and kiosk displays can look at shared tabs, but only
  the owner can change anything. There are no groups or per-user
  permissions.
- **Fewer integrations.** Homepage and Homarr support many more services. The
  [generic JSON widget](modules/apiwidget.md) covers many gaps; beyond that,
  a missing integration means writing a module
  ([developer guide](developer-guide.md)).
- **No configuration as code yet.** Everything is set in the UI; a YAML
  export and import is planned (#88).
- **Two languages.** Translations are easy to add (see CONTRIBUTING).
- **No hosted demo.** Screenshots and a one-command local demo
  ([demo mode](demo.md)) instead.
