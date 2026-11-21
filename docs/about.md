# What Waylight is and is not

This page is for deciding whether Waylight is worth installing. It
describes the ideas behind the project, what you need to run it and where it
falls short today.

## The idea

Most homelab dashboards are a wall of service links and status lights.
Waylight started from a different question: what would you want to see
first thing in the morning? The answer was a mix. The weather and the rain
radar, the next train, today's appointments and to-dos, the bin that goes
out tonight, and, below that, whether the server, the backups and the VPN
are fine.

So Waylight is a personal start page first and a server dashboard second.
Both halves live on the same grid, and you decide which widgets go on which
tab.

## Principles

**Configured in the browser.** You add, move and resize widgets with the
mouse or the keyboard. There is no YAML file to keep in sync. Settings that
change what the server is allowed to do, such as restarting containers, stay
in the environment on purpose (see below).

**Private by default.** Every account signs in with a password, optionally
with two-factor authentication or a passkey. Only networks on an allow-list
can reach the app at all. There is no telemetry, and the server only talks to
the services you configure. The [outbound requests](outbound-requests.md)
page lists every one of them.

**Least privilege.** The app reads Docker through a read-only proxy and never
sees your WireGuard keys. Anything that can change your system (restarting a
container, switching a light, pausing the DNS filter, starting a VM) is off
until you list the exact containers, entities or guests in the environment.
A stolen browser session cannot widen that list. The
[security model](security-model.md) explains each boundary.

**Lean.** One binary with the web interface built in, a container image of
under 9 MB, roughly 75 MB of memory when idle. Image size, memory and the
JavaScript loaded on start are budgets that the CI enforces
([ADR 0011](adr/0011-lean-by-default.md)). Every module can be switched off.

**Built to be maintained.** The Go backend and the Svelte frontend follow a
strict Clean Architecture, checked by linters and tests. A new widget is a
set of files of its own, so contributions rarely collide
([architecture](architecture.md)).

## What you need

- A machine that runs around the clock: a home server, a NAS with Docker, a
  small VM or a Raspberry Pi (3B, Zero 2 W or newer, 64-bit). amd64 and arm64
  are supported.
- Docker with the Compose plugin, *or* Debian, Ubuntu or Raspberry Pi OS for
  the [native package](native-install.md), *or* an SD card for the
  [Raspberry Pi image](raspberry-pi.md).
- A browser. Current Firefox, Chrome, Edge and Safari work, on the desktop
  and on phones.
- For installing it as an app, offline mode and passkeys: HTTPS, usually
  through a [reverse proxy](reverse-proxy.md).

## Current limits

Waylight is a young project with one maintainer. That shows in a few
places:

- **One owner.** Guests can see the tabs you share with them, but they cannot
  edit anything. There are no groups or fine-grained permissions.
- **Fewer integrations than the big dashboards.** Homepage and Homarr support
  far more services. The [generic JSON widget](modules/apiwidget.md) covers
  many gaps; for the rest you would write a module.
- **No configuration as code yet.** A YAML export and import is planned
  ([#88](https://github.com/Waylight-App/waylight/issues/88)).
- **English and German only.** Adding a language is a matter of translating
  one file per feature ([contributing](https://github.com/Waylight-App/waylight/blob/main/CONTRIBUTING.md#ui-texts-and-translations)).
- **Several integrations are Germany-specific:** weather warnings, rain radar
  and pollen from the DWD, fuel prices from Tankerkönig, public and school
  holidays. They simply stay unused elsewhere.
- **Accessibility is checked automatically, not yet by hand** with screen
  readers ([accessibility](accessibility.md)).

If any of these is a deal breaker, the [comparison](comparison.md) suggests
alternatives.

## Try it first

You can run a local demo with sample data in one command, without creating an
account. See [demo mode](demo.md).
