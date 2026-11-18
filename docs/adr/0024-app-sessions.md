# ADR 0024: App sessions for the mobile app

- Status: Proposed
- Date: 2026-11-18
- Issues: [#559](https://github.com/Waylight-App/waylight/issues/559)

## Context

The mobile app ([Waylight-App/waylight-mobile](https://github.com/Waylight-App/waylight-mobile))
signs in through the browser session endpoints: a cookie plus the
`X-CSRF-Token` header. Browser sessions expire after the browser session
lifetime, do not name the device and need cookie handling that is built for
browsers. Personal API tokens are no alternative: they are scoped per
module and cannot reach `/api/auth`.

## Decision

1. **An app session is a session with a device name.** It lives in the
   `auth_sessions` table next to browser sessions, so the session overview,
   revoking, "sign out everywhere", disabling an account and the admin's
   maximum session age apply without extra code.
2. **Bearer token with its own prefix.** The token starts with `wa_` and is
   sent as `Authorization: Bearer …`. The guard recognises the prefix
   without a database lookup and checks the session like a cookie session,
   but without the CSRF token: a browser never attaches a bearer token on
   its own, so a cross-site request cannot use it. Only the SHA-256 hash is
   stored, like for every session.
3. **Lifetime.** 90 days, extended on use (at most once an hour, like
   browser sessions). A phone is opened less regularly than a browser tab.
4. **Two ways in.** `POST /api/auth/app-sessions` takes the fields of the
   login (password, second factor; same rate limit, lockout and security
   log) or a **pairing code**. The code is created by a signed-in account
   under *Settings → Mobile app* and shown as QR code with the link
   `waylight://pair?server=…&code=…`. It is stored hashed, valid for five
   minutes and redeemed once; failed redemptions count like failed logins.
   Kiosk displays cannot create codes.

## Consequences

- Accounts with only a passkey cannot sign the app in with a password; they
  pair it with the QR code. A browser handover for passkeys, Microsoft and
  single sign-on follows in #560.
- A stolen phone keeps access until the session is revoked in the session
  overview, where it is listed with its device name.
- The design can later grow into the OAuth 2.0 device authorization flow if
  other clients need it.
