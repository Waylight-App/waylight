# Native install (Debian, Ubuntu, Raspberry Pi OS)

Waylight is a single static binary, so it does not need Docker. Every
release attaches `.deb` packages for `amd64` and `arm64` that install it as a
hardened systemd service. This is also what the Raspberry Pi image is built
from ([ADR 0015](adr/0015-raspberry-pi-image.md)).

Docker stays the primary setup on a server that already runs containers; the
native package is for machines without Docker and for small single-purpose
devices.

## Install

Pick the package for your architecture (`dpkg --print-architecture`) from the
[latest release](https://github.com/Waylight-App/waylight/releases/latest),
check it and install it:

```sh
VERSION=1.2.3   # the release you want
ARCH=$(dpkg --print-architecture)
curl -fLO "https://github.com/Waylight-App/waylight/releases/download/v$VERSION/waylight_${VERSION}_$ARCH.deb"
curl -fLO "https://github.com/Waylight-App/waylight/releases/download/v$VERSION/waylight_${VERSION}_packages.sha256"
sha256sum --check --ignore-missing "waylight_${VERSION}_packages.sha256"
sudo apt install "./waylight_${VERSION}_$ARCH.deb"
```

The service starts right away on port 8090. Show the address and the
one-time setup token, open the address and enter the token:

```sh
sudo waylight-setup-token
```

## Setup token

Like `SECRET_KEY`, the setup token is generated on the device on the first
start (20 characters without look-alikes, e.g. `K7MP-Q2XA-9HTD-WE4R-ZN6C`) and
is only valid until the account exists. Ways to get it — all of them need
access to the device itself:

- `sudo waylight-setup-token` over SSH or on the console
- on a monitor attached to the device, above the login prompt
- on a Raspberry Pi: `waylight-setup.txt` on the boot partition of the SD
  card (readable on any computer)
- in the log: `sudo journalctl -u waylight | grep setup_token`

Console text and boot partition file are deleted automatically once the
account exists (`waylight-setup-info.timer`). To choose the token yourself
(automated setups), set `SETUP_TOKEN=` (at least 16 characters) in
`waylight.env`.

## What the package installs

| Path | Purpose |
| --- | --- |
| `/usr/bin/waylight` | The binary (same as in the container image) |
| `/etc/waylight/waylight.env` | Configuration (environment variables, see [.env.example](https://github.com/Waylight-App/waylight/blob/main/.env.example)) |
| `/etc/waylight/secret.env` | `SECRET_KEY` and `SETUP_TOKEN`, generated on the first start |
| `/var/lib/waylight/` | Database and automatic backups |
| `waylight.service` | systemd unit, runs as the system user `waylight` |

**`SECRET_KEY` is created on the device** on the first start (32 random bytes)
— it is never part of a package or an image. It encrypts stored integration
passwords: keep a copy of `/etc/waylight/secret.env` with your backups.

After changing the configuration: `sudo systemctl restart waylight`.

Defaults differ from Docker where Docker needed workarounds: host metrics are
read from `/proc` and `/sys` directly, Wake-on-LAN broadcasts without a relay,
and the Docker module is disabled (`MODULES_DISABLED=docker`) because there is
no Docker socket proxy.

## Sandbox

The service runs without capabilities and with a read-only view of the
system; only `/var/lib/waylight` is writable. Check it with
`systemd-analyze security waylight` (CI keeps the exposure at or below
2.0). Changes go into a drop-in (`sudo systemctl edit waylight`), never
into the packaged unit:

```ini
# Listen on port 80 (also set LISTEN_ADDR=:80 in waylight.env)
[Service]
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE
```

```ini
# Backups on a USB stick (also set BACKUP_DIR=/mnt/usb/waylight in waylight.env)
[Service]
ReadWritePaths=/mnt/usb/waylight
```

## Robustness

- **Watchdog:** the service reports readiness to systemd and pings its
  watchdog every 30 seconds while the database answers. If it hangs, systemd
  restarts it after 60 seconds (`WatchdogSec=`).
- **Power loss:** SQLite runs in WAL mode with `synchronous=NORMAL`. After a
  sudden power cut the database is always consistent; at worst the last
  change before the cut is lost. Daily backups cover the rest.

## Update and rollback

**With the APT repository** (recommended; the Raspberry Pi image is set up
this way and updates itself every night):

```sh
sudo curl -fsSLo /usr/share/keyrings/waylight-archive-keyring.asc \
  https://waylight-app.github.io/waylight/apt/waylight-archive-keyring.asc
sudo tee /etc/apt/sources.list.d/waylight.sources >/dev/null <<'EOF'
Types: deb
URIs: https://waylight-app.github.io/waylight/apt
Suites: stable
Components: main
Signed-By: /usr/share/keyrings/waylight-archive-keyring.asc
EOF
sudo apt update && sudo apt install waylight
```

Afterwards `sudo apt upgrade` updates Waylight like every other package,
and `unattended-upgrades` can do it automatically (add
`"origin=waylight,codename=stable";` to `Unattended-Upgrade::Origins-Pattern`).
Check the key's fingerprint against the one in the release notes before
trusting it.

**Without the repository**, install the newer `.deb` the same way as the
first one.

Either way configuration, key and data are kept, the service restarts, and
**before every upgrade the database is copied** to
`/var/lib/waylight/backups/pre-upgrade/` (the three newest copies stay).

**Rollback:** the repository keeps the last five versions.

```sh
sudo apt install --allow-downgrades waylight=1.2.3
```

Database migrations only add columns and tables, so an older version runs on
newer data. To return to the exact data of before the upgrade:
`sudo systemctl stop waylight`, delete `waylight.db-wal` and
`waylight.db-shm` in `/var/lib/waylight`, copy the matching
`backups/pre-upgrade/waylight-<version>-<time>.db` to `waylight.db`
(and its `-wal` file, if there is one, to `waylight.db-wal`), then start
the service again.

## Forgot the password?

```sh
sudo -u waylight sh -c 'set -a; . /etc/waylight/secret.env; . /etc/waylight/waylight.env; exec waylight reset-account'
sudo systemctl restart waylight
sudo waylight-setup-token
```

## Moving from Docker

1. Stop the container and copy the database out of the volume (see
   [backups](backups.md#copies-away-from-the-server)) to `/var/lib/waylight/waylight.db`.
2. Put your existing `SECRET_KEY` into `/etc/waylight/waylight.env` — a
   key there always wins over the generated one.
3. `sudo chown -R waylight:waylight /var/lib/waylight` and
   `sudo systemctl restart waylight`.

## Uninstall

- `sudo apt remove waylight` stops the service and keeps configuration and
  data.
- `sudo apt purge waylight` also deletes `/etc/waylight`. The data in
  `/var/lib/waylight` is **never deleted automatically**; the generated key
  is moved next to it (`secret.env.purged`) so a reinstall can still decrypt
  it. Delete the directory yourself if you no longer need it.
