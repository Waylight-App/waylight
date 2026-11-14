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
