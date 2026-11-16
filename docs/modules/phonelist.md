# Phone list

**What it does:** a searchable directory of the people in your organisation
or household, with name, role, team, phone number and e-mail address. Numbers
and addresses are links, so a click calls or writes.

## Using it

Add the *Phone list* widget from the *Everyday* category. Admins add people
with ＋ and change or remove them with ✎. Everyone else who can see the
widget reads the list.

Each entry needs a name and at least a phone number or an e-mail address.
Role and team are optional and searchable.

The ⧉ button next to a number copies it, which helps on a computer without
a phone app. In the widget settings, *Group by team* shows a heading per
team, with people without a team last under *Other*. Admins can download
the list as a CSV file (⤓), with the columns name, role, team, phone and
e-mail; it opens in spreadsheet apps with umlauts intact.

The search field of the start page answers too: type `phone anna`,
`telefon weber` or `nummer reception`, and the answer shows the number and
the address. Press Enter to copy the number.

## Address book

Admins can also read people from a CardDAV address book, for example a
shared address book in Nextcloud: *Settings → Phone list*, with the address
book URL, a user name and an app password. The book is read once when you
save, to check it, and then every hour. Its people appear next to the
entries typed in by hand, marked with 📇, and can only be changed in the
address book. Each card needs a name and a phone number or an e-mail
address. The list takes the formatted name, the title (or role), the unit
of the organisation as team, and the preferred phone number and address.

If a sync fails, the people from the last successful one stay, and the
settings show the error. Clearing the URL switches the book off and
removes its people from the list.

## Outbound requests

Only with an address book: one CardDAV `REPORT` to the configured URL every
hour, with the stored user name and app password. The password is encrypted
in the database and left out of exports.

## API

`GET /api/phonelist/entries` (with `?q=` to search),
`POST /api/phonelist/entries`, `PUT` and `DELETE /api/phonelist/entries/{id}`,
`GET` and `PUT /api/phonelist/settings`. Changes need an admin. See `api/modules/phonelist.yaml`.
