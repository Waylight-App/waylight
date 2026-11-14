# Getting started

This guide gets trisse-home running and takes you through the first login.
It takes about ten minutes.

## Pick an installation

| You have … | Use | Guide |
| --- | --- | --- |
| A server that already runs Docker | Docker Compose | below |
| A Debian, Ubuntu or Raspberry Pi OS machine without Docker | The `.deb` package, a hardened systemd service | [Native install](native-install.md) |
| A spare Raspberry Pi and nothing else | The ready-made SD card image, updates itself | [Raspberry Pi](raspberry-pi.md) |

All three run the same program and store the same database, so you can move
between them later. Docker Compose is the most common setup and the one this
page describes.

## Install with Docker Compose

You need Docker with the Compose plugin. The repository is only used for
`compose.yaml`, the example configuration and a few optional helper scripts;
the image itself comes from `ghcr.io/itadori26/trisse-home` (amd64 and
arm64).

```sh
git clone https://github.com/itadori26/trisse-home.git
cd trisse-home
cp .env.example .env
sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
sudo docker compose up -d
```

`SECRET_KEY` encrypts the passwords and tokens you later enter for your
integrations. Keep a copy somewhere safe: a database backup is of limited use
without it.

Two settings in `.env` are worth a look before you go further:

- `BIND_ADDRESS`: set it to your server's LAN address so the port is not
  published on every interface.
- `ALLOWED_CIDRS`: the networks that may open the app. The default allows
  loopback and all private ranges. Narrow it if your LAN is shared.

Every other variable is described in the [configuration reference](configuration.md).

## Create your account

The first start prints a one-time setup token to the log:

```sh
sudo docker compose logs trisse-home | grep setup_token
```

Open `http://<server-ip>:8090`, enter the token and choose a password of at
least 12 characters. The token stops working as soon as the account exists.

The setup also asks what your start page should show:

- **Everyday** (preselected): weather, calendar, to-dos, news. Widgets and
  settings for a home server or network (Docker, Proxmox, Pi-hole, …) are
  hidden.
- **Everyday and home network:** everything, including your own devices and
  services.

You can switch at any time under **Settings → Appearance**; switching only
shows or hides widgets and settings and never deletes anything.

You land on a dashboard that is already filled with sample data: a *Home*
tab with search, clock, weather, to-dos, a welcome note, bookmarks, news and
a calendar, and, with "Everyday and home network", a *Server* tab with
service tiles, host metrics, Docker and AdGuard Home. It is there to show what each widget does. Change or delete
anything you like. If you would rather start from an almost empty page, set
`SAMPLE_DATA=false` before the first start.

A short guided tour opens after the first login. You can restart it later
under Settings → About.

## Secure the account

Two steps are worth doing right away:

1. **Turn on two-factor authentication** under Settings → Account. You get
   ten recovery codes; store them outside trisse-home.
2. **Do not forward the port on your router.** Use a VPN such as WireGuard to
   reach the dashboard from outside. If you want HTTPS, a public name or
   passkeys, put trisse-home behind a [reverse proxy](reverse-proxy.md).

## Next steps

- [First steps](first-steps.md) shows you how to arrange tabs and widgets.
- [Backups](backups.md) explains what is backed up automatically and how to
  copy backups off the server.
- [Updates](updates.md) covers updating, pinning a version and rolling back.

## If something goes wrong

**The page does not load.** Check that the container is healthy with
`sudo docker compose ps`. If your browser runs on a network outside
`ALLOWED_CIDRS`, the server refuses the connection and logs the address it
rejected: `sudo docker compose logs trisse-home | grep -i denied`.

**You lost the setup token.** Run the `grep` command above again; the token
is printed on every start until the account exists.

**You forgot your password.** Reset the account and set it up again with a
new token. This removes the owner's sign-in data (password, two-factor,
passkeys, sessions); your tabs, widgets and everything in them stay.

```sh
sudo docker compose exec trisse-home /trisse-home reset-account
sudo docker compose restart trisse-home
sudo docker compose logs trisse-home | grep setup_token
```
