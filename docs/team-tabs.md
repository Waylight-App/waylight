# Team tabs

With a licence key that includes `team-tabs` ([licence keys](licence-keys.md)),
an admin shares one of their own dashboard tabs with a group or with everyone.
The tab then appears for those people next to their own tabs, marked with 👥.
This needs a server built with the `ee` features.

## Sharing a tab

1. Build the tab on your own dashboard: name, widgets, layout, background.
2. Open *Settings → Team tabs* and choose who else sees it: *Only me*, a
   group, or *Everyone*.

Change the tab like any of your own; everyone who sees it gets the change
right away. Choosing *Only me* makes it private again.

## A default tab for a group

Under *Default tab per group* in the same section, pick the team tab that
opens first for each group. It only applies when the person's own tab rules
(time of day, network) pick no tab. Someone in several groups gets the
default of the first group, in a fixed order, that has one.

## What the others can do

- They see the tab and use its widgets as usual: notes, to-dos and other
  content follow the sharing rules of their own module.
- They cannot rename, move, delete or rearrange it. In edit mode the tab
  stays as it is and a hint says who maintains it; their own tabs can still be
  arranged.
- Guests and kiosk displays keep seeing exactly the tabs shared with them in
  their own settings.

## Without a licence

When the licence no longer includes team tabs, nobody but the owner sees
shared tabs, and a build without the `ee` features never shows them. With
an expired key shared tabs stay visible, but their sharing cannot be changed.

## API

`GET /api/team-tabs` (admins: own tabs with their visibility),
`PUT /api/team-tabs/{id}` with `{"visibility": "private" | "everyone" |
"group:<id>"}`, `GET /api/team-tabs/defaults` (group ID → tab ID) and
`PUT /api/team-tabs/defaults/{groupId}` with `{"tabId": "…"}` (empty removes
it). The dashboard (`GET /api/dashboard`) marks tabs shared with
the viewer with `"mine": false`.
