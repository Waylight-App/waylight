# Admin policies

With a licence key that includes `policies` ([licence keys](licence-keys.md)),
admins set rules that hold for every account. This needs a server built with
the `ee` features.

| Policy | Effect |
| --- | --- |
| Require a second factor | An account without an authenticator app or a passkey is asked to set one up at its next request. Until then it can only reach its own account settings; `/api/auth/status` reports `secondFactorRequired`. Kiosk displays are exempt. |
| Minimum password length | 12 to 64 characters, applied wherever a password is set: setup, password change, new members and guests, invitation and reset links. Existing passwords keep working until they are changed. |
| Maximum session length | Up to 2160 hours (90 days) after sign-in, even while the session is in use. Without it, sessions slide as long as they are used. |
| Modules for members | Members only see and use the listed modules; the rest disappears with its widgets, and their API answers 403. The dashboard, appearance, search and the member's own account always stay. Admins keep every module. |

Admins read and change the policies at `GET` and `PUT /api/policies`:

```json
{
  "requireSecondFactor": true,
  "minPasswordLength": 16,
  "maxSessionHours": 12,
  "memberModules": ["notes", "todos", "bookmarks", "phonelist"]
}
```

`memberModules: null` allows all modules. Every change is recorded in the
security log.

If the licence key expires or covers too few accounts, the policies keep
applying and can no longer be changed until a valid key is entered. Without
a key, no policy applies.
