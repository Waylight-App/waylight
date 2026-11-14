# Host integrations

Some widgets need a little setup on the server itself, because trisse-home
deliberately does not get the access they would otherwise require. Each
section below starts with the reason, then shows the steps. The
[security model](security-model.md) explains the boundaries in more depth.

The examples assume Docker Compose with the repository checked out in
`/opt/trisse-home`. Adjust the path to your checkout. The settings pages for
WireGuard, Wake-on-LAN and the services generate the exact lines for your
setup, so you can usually copy them from there.

## WireGuard

**Why an extra step:** `wg` needs network admin rights, and its output
includes the server's private key. Instead of giving the container either, a
small script on the host exports only the public peer status.

Run the export every minute with cron:

```sh
sudo tee /etc/cron.d/trisse-home-wireguard <<'CRON'
* * * * * root /opt/trisse-home/scripts/wireguard-export.sh
CRON
```

The script runs `wg show all dump` inside the `wireguard` container (set
`WG_CONTAINER=` to an empty value if WireGuard runs on the host). It drops the
interface line with the private key, blanks preshared keys and writes the
result to `./wireguard`, which Compose mounts read-only into trisse-home.
With `linuxserver/wireguard`, peer names are taken from its `peer_<name>`
folders; you can rename peers under Settings → WireGuard.

A peer counts as online when its last handshake is less than three minutes
old.

## Disk health

**Why an extra step:** reading S.M.A.R.T. data needs raw access to the disks,
which the container must never have. A host script writes the `smartctl`
output to a folder instead.

Install smartmontools 7.0 or newer (`sudo apt install smartmontools`) and run
the export every 30 minutes:

```sh
sudo tee /etc/cron.d/trisse-home-smart <<'CRON'
*/30 * * * * root /opt/trisse-home/scripts/smart-export.sh
CRON
```

The script writes `smartctl --json -a` for every disk that
`smartctl --scan-open` finds into `./smart` and removes files of disks that
are gone. trisse-home rates each disk like this:

| Finding | Status |
| --- | --- |
| Failed self-assessment, an attribute below its threshold, an NVMe critical warning, spare below threshold, wear of 100 % or more | critical |
| Reallocated, pending or uncorrectable sectors, NVMe media errors, wear of 90 % or more | warning |
| 60 °C or more for HDD and SATA, 70 °C or more for NVMe (critical from 70 and 80 °C) | warning |
| Export older than `DISKHEALTH_STALE_AFTER` (default two hours) | unknown |

Changes of status are sent through [notifications](notifications.md) and
recorded on the timeline. The export contains model and serial numbers, but
nothing stored on the disks.

## Wake-on-LAN

**Why an extra step:** magic packets are broadcasts, and broadcasts cannot
leave a Docker bridge network. Moving trisse-home into the host network would
remove its isolation, so a tiny relay in the host network does this one job.

Service tiles with a MAC address show a power button while the service is
offline. Turn on the relay in `.env` and restart the stack:

```sh
COMPOSE_PROFILES=wol
# Only if 255.255.255.255 leaves through the wrong interface:
# WOL_BROADCAST=192.168.178.255:9
```

The `wol-relay` service runs from the same image. It has no capabilities, a
read-only file system and no access to data, Docker or secrets. It accepts
nothing but valid magic packets on UDP port 9009, with a rate limit, and
broadcasts them into your LAN. Anyone in your LAN can send it packets, which
lets them do what every LAN device can do anyway: wake a computer.

| Setup | What to do |
| --- | --- |
| Docker Compose | `COMPOSE_PROFILES=wol` |
| Package, Pi image or a container in the host network | `WOL_RELAY=` (empty); trisse-home broadcasts itself |
| Docker without the extra container | Enable directed broadcast forwarding on the host (`sysctl net.ipv4.conf.all.bc_forwarding=1` and the same for your LAN interface), then set `WOL_RELAY=` and `WOL_BROADCAST=<your-lan>.255:9` |

The target computer must have Wake-on-LAN enabled in its firmware and its
network driver.

## Docker actions

**Why an extra step:** write access to Docker amounts to root access on the
host. Instead of loosening the read-only proxy, a second proxy allows nothing
but restarts, and trisse-home only restarts containers on your list.

The Docker widget can restart containers and show their last 200 log lines.
Add to `.env`:

```sh
COMPOSE_PROFILES=docker-actions        # or wol,docker-actions
DOCKER_ACTIONS=true
DOCKER_ACTIONS_CONTAINERS=jellyfin,paperless
```

The read-only proxy stays as it is. The additional `docker-actions-proxy`
accepts only `POST /containers/<name>/stop|restart|kill` and refuses
everything else, including reads, creating containers and `exec`. trisse-home
itself only calls *restart*, and only for the listed containers.

Logs are read through the read-only proxy and can contain secrets. List only
containers whose logs you are happy to see on the dashboard.

## Docker autodiscovery

Containers can bring their own service tile. Add labels in your compose
file, then turn on Settings → Services → *Docker autodiscovery*, or press
*Sync now*. The labels are read through the existing read-only proxy; no new
permission is needed.

```yaml
services:
  jellyfin:
    labels:
      trisse-home.name: Jellyfin
      trisse-home.url: https://jellyfin.example.com
      trisse-home.icon: 🎬
      trisse-home.description: Films and series
```

| Label | Also accepted | Meaning |
| --- | --- | --- |
| `trisse-home.url` | `homepage.href` | Link and health-check address of the tile. Required. |
| `trisse-home.name` | `homepage.name` | Name of the tile; defaults to the container name. |
| `trisse-home.icon` | `homepage.icon` | An emoji or an image URL. |
| `trisse-home.description` | `homepage.description` | A short description. |
| `trisse-home.enable` | | `false` skips the container. |

Stacks labelled for [Homepage](https://gethomepage.dev/configs/docker/) work
unchanged; `homepage.group` and `homepage.widget.*` are ignored.

Tiles are matched by container name, so recreating a container keeps its
tile, position and history. A few rules:

- **Edits.** You can edit a discovered tile. With *Labels win on the next
  sync* (the default), the labels overwrite your changes on the next sync;
  with *Keep my changes*, an edited tile is left alone.
- **Removed containers.** The tile is hidden, not deleted, and marked in the
  settings. It returns with the container. Hidden tiles are not checked, so a
  removed container does not trigger an offline alert.
- **Interval.** Between 1 and 1440 minutes, five by default. Labels that do
  not make sense, such as a URL without `http://` or `https://`, are skipped
  and counted in the result of the last sync.

## Home Assistant

The smart home widget shows the entities you pick, with their state and last
change. In Home Assistant, create a long-lived access token (your profile →
*Security*), then enter the address and the token under Settings → Home
Assistant. The token is stored encrypted and never shown again. States are
read on demand, at most every ten seconds and only while a widget is open;
nothing is polled in the background.

**Switches and buttons** are off by default. A Home Assistant token can do
everything its user can, so trisse-home keeps its own list of entities that
may be controlled:

```sh
SMARTHOME_ACTIONS_ENTITIES=switch.coffee_machine,button.garage_door
```

Switches, lights, fans and `input_boolean` entities are toggled; buttons,
`input_button`, scripts and scenes are pressed. Each action asks for
confirmation and is written to the security log. MQTT is not supported.

## Proxmox VE

The Proxmox widget shows each node's CPU, memory and root disk, and every VM
and container with its state and uptime. It needs only a read-only API
token. Create one on any node:

```sh
pveum user add monitor@pve
pveum aclmod / -user monitor@pve -role PVEAuditor
pveum user token add monitor@pve trisse-home --privsep 0   # prints the secret once
```

Enter the address (`https://<node>:8006`), the token ID
(`monitor@pve!trisse-home`) and the secret under Settings → Proxmox. Proxmox
uses a self-signed certificate, so also paste its SHA-256 fingerprint, which
you find under *node → System → Certificates* or with:

```sh
openssl x509 -in /etc/pve/local/pve-ssl.pem -noout -fingerprint -sha256
```

trisse-home then trusts exactly this certificate. Certificate checks are
never switched off.

**Start and shutdown** are off by default. List the guests in `.env` and
give the token the right to power exactly these:

```sh
PROXMOX_ACTIONS_GUESTS=100,105
```

```sh
pveum aclmod /vms/100 -user monitor@pve -role PVEVMUser   # once per guest
```

The widget then offers *start* for stopped guests and *shutdown* for running
ones. Shutdown is always a clean ACPI or container shutdown, never a hard
stop. Each action asks for confirmation and is written to the security log.

## DNS filter pause

The AdGuard Home and Pi-hole widget can switch blocking off for 5, 30 or 60
minutes, for when a site breaks because of a blocked domain. That affects
your whole network, so it is off by default:

```sh
DNS_PAUSE_PROVIDERS=adguard        # or pihole, or adguard,pihole
```

The DNS filter turns itself back on when the time is up. Every pause is
written to the security log.
