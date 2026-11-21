# Presence

With a licence key that includes `presence` ([licence keys](licence-keys.md)),
the *Presence* widget shows who is in the office, working from home or away
today. This needs a server built with the `ee` features.

## Setting your status

Everyone who signs in (the owner, admins and members) sets their own status
in the widget: **Office**, **Home** or **Away**, with an optional day of
return when away. Choosing the active status again clears it.

The search field on the start page sets it too: type `office`, `home` or
`away` (or `büro`, `homeoffice`, `weg`) and choose *Set my status* under
*Actions*. The first letters are enough, for example `hom`.

A status counts for the day it was set and lapses at midnight; statuses of
earlier days are deleted within the next hour. Nothing is ever guessed:
without a status, the widget shows *not set*. Presence keeps no history.

A planned absence from [absences](absences.md) wins over the status someone
set: a remote day shows *home*, any other absence *away*, with the day of
return. A private absence shows only *away*.

Guests, kiosk displays and API tokens can read the widget where its tab is
shared with them, but cannot set a status.

## Widget

The widget lists everyone, or the members of one group chosen in its
settings, sorted by status and name, with the counts per status at the
bottom. It refreshes every 30 seconds while the page is visible.

In a small or strip tile the widget shows its compact view: *3 in the
office · 2 at home*, the number away below, and **Office**, **Home** and
**Away** as chips for your own status, the current one selected. Guests,
kiosk displays and an expired licence key get the counts without chips.

## Switching it off

Admins can switch presence off for everyone in *Settings → Presence*. While
it is off, the API answers `404` and all statuses are deleted. With an
expired licence key or too few seats, presence stays readable, but nobody
can set a status.

## API

| Method | Path | |
| --- | --- | --- |
| GET | `/api/presence?group=` | Today's `day`, `people` (`accountId`, `name`, `status` of `office`, `home`, `away` or `""`, `source` of `self`, `absence` or `none`, `returnOn`), `counts` and `mine` (`null` for guests and kiosk displays). Without `group`, everyone. |
| PUT | `/api/presence/me` | `{status, returnOn}`; `status` `""` clears it |
| GET | `/api/presence/compact?group=` | Compact state of the widget: counts in the office and at home, the number away, and the own status as chips (`action` with `choices`) |
| GET | `/api/presence/settings` | `{enabled}` |
| PUT | `/api/presence/settings` | `{enabled}`; admins only |
