# Integrations

**What it does:** an admin saves the address and the credential of a
service once, as a named connection such as "Home" or "Office", and widgets
of that service choose it. A service can have several connections, for
example two Proxmox clusters.

## Status

This is the store and its API. Widgets and the settings pages of the
integrations do not use it yet; their own settings stay as they are until
they move to connections (#610).

## Rules

- Only admins see and change connections; they belong to the instance.
- A name is unique per service, compared without case. An instance keeps
  at most 100 connections.
- The credential is a set of named fields (for example `apiKey`, or
  `username` and `password`), at most four. It is stored encrypted with
  `SECRET_KEY` and never returned: the API only says which fields are set.
- Moving a connection to another address needs the credential again, so a
  stored one is never sent to a new host.
- Removing a connection that widgets name in their settings
  (`connectionId`) is refused until it is confirmed; the list shows which
  widgets these are.
- If `SECRET_KEY` changes, the credentials can no longer be read. The list
  still opens and shows no fields; enter the credentials again.

## Config as code

The `integrations` section of [waylight.yaml](../config-as-code.md) lists
service, name and address. It never contains a credential. While the file
manages the section, connections cannot be added, renamed or removed in the
app, but credentials can still be entered.

## API

`GET /api/integrations/connections` (optionally `?service=immich`)
lists the connections with `fields` and `usedBy`; `POST /api/integrations/connections`,
`PUT /api/integrations/connections/{id}`,
`PUT /api/integrations/connections/{id}/credential` and
`DELETE /api/integrations/connections/{id}?confirm=true` change them. See
`api/modules/integrations.yaml`.

## Outbound requests

None.
