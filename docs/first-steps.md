# First steps

This tutorial turns the sample dashboard into your own. You will add a tab,
place a few widgets, connect your calendar and try the search. Plan about ten
minutes. It assumes you have just finished [getting started](getting-started.md)
and are signed in.

## 1. Look around

The bar at the top holds, from left to right, the tabs, the search field, the
launcher and the account menu. Below it sit the widgets of the current tab on
a grid of twelve columns.

Press `?` to see every keyboard shortcut. `1` to `9` switch tabs, and `Ctrl+K`
or `/` opens the quick search from anywhere.

## 2. Make a tab of your own

1. Press `e`, or choose *Edit* in the account menu. The dashboard switches to
   edit mode: widgets get a frame, and the tab bar shows **+ Tab**.
2. Click **+ Tab**, name it *Morning* and pick *Empty tab* under *Start from*.
   You could also start from one of your saved tab templates.
3. Click **+ Add widget**. The picker groups the widgets into *Everyday*,
   *Information*, *Server* and *Network & integrations*; type to filter.
   Add **Weather**, **Calendar** and **To-dos**.

## 3. Arrange the widgets

Drag a widget by its frame to move it and pull its lower right corner to
resize it. Neighbouring widgets make room. With the keyboard, tab to a widget
and use the arrow keys to move it, or `Shift` and the arrow keys to resize it.

Removed something by mistake? `Ctrl+Z` brings it back, `Ctrl+Shift+Z` redoes
the change. When the layout looks right, press `e` again to leave edit mode.
Changes are saved as you go.

## 4. Point the weather at your town

Open the settings of the weather widget (the settings button on its frame in
edit mode)
and search for your town. Weather comes from Open-Meteo and needs no API key.
For locations in Germany you also get the official DWD warnings, and you can
add the rain radar as a separate widget.

## 5. Connect your calendar

Calendars use CalDAV, which Nextcloud, iCloud (with an app password),
Fastmail, Radicale and most other calendar servers speak. The settings open
on an overview of eight areas, from *My account* to *System & info*; each tile
says what it holds and, where it can, how things stand, for example how many
notification channels are on or when the last backup ran. Open Settings →
*Dashboard & widgets* → Calendar and enter the address of each calendar (one per line), your user
name and an app password. In Nextcloud you find a calendar's address in its
menu under *Copy private link*. Read-only ICS links, such as a school
timetable or a football club's fixtures, can be subscribed there as well.

The password is stored encrypted and is never sent back to the browser.

## 6. Try the search

Click the search field or press `/`, then try:

- a few letters of a bookmark or note: matching entries appear, `Enter`
  opens the first one;
- `!w Raspberry Pi`: DuckDuckGo bangs work, here a Wikipedia search;
- `19 % of 240` or `5 km in mi`: the answer appears while you type;
- `/cpu`: the current load of your server.

To add a to-do without leaving the page, type `t buy milk` and press
`Enter`. The [search guide](search.md) lists all prefixes.

## 7. Have it open by default

The *Morning* tab is most useful when it opens on its own. In edit mode,
click **Open by default** next to the tab's name. Add a rule for Monday to
Friday from 06:00 to 09:30. Outside that window the dashboard opens your
first tab as before. Rules can also depend on the network the browser is in,
so a laptop can show a different tab at work.

## Where to go from here

- Share a tab with your household as a [guest account](accounts.md#guests),
  or show it on a wall tablet as a [kiosk display](accounts.md#kiosk-displays).
- Get [notifications](notifications.md) on your phone when a service goes
  down or a backup fails.
- Install the dashboard as an [app on your phone](apps-and-offline.md).
- Browse the [feature overview](features.md) for widgets you have not tried.
