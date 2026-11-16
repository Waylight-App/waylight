# Branding

With a licence key that includes `branding` ([licence keys](licence-keys.md)),
admins give the instance their organisation's name, logos, favicon and
colours. This needs a server built with the `ee` features.

## What can be set

- **Name**, up to 40 characters: on the sign-in page, in the header, in the
  browser tab and in the app manifest (installed app, home screen).
- **Logo** with an optional **dark variant**, and a **favicon**: PNG, JPEG
  or GIF up to 2 MB. Every upload is decoded and stored again as a PNG of at
  most 512 pixels on the longer edge; SVG is not accepted, because a vector
  file can carry scripts. The images are kept in the database and are part
  of backups.
- **Accent** and **secondary colour** as `#rrggbb`. The settings warn when a
  colour has less than 4.5:1 contrast (WCAG AA) on the light or the dark
  theme; it can still be saved.
- **Hide "Powered by trisse-home"** on the sign-in page, in the footer and
  in the about section.

## Licence

Without a key that includes `branding`, the product's name and logo are
shown and the settings cannot be changed. When the key expires or is
removed, the stored branding is kept but no longer shown; with a valid key
it is back, nothing has to be entered again.

## API

Reading is public, because the sign-in page shows the branding:

| Method | Path | |
| --- | --- | --- |
| GET | `/api/branding` | `{name, custom, accent, secondary, hidePoweredBy, logo, logoDark, favicon}`; image fields are addresses, empty for the product's |
| GET | `/api/branding/{logo,logo-dark,favicon}` | The image; 404 without one |
| GET | `/manifest.webmanifest` | The app manifest with the organisation's name, colour and favicon |

Changing it is for admins:

| Method | Path | |
| --- | --- | --- |
| GET, PUT | `/api/branding/settings` | `{name, accent, secondary, hidePoweredBy}`; the answer adds `assets`, `applied` and `warnings` |
| PUT | `/api/branding/assets/{logo,logo-dark,favicon}` | The image as the request body |
| DELETE | `/api/branding/assets/{kind}` | Back to the product's image |

Not yet covered here: a default theme for new accounts, and the name and
logo in e-mails, the daily briefing and kiosk mode.
