# Birthdays

**What it does:** Shows the next birthdays, with the age when the year of birth is known. Entries come from three places: typed in by hand, imported from a vCard file, or read from a CardDAV address book. A push reminder on the day and a line in the daily briefing are optional (#287).

## Configuration

Add the **Birthdays** widget. Its **+** button adds a person by hand; the year can stay empty.

Under **Settings → Birthdays** you can:

- **import a vCard file** (`.vcf`, as exported by most phones, Google Contacts, Outlook or Thunderbird). Only contacts with a birthday are taken; a person with the same name and day that is already in the list is skipped, so importing the same file twice does no harm.
- **connect a CardDAV address book**, for example Nextcloud (`https://cloud.example.org/remote.php/dav/addressbooks/users/<name>/contacts/`). Use an app password. The address book is read when you save and then every six hours; its entries replace the previously synced ones and are changed or removed in the address book itself, not in trisse-home.
- **turn on the reminder**: at the chosen time each day, one push per birthday through the notification channels.

The daily briefing gets a **Birthdays** section with today's and tomorrow's birthdays; it can be switched off or moved in the briefing's settings like the other sections.

## Outbound requests

Only to the CardDAV server you enter: one `REPORT` request when you save and every six hours. Nothing goes out until an address book is connected. The password is stored encrypted and never sent to the browser.

## Limits

At most 2,000 entries. A birthday on 29 February is shown on 28 February in other years. Contacts without a name or without a valid `BDAY` are skipped.

## API

`GET /api/birthdays?limit=5`, `POST /api/birthdays`, `DELETE /api/birthdays/{id}`, `POST /api/birthdays/import`, `GET`/`PUT /api/birthdays/settings` – see `api/modules/birthdays.yaml`.
