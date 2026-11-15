# Accounts, guests and kiosk displays

trisse-home has one owner, who can change everything. The owner can let
other people and devices look at selected tabs: guests sign in with their own
password, kiosk displays with a secret link. Neither can change anything.
This page covers all three and how to protect the owner account.

## The owner account

The account you create during setup is the owner. Under Settings → Account
you can:

- change the password (at least 12 characters);
- set up two-factor authentication with any authenticator app (TOTP). You
  get ten recovery codes, each usable once instead of a code. Download them
  and keep them outside trisse-home;
- add passkeys as a second factor, if trisse-home runs behind HTTPS with
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

## Members and groups (in preparation)

Members are full accounts for a second adult in the household or a colleague
in a small office: their own tabs and their own notes, to-dos and lists, with
the instance settings left to the admin (today the owner). Groups name a set
of members and guests, so a record can later be shared with "Office" or
"Family" instead of everyone. [ADR 0018](adr/0018-data-ownership.md)
describes the model.

The accounts, groups and the server-side rules are in place, but **members
cannot be created yet**: the modules do not separate personal content per
account so far, and a member would see and change your notes. They are
switched on once notes, to-dos and the other personal modules store an owner
per record (#294). Until then the server answers a new member with an
explanation instead of creating it.

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
`trisse-home.yaml` describes your (the admin's) tabs, never a member's.

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
4. Click *Create display*. trisse-home shows a link and a QR code **once**.
   Open the link on the device or scan the code with it.

The device signs in and switches to a clean full-screen view of the chosen
tabs. It keeps the screen awake while it is open, as far as the browser
allows. Buttons that would change something are hidden, and the server
refuses every change from a display anyway.

The secret sits after the `#` of the link, so it never appears in server or
proxy logs. If a device is lost or the link leaks, click **New link** (the old
link stops working at once) or delete the display. A display is subject to
the same network allow-list as everybody else.

## API tokens

Scripts, iOS Shortcuts and Home Assistant can talk to trisse-home with an API
token instead of a session. Each token is limited to the scopes you choose,
for example only adding to-dos. See [API tokens](api-tokens.md).
