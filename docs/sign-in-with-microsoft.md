# Sign in with a personal Microsoft account

You can sign in to Waylight with your personal Microsoft account
(outlook.com, hotmail.com, live.com) instead of your password. You link the
Microsoft account to your account here once, while you are signed in; after
that, **Sign in with Microsoft** on the sign-in page is enough.

The owner can always do this. Members and guests need a licence key that
includes `sso` ([licence keys](licence-keys.md)).

## Without any setup: sign in by code

Waylight comes with its own registration at Microsoft, so there is
nothing to set up, and no HTTPS address is needed: it works under any
address, in your home network and over a VPN.

1. Sign in with your password as usual.
2. Open *Settings → Account → Sign in with Microsoft* and choose **Connect
   Microsoft**.
3. A window opens at Microsoft and Waylight shows a short code, already
   copied to the clipboard. Enter it, sign in at Microsoft and confirm.
4. The window closes by itself; Microsoft is linked.

From then on, **Sign in with Microsoft** on the sign-in page works the same
way: a window, the code, confirm. If you are already signed in at Microsoft
in that browser, it is a matter of seconds.

Microsoft shows Waylight on its confirmation page and asks only for your
name and e-mail address; Waylight gets no access to mail, files or
anything else. `MICROSOFT_CLIENT_ID` replaces the built-in registration with
your own, and an empty value turns signing in by code off
([configuration](configuration.md)).

Signing in by code takes personal accounts only. For work or school
accounts, see [With a work or school account](#with-a-work-or-school-account)
below.

The rest of this page sets up the other way, a sign-in that sends you to
Microsoft and back. It needs your own registration and an HTTPS address.

## What you need

- Waylight reachable under a fixed address, such as
  `https://home.example.org`. The address only has to work in your own
  browser; it does not have to be reachable from the internet.
- A Microsoft account to register Waylight with Microsoft. Your personal
  account works for that too; registering is free.

## Register Waylight with Microsoft

1. Open the [Microsoft Entra admin center](https://entra.microsoft.com) with
   your Microsoft account and go to *App registrations → New registration*.
2. Name: `waylight`. *Supported account types*: **Personal Microsoft
   accounts only**.
3. *Redirect URI*: platform **Web**, address
   `https://home.example.org/api/auth/oidc/callback` (your address).
4. After *Register*, copy the *Application (client) ID*.
5. Under *Certificates & secrets → New client secret*, create a secret and
   copy its *Value* right away; it is shown only once. Note when it expires
   and renew it before then.

## Configure Waylight

Add to `.env` and restart Waylight:

```sh
OIDC_ISSUER=https://login.microsoftonline.com/9188040d-6c67-4c5b-b112-36a304b66dad/v2.0
OIDC_CLIENT_ID=<application (client) id>
OIDC_CLIENT_SECRET=<secret value>
OIDC_REDIRECT_URL=https://home.example.org/api/auth/oidc/callback
OIDC_BUTTON_LABEL=Microsoft
OIDC_BUTTON_ICON=microsoft
```

The long ID in `OIDC_ISSUER` is the same for everyone: it stands for
personal Microsoft accounts. `OIDC_ALLOWED_EMAILS` is not needed. Microsoft
does not mark the address of a personal account as verified, so Waylight
does not let anyone in by address; it only trusts the link you make in the
next step.

## Link your Microsoft account

1. Sign in with your password as usual.
2. Open *Settings → Account → Sign in with Microsoft* and choose **Connect
   Microsoft**.
3. Sign in at Microsoft and confirm. You come back to the settings, which
   say that Microsoft is linked.

From now on, **Sign in with Microsoft** on the sign-in page signs you in. If
you are already signed in at Microsoft in that browser, it takes one click.

**Disconnect** in the same place removes the link; you then sign in with
your password again. One Microsoft account can be linked to one account
here only.

## Invite someone who signs up with Microsoft

With a licence key that includes `sso`, the invitation page offers
**Create account with Microsoft** next to the password form
([accounts](accounts.md)). The invited person signs in at Microsoft and comes
back signed in, with the name, role and tabs of the invitation and their
Microsoft account linked. The invitation is used up as with a password.

Such an account has no password at first, so Microsoft is its only way in:
*Disconnect* stays unavailable until a password is set, for example with a
reset link from an admin or *Forgot password* on the sign-in page.

## With a work or school account

Accounts of an organisation (Microsoft Entra ID) sign in through the
organisation's own registration, so its administrators keep control of who
may sign in, multi-factor rules and conditional access. Follow the steps
above with two changes:

- *Supported account types*: **Accounts in this organizational directory
  only**.
- `OIDC_ISSUER` names the organisation's directory instead of personal
  accounts:

  ```sh
  OIDC_ISSUER=https://login.microsoftonline.com/<directory (tenant) id>/v2.0
  ```

  The *Directory (tenant) ID* is on the overview page of the registration.

Microsoft does not mark addresses as verified for work accounts either, so
here too people get in through the link they make while signed in, as for
a personal account, not through `OIDC_ALLOWED_EMAILS`.

## Without a password

Once the link works, you can turn the password sign-in off with
`PASSWORD_LOGIN=false`, and with `OIDC_AUTO_REDIRECT=true` opening
Waylight goes straight to Microsoft
([configuration](configuration.md)). `/?local=1` always shows the sign-in
page, for example when Microsoft cannot be reached. Keep the recovery
instructions for the owner's account at hand
([accounts](accounts.md)).
