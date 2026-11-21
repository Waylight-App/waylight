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

The grid has twelve columns. Drag a widget by its frame to move it and drag
its lower right corner to resize it; other widgets move out of the way. With
the keyboard, focus a widget and use the arrow keys to move it or `Shift`
with the arrow keys to resize it. A screen reader announces the new
position.

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

## Widget status

Widgets that load data from the server refresh on their own, at an interval
that suits the source. They pause while the browser tab is in the background
or the widget is scrolled out of view. Hover over a widget or move the keyboard focus into
it to see when its data was last updated and to refresh it by hand. When the data is older than expected or the last update failed, the
badge stays visible and changes colour, so a broken integration does not
look like current data.

Some widgets also have a short state: one line such as *3 open · 1 overdue*
for to-dos, *2 of 5 today* for habits or *50 min focus today* for the focus
timer, with a level (none, ok, attention, problem). Small tiles, the app's
home screen widgets and the watch show it instead of the full widget; it is
served at `GET /api/<module>/compact` with the same permissions as the
module's other endpoints.

## What needs your attention

Waylight keeps a list of what needs your attention: widgets on your own tabs
that miss a setting (a departures widget without a stop) and services behind
your widgets that are not set up or stopped answering (Immich, Home
Assistant, Nextcloud, Grafana, the FRITZ!Box for now). Admins also see
broken connections of the instance that none of their widgets use. Shared
tabs of others and integrations you never set up are left out. Each entry
names one thing to do, such as *choose a stop* or *check again*.

The web interface and the app read the list from `GET /api/attention`
(see the API reference). A service's state is checked at most every two
minutes, so a fixed problem can take that long to disappear.

## Related pages

- [Feature overview](features.md): every widget and what it can do.
- [Accounts, guests and kiosk displays](accounts.md): share tabs with others.
- [Accessibility](accessibility.md): all keyboard shortcuts.
