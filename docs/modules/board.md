# Family board

**What it does:** Short notes for everyone in the household, which can be pinned and can expire on a chosen day.

## Using it

Add the *Family board* widget and write a note with **+ New note**. A note
holds up to 500 characters. Pin the ones that should stay at the top, such as
the guest Wi-Fi, and give short-lived ones a last day: from the day after,
the note is no longer shown and is deleted within the hour.

Pinned notes come first, then the newest. The compact view counts the notes
and shows the first one.

## Who sees what

Notes are shared with everyone on the instance by default. When you write a
note you can narrow it to one of your groups or to yourself. Only the person
who wrote a note can change, pin or remove it. Guests see the notes shared
with everyone when the widget is on one of their tabs, which makes the board
work on a wall display with a kiosk link.

## Configuration

No environment variables. `MODULES_DISABLED=board` switches the module off.

## Outbound requests

None; the module only uses the database.

## API

`/api/board` (list, create, update, delete) and `/api/board/compact` – see
`api/modules/board.yaml`.
