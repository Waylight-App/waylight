# ADR 0022: Several admins next to the owner

- Status: Accepted (2026-11-16, by the maintainer)
- Date: 2026-11-16
- Issue: [#434](https://github.com/itadori26/waylight/issues/434), part of [#292](https://github.com/itadori26/waylight/issues/292)
- Builds on: [ADR 0012](0012-owner-and-guest-roles.md), [ADR 0018](0018-data-ownership.md)

## Context

ADR 0018 calls the owner "the first admin" and keeps every management task
with that one account. An organisation, and a household with two adults,
needs more than one person to manage members, guests and settings. Members
can come from single sign-on or a directory, where the people who manage an
instance are usually already in a group.

The owner account is special in two ways that a second admin must not
inherit: its ID doubles as the instance's (instance data, API tokens and
background jobs act under it), and its password is reset only on the server.

## Decision

### Roles

- **owner**: exactly one, created at setup. Unchanged.
- **admin**: any number. A person like a member (own tabs, own personal
  content that starts private, can be in groups), who also manages the
  instance.
- **member**, **guest**, **kiosk**: unchanged.

### What an admin may do

Everything the owner may, with these exceptions:

- **Not on the owner.** The owner is not a member account, so no member
  operation reaches it: an admin cannot disable, delete or demote the owner,
  nor change the owner's password or second factor.
- **Not on themselves.** An admin cannot demote, disable or delete their own
  account; the owner or another admin does that. So nobody locks themselves
  out by mistake.
- **Owner alone:** backups (create, download, restore, remote backups), the
  data transfer (export, import, configuration) and API tokens. Backups and
  the transfer contain or replace the data of everyone, private records and
  the owner's account included; API tokens act under the owner's ID.
- **No reset links for admins.** Like the owner's, an admin account must not
  be taken over through a link; the owner or another admin sets a new
  password.

Admins see the same records as everyone: their own, those shared with
everyone and those of their groups. Being an admin gives no access to
other people's private records.

### Who makes someone an admin

- **By hand (core):** the owner and admins make a member an admin and an
  admin a member under *Settings → Members and groups*. A demoted admin's
  sessions end.
- **From the identity provider (`ee/`):** `OIDC_ADMIN_GROUPS` and
  `LDAP_ADMIN_GROUPS` name groups whose people become admins at sign-in. The
  account remembers where the role came from (`role_source`: empty by hand,
  `sso` or `directory`). A provider takes back only a role it granted, when
  the person is no longer in the group at their next sign-in; an admin made
  by hand is never demoted by a provider. With an expired licence key, roles
  are not changed by providers, like groups.

Every role change is written to the security log.

### Seats

Admins count as seats like members ([licence keys](../licence-keys.md)).

## Consequences

- `IsAdmin` means owner or admin everywhere; code that needs the owner
  checks the role `owner` explicitly. Places that meant "a person with a
  space of their own" check member or admin.
- The API reports the role `admin` in the status and in the member list;
  the settings show it and offer the switch.
- A new management feature is open to admins by default. One that contains
  everyone's data or acts as the owner adds its routes to the owner-only
  list in the auth module and to this ADR.
