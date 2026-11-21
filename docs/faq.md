# Frequently asked questions

## Can I reach my dashboard from outside my home?

Yes, but not by forwarding a port on your router. Connect to your home
network through a VPN such as WireGuard, and Waylight behaves as if you
were at home. If you want a public `https://` address anyway, put it behind
a [reverse proxy](reverse-proxy.md), keep two-factor authentication on and
consider [single sign-on](reverse-proxy.md#single-sign-on-optional) in front
of it.

## Does it work without an internet connection?

Mostly. Notes, to-dos, bookmarks, countdowns, habits, the focus timer and
everything about your server work in the LAN alone. Widgets that depend on
outside services, such as the weather, departures or parcel tracking, show
their last data until the connection is back. Phones with the
[installed app](apps-and-offline.md) keep showing the last data even when
they cannot reach the server.

## Can my family use it too?

They can look at the tabs you share with them through
[guest accounts](accounts.md#guests), each with their own password, and a
wall tablet can show tabs as a [kiosk display](accounts.md#kiosk-displays).
Only the owner can change things. If everyone needs their own editable
dashboard, Waylight is not the right tool today; see the
[comparison](comparison.md).

## Is a Raspberry Pi fast enough?

Yes. Waylight needs around 75 MB of memory and very little CPU. A
Raspberry Pi 3B runs it; a Pi 4 or 5 leaves plenty of room for other
services. The [ready-made image](raspberry-pi.md) is the easiest way to set
one up.

## Where is my data stored?

In one SQLite file: `/data/waylight.db` in the Docker volume, or
`/var/lib/waylight/waylight.db` with the package and the Pi image.
Daily [backups](backups.md) sit next to it. Nothing is stored anywhere
else, and nothing is sent to the project.

## I use Homepage. Can I keep my Docker labels?

Yes. Service tiles can be created from Docker labels, and the `homepage.*`
labels are understood as well. See
[Docker autodiscovery](host-integrations.md#docker-autodiscovery). Homepage's
service widgets do not carry over; check the [feature overview](features.md)
for the Waylight equivalents.

## Can I configure it with a file instead of the web interface?

Not yet. Settings that change what the server may do are environment
variables ([configuration](configuration.md)); everything else is set in the
browser. Configuration as code is planned in
[#88](https://github.com/Waylight-App/waylight/issues/88). Until then, the
[JSON export](backups.md#export-and-import) gives you a readable copy of
everything.

## Which service I use is missing. What now?

If the service has a JSON API, the [JSON widget](modules/apiwidget.md) can
probably show what you need. Otherwise open a feature request in the
[issue tracker](https://github.com/Waylight-App/waylight/issues), or write a
module yourself with the [developer guide](developer-guide.md); a simple one
takes an afternoon.

## Who is behind the project?

Waylight is maintained by one person in their spare time and published
under the MIT licence. Features for teams are planned in a separate directory,
`ee/`, under the Elastic License 2.0: it is not open source, using them
needs a licence key, and the code never becomes MIT
([ADR 0020](adr/0020-ee-stays-proprietary.md)). Nothing that exists today moves there.
Planned work is on the
[roadmap](https://github.com/Waylight-App/waylight/blob/main/ROADMAP.md).
