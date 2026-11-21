# Tabs and widgets

Everything you see in Waylight is a widget on a tab. This page explains
how to arrange them, how the phone layout works and how to make a tab open
by itself. For a guided first run, see [first steps](first-steps.md).

## Edit mode

Press `e`, or choose *Edit* in the account menu, to change the dashboard.
Edit mode shows a frame around each widget with buttons for its settings,
duplicating and removing it, and the tab bar gains controls for adding,
renaming, moving and deleting tabs. Press `e` again or choose *Done* to
leave. There is no save button; each change is stored as soon as you make
it.

`Ctrl+Z` undoes the last change in edit mode and `Ctrl+Shift+Z` or `Ctrl+Y`
redoes it. A removed widget comes back with its settings and its content.

## Adding and arranging widgets

**+ Add widget** opens the picker. Widgets are grouped into *Everyday*,
*Information*, *Server* and *Network & integrations*. Type to filter, use the
arrow keys to move through the list and `Enter` to add it to the current
tab.

A click on a widget adds it in its default size. The ⊞ button in the corner
of each entry shows the widget in every [size](#widget-sizes) side by side
instead, with example data, so you can see what a small tile or a strip will
say before the widget is set up. Choosing a preview adds the widget in that
size.

The grid has twelve columns. Drag a widget by its frame to move it and drag
its lower right corner to resize it; other widgets move out of the way. With
the keyboard, focus a widget and use the arrow keys to move it or `Shift`
with the arrow keys to resize it. A screen reader announces the new
position.

## Widget sizes

In edit mode every widget has a size menu in its title bar with the same
sizes as the phone layout: *Small*, *Medium*, *Large* and *Strip*. The size
decides what the widget shows. Medium and large tiles show the full widget.
Small tiles and strips show its compact view, one or two lines such as
*3 open · 1 overdue*, and a click on the tile opens the full widget in a
panel at the side of the screen (a sheet from the bottom on phones). `Escape`
or the close button closes the panel and puts the focus back on the tile.
Below the widget the panel offers *Edit widget* when you may change the tab
and the widget has settings, and for widgets that watch something that can
fail (services, backups, drives) admins get *Notify me when this fails*,
which opens the notification settings. Kiosk displays show the compact view
without the panel.

When a narrow window makes a medium or large tile too small for its full
view, the widget switches to its compact view instead of cutting its content
off.

Dragging the corner of a sized widget snaps to the nearest size; with the
keyboard, `Shift` and the arrow keys step through the sizes. *Free size* in
the menu turns a widget back into a freely sized tile, which always shows the
full widget. Widgets placed before sizes existed stay free until you pick a
size, so existing dashboards keep their look.

Some compact views carry one quick action next to the short state, so the
most common step needs no panel: habits tick off the next habit due today
and to-dos tick off the most urgent one (overdue first, then the earliest due
date). The button gives feedback at once; when the server refuses, its
message is shown and nothing changes. Guests and kiosk displays do not see
quick actions.

Compact views exist for services and the status overview (*11 of 12
running*, and which service is down), the backup status, drive health, the
calendar (the next event), to-dos (the built-in list), habits and the focus
timer. Widgets whose content is the point, such as the clock, bookmarks or
notes, show their full view at every size.

You can put the same kind of widget on a tab more than once, for example two
weather widgets for two towns. **Duplicate** copies a widget with its
settings.

## Tabs

A tab has a name, an optional background image and its own set of widgets.
In edit mode you can:

- add a tab with **+ Tab**, either empty or from a template;
- save a tab as a template, so you can stamp out similar tabs later
  (templates keep the layout and widget settings, not the content such as
  notes or to-dos);
- choose a background for one tab, which overrides the background of all
  tabs (see [background image](appearance.md#background-image));
- move tabs left and right, rename and delete them.

`1` to `9` switch to the first nine tabs.

## Opening a tab by default

Normally the dashboard opens the first tab. With **Open by default** you can
give any tab rules for when it should open instead. A rule combines
weekdays, a time window and client networks, and each part is optional:

| Rule | Opens the tab … |
| --- | --- |
| Mon–Fri, 06:00–09:30 | on weekday mornings |
| 192.168.10.0/24 | when the browser is in that network, for example at work over VPN |
| Sat–Sun | at weekends, all day |

A tab can have several rules; any one of them is enough. When several tabs
match, the first tab in the tab bar wins. Rules use the server's time zone
(`TZ`). Network rules see your real address only if Waylight trusts your
reverse proxy (`TRUSTED_PROXIES`, see [configuration](configuration.md)).

## The phone layout

On narrow screens the widgets are stacked in one column. By default the
order follows the desktop layout, row by row. Each tab can have its own
phone layout instead: in edit mode on a phone, move widgets up and down, make
them shorter or taller, or hide them on phones only. **Reset from desktop
layout** discards the phone layout of that tab.

Give a widget a size to place it on a grid of four columns, as on a phone's
home screen: **Small** (two by two, two side by side), **Medium** (full width,
two cells high), **Large** (four by four) or **Strip** (one cell high). Widgets
without a size keep the full width and their height. The mobile app shows the
same grid.

The phone layout belongs to the tab, not to the device, so it applies to
every phone you use.

## The attention line and the system summary

When something on a tab needs a look, one line at the top says what it is,
for example *Services: 11 of 12 running · Nextcloud is down*, and how many
other things need a look as well. **View** scrolls to the widget and puts the
focus on it. The line takes the status widgets of the tab whose short state
is amber or red (services, backups, drives) and connections that stopped
working, such as an expired sign-in. It disappears when all is well. Colour
is used only for what needs action: amber asks for a look, red means
something is broken.

A tab with two or more status widgets also gets a summary line, for example
*Services: 12 of 12 running · Backups: Backup OK · Disk health: 3
drives healthy*. **Collapse** folds the status widgets into that line and
**Show all** brings them back. The choice is remembered per tab in this
browser; in edit mode all widgets are shown.

Zones that group a tab into *Today*, *My things* and *System* follow in #619.

## Widget status

Widgets that load data from the server refresh on their own, at an interval
that suits the source. They pause while the browser tab is in the background
or the widget is scrolled out of view. Hover over a widget or move the keyboard focus into
it to see when its data was last updated and to refresh it by hand. When the data is older than expected or the last update failed, the
badge stays visible and changes colour, so a broken integration does not
look like current data.

Some widgets also have a short state: one line such as *3 open · 1 overdue*
for to-dos, *2 of 5 today* for habits or *50 min focus today* for the focus
timer, with a level (none, ok, attention, problem). Small tiles and strips
on the dashboard ([widget sizes](#widget-sizes)), the app's home screen
widgets and the watch show it instead of the full widget; it is
served at `GET /api/<module>/compact` with the same permissions as the
module's other endpoints.

## What needs your attention

Waylight keeps a list of what needs your attention: widgets on your own tabs
that miss a setting (a departures widget without a stop) and services behind
your widgets that are not set up or stopped answering (Immich, Home
Assistant, Nextcloud, Grafana, the FRITZ!Box, Proxmox, the media server,
Sonarr/Radarr/Prowlarr and Speedtest Tracker for now). Admins also see
broken connections of the instance that none of their widgets use. Shared
tabs of others and integrations you never set up are left out. Each entry
names one thing to do, such as *choose a stop* or *check again*.

In the web interface a dot with the number of entries appears in the top
bar as soon as something needs you; it is hidden while everything works.
It opens **Settings → Needs your attention**, which lists what stopped
working first and what is not set up yet below. The button next to an entry
opens the widget's own settings on its tab, opens the integration's
settings, or asks the service again. The list updates by itself while the
page is open, so a fixed entry disappears without reloading.

The app reads the same list from `GET /api/attention` (see the API
reference). A service's state is kept for two minutes; *Check again* and
closing the settings after a change ask the services right away
(`?refresh=true`).

## Related pages

- [Feature overview](features.md): every widget and what it can do.
- [Accounts, guests and kiosk displays](accounts.md): share tabs with others.
- [Accessibility](accessibility.md): all keyboard shortcuts.
