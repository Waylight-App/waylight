# Accounts, guests and kiosk displays

Waylight has one owner, who can change everything. The owner can let
other people and devices look at selected tabs: guests sign in with their own
password, kiosk displays with a secret link. Neither can change anything.
This page covers all three and how to protect the owner account.

## The owner account

The account you create during setup is the owner. Under Settings → Account
you can:

- change the password (at least 12 characters);
- set up two-factor authentication with any authenticator app (TOTP). You
  get ten recovery codes, each usable once instead of a code. Download them
  and keep them outside Waylight;
- add passkeys as a second factor, if Waylight runs behind HTTPS with
  `PUBLIC_URL` set (see [reverse proxy](reverse-proxy.md#passkeys));
- see every signed-in session with its device and last use, and sign out a
  single session or all others;

Settings → Security log lists sign-ins, failed attempts, lockouts and
changes to accounts. With [notifications](notifications.md) set up, a sign-in
from a new device and a lockout are also pushed to your phone.

After five failed sign-ins from one address within 15 minutes, further
attempts from that address are refused for a while, and the wait grows with
every new failure.

If you already run an identity provider such as Authelia, Authentik, Keycloak
or Pocket ID, the owner can sign in through it instead. See
[single sign-on](reverse-proxy.md#single-sign-on-optional).

**Forgotten password:** see [getting started](getting-started.md#if-something-goes-wrong).
The reset needs shell access to the server, by design.

## Guests

A guest account lets someone in your household see the tabs you choose, say
the shopping list, the waste calendar and the next departures, without
being able to edit them and without knowing your password.

Create guests under Settings → Guests. Each guest has:

- a name of 2 to 32 characters (lowercase letters, digits, dots, dashes and
  underscores), which they type to sign in;
- a password you set and can reset later;
- a list of **shared tabs**. The guest sees only these, in the same order as
  yours.

Guests sign in on the normal login page with *Sign in as guest*. They can
change their own password, set up two-factor authentication and manage their
own sessions. Everything else is read-only: no edit mode, no settings, no
buttons that change data. You can disable a guest temporarily or delete it.
There can be up to 20 guests.

Two details are worth knowing:

- A guest sees the data of every widget on a shared tab. If a tab holds your
  calendar, the guest sees your appointments. Build a separate tab for
  guests rather than sharing your personal one.
- The server enforces the read-only rule centrally, for every request, not
  only by hiding buttons. New features are closed to guests until they are
  explicitly allowed. [ADR 0012](adr/0012-owner-and-guest-roles.md) has the
  details.

## Members and groups

Members are full accounts for a second adult in the household or a colleague
in a small office: their own tabs and their own notes, to-dos and lists, with
the instance settings left to the admin (today the owner). Groups name a set
of members and guests, so a record can later be shared with "Office" or
"Family" instead of everyone. [ADR 0018](adr/0018-data-ownership.md)
describes the model.

**Add a member** under *Settings → Members and groups*, either with a name
and a password you hand over, or with an invitation link (see below) so the
person picks the password. Members sign in on the sign-in page with
*Sign in as guest or member*.

Before the first member signs in, decide whether your own entries stay
visible to everyone on the instance: the switch *My new entries are private*
in the same section makes your new notes, to-dos and lists private by
default; existing ones keep their visibility until you change it.

Features for organisations in `ee/` are not part of this: they need a
licence key, and without one no account, admin or member, can use them.

What a member may do, enforced for every request like the guest rules:

- read and change modules with personal content (notes, to-dos, bookmarks,
  lists) and their own personal settings (theme, search shortcuts);
- read instance data such as services or the weather, but not settings pages
  of instance modules, which can name hosts and keys;
- nothing in account management, backups, the export, notification channels,
  Docker or the system pages. A module without a declared kind stays closed
  for changes by members.

Every account has its own dashboard tabs: a member sees only their own, you
see only yours, and guests keep seeing exactly the tabs you shared with them.
`waylight.yaml` describes your (the admin's) tabs, never a member's.

Account and group changes are written to the security log.

### Sharing notes and to-dos

A note or a to-do belongs to the account that created it, and only that
account can decide who else sees it. The note editor and the to-do dialog
have a *Who can see this* choice:

- **Only me**: nobody else sees the record.
- **Group: …**: the members and guests of that group can read it. You can
  pick the groups you belong to; an admin can pick any group.
- **Everyone on this dashboard**: every account on the instance can read it.

Shared records are read-only for everyone but the owner. This also applies to
admins: an admin sees another account's notes and to-dos only when they are
shared with the admin, a group the admin is in, or everyone.

If a group is deleted, the records shared with it become private again. The
choice then shows *A group that no longer exists* until the owner picks
something else. To-dos from an external list (CalDAV or another provider)
cannot be shared and have no sharing choice.

### Link directories for groups

A bookmark folder can be shared the same way. In *Settings → Bookmarks*,
pick who can see a folder: only you, a group or everyone. All
bookmarks in the folder follow that choice, including ones you add later.

Shared with a group, the folder becomes that group's link directory: its
members see the links in the launcher and the search next to their own, and
in the bookmarks widget marked with 👥. Only the folder's owner can change, add
or remove links; for everyone else they are read-only, and nobody can file
their own bookmarks into someone else's folder. Personal bookmarks outside
shared folders stay private.

## Admins

The owner can share the management of the instance: under *Settings →
Members and groups*, switch a member to **Admin**. Admins manage members,
guests, groups, kiosk displays and the instance settings like you, and keep
their own tabs and private notes like any member. You can have as many as
you like; [ADR 0022](adr/0022-several-admins.md) has the full rules.

What stays with you as the owner:

- your own account: no admin can disable, delete or demote it, or change
  its password or second factor;
- backups, restoring a backup and the data export and import, because they
  hold everyone's data;
- API tokens, because they act as the owner.

An admin cannot demote, disable or delete their own account; you or another
admin do that, and a demoted admin is signed out. Admins get no password
reset link by e-mail: you or another admin set a new password for them.

With single sign-on or a directory, admins can also come from a group of
the identity provider (`OIDC_ADMIN_GROUPS`, `LDAP_ADMIN_GROUPS`); see
[single sign-on for members](single-sign-on-members.md) and
[directory sync](directory-sync.md).

## Invitations and reset links

Instead of choosing a password for someone, you can send them a link.

**Invite:** *Settings → Invitations* creates a link for a new guest (with the
tabs it may see) or, once members are available, a new member. You pick the
name and how long the link waits: one day, a week or 30 days. The invited
person opens the link, sets a password and then signs in with the name and
that password. Open invitations are listed there and can be revoked.

**Reset a password:** *Reset link* next to a member or guest creates a link
that sets a new password. It is valid for 24 hours, and a new reset link
replaces the previous one. Using it signs the account out on all devices.

Both kinds of link:

- are shown once, as text and as a QR code; send them on a channel you trust;
- work once and stop working when they expire;
- carry the secret after `#`, so it is not written to server or proxy logs;
- are limited like the sign-in: repeated wrong links lock the client out for
  a while;
- appear in the security log when they are created and used.

The link does not sign anyone in. After setting the password, the person
signs in normally, including the second factor if one is set up.

**Reset by e-mail:** members and guests can store an e-mail address under
*Settings → Account → E-mail for password reset*. On the sign-in page,
*Sign in as guest or member → Forgot your password?* then sends a reset link
to that address. This needs two things on the instance:

- an enabled e-mail channel under *Settings → Notifications*; its mail
  server sends the reset mails, but its own recipients get nothing;
- `PUBLIC_URL`, because the link in the mail is built from it and never
  from the address the request came in on.

The mailed link is valid for one hour, and an account gets at most one such
mail every five minutes. The sign-in page answers the same whether or not
the name exists or has an address, so it does not tell anyone which accounts
exist. The address is used for nothing else and does not appear in the
security log.

There is no reset link for the admin account. Its password is reset on the
server, as described above.

## Kiosk displays

A kiosk display shows tabs on a device that hangs on the wall or stands in
the kitchen, such as an old tablet or a Raspberry Pi with a screen. It needs
no password and shows no controls.

Create one under Settings → Kiosk displays:

1. Give it a name (for example `hallway-tablet`) and pick the tabs it shows.
2. Optionally set a **tab rotation**: the display moves to the next tab every
   few seconds.
3. Optionally **dim at night** between two times.
4. Click *Create display*. Waylight shows a link and a QR code **once**.
   Open the link on the device or scan the code with it.

The device signs in and switches to a clean full-screen view of the chosen
tabs. It keeps the screen awake while it is open, as far as the browser
allows. Buttons that would change something are hidden, and the server
refuses every change from a display anyway.

The secret sits after the `#` of the link, so it never appears in server or
proxy logs. If a device is lost or the link leaks, click **New link** (the old
link stops working at once) or delete the display. A display is subject to
the same network allow-list as everybody else.

## Apps

The [Waylight app](https://github.com/Waylight-App/waylight-mobile) for
Android and iOS signs in with an **app session** of its own. Every account
except kiosk displays can use it.

To sign a phone in:

1. Open *Settings → Mobile app* in the browser and choose
   **Show pairing code**.
2. Scan the QR code with the app. The code works once and for five
   minutes; **New code** replaces it.

The app can also sign in with the password and, if set up, the code of the
authenticator app.

The phone then appears under *Active sessions* with its device name, for
example "App on Pixel 9". It stays signed in for 90 days after its last use.
Sign it out there like any other session; changing the password signs out
all apps as well.

For developers: the app sends its token as `Authorization: Bearer wa_…`
and needs no CSRF token. The endpoints are `POST /api/auth/app-sessions`
and `POST /api/auth/app-sessions/pairing-code` (see the API reference);
the design is in [ADR 0024](adr/0024-app-sessions.md).

## API tokens

Scripts, iOS Shortcuts and Home Assistant can talk to Waylight with an API
token instead of a session. Each token is limited to the scopes you choose,
for example only adding to-dos. See [API tokens](api-tokens.md).
