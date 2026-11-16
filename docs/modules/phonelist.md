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

The search field of the start page answers too: type `phone anna`,
`telefon weber` or `nummer reception`, and the answer shows the number and
the address. Press Enter to copy the number.

## Outbound requests

None. The list is stored in the instance's database.

## API

`GET /api/phonelist/entries` (with `?q=` to search),
`POST /api/phonelist/entries`, `PUT` and `DELETE /api/phonelist/entries/{id}`.
Changes need an admin. See `api/modules/phonelist.yaml`.
