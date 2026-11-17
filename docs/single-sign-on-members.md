# Single sign-on for members

With a licence key that includes `sso` ([licence keys](licence-keys.md)),
everyone your identity provider allows can sign in through it, not only the
owner. A person gets a member account on their first sign-in, and their
groups follow a claim of the ID token. Without the key, single sign-on stays
the owner's, as described in [single sign-on](reverse-proxy.md#single-sign-on-optional).

This needs a server built with the `ee` features and members switched on.

## Who gets in

Set the OpenID Connect variables for the owner first, then add:

```sh
# People with a verified address in this domain, or these addresses.
OIDC_MEMBER_EMAILS=@example.org,freelancer@mail.example
# Or everyone in one of these groups of the provider.
OIDC_MEMBER_GROUPS=waylight-users
# The claim that lists a person's groups.
OIDC_GROUPS_CLAIM=groups
# Optional: only these provider groups become waylight groups.
OIDC_SYNC_GROUPS=Office,Sales
```

The owner keeps signing in through `OIDC_ALLOWED_EMAILS` or
`OIDC_ALLOWED_SUBJECTS`. Everyone else who matches `OIDC_MEMBER_EMAILS` or
`OIDC_MEMBER_GROUPS` becomes a member. The account name comes from
`preferred_username`, else from the e-mail address; a taken name gets a
number.

## What happens at each sign-in

- **Groups:** the person is added to every group named in the claim (groups
  that do not exist yet are created) and removed from groups that the
  provider gave them earlier but no longer does. Groups an admin added by hand
  stay.
- **Access lost:** someone the provider no longer lets in is refused, their
  member account is disabled and their open sessions end. When the provider
  lets them in again, the account is enabled again, unless an admin disabled
  it.
- **Licence key expired or too few seats:** existing members still sign in,
  but no new accounts are created and groups are not changed.

- **Admins:** with `OIDC_ADMIN_GROUPS=waylight-admins` (comma-separated),
  everyone in one of these provider groups becomes an admin at sign-in
  ([admins](accounts.md#admins)). Someone who left the group is a member
  again at their next sign-in. An admin you made by hand stays one whatever
  the provider says. The group must be in the groups claim, like the groups
  above.

The owner is never changed by the provider. Changes at the provider take effect
when the person next signs in; to lock someone out at once, disable them
under *Settings → Members and groups*.

## Microsoft Entra ID

1. In the Entra admin center, open *App registrations → New registration*.
   Choose *Web* and enter `https://home.example.org/api/auth/oidc/callback`
   as redirect URI.
2. Under *Certificates & secrets*, create a client secret.
3. Under *Token configuration*, add the *groups* claim for ID tokens and
   choose *Groups assigned to the application*. By default Entra puts the
   groups' object IDs into the claim; Waylight then names its groups
   after these IDs, and `OIDC_MEMBER_GROUPS` and `OIDC_SYNC_GROUPS` take the
   IDs too. Where Entra offers to emit group names instead (for groups synced
   from Active Directory, and for cloud-only groups in newer tenants), the
   claim and the variables use names. This guide has not been tested against
   a real tenant yet; please report what you find.
4. Under *Enterprise applications*, open the app, set *Assignment required*
   and assign the groups that may use Waylight.
5. Set:

   ```sh
   OIDC_ISSUER=https://login.microsoftonline.com/<tenant-id>/v2.0
   OIDC_CLIENT_ID=<application (client) id>
   OIDC_CLIENT_SECRET=<secret value>
   OIDC_REDIRECT_URL=https://home.example.org/api/auth/oidc/callback
   OIDC_GROUPS_CLAIM=groups
   OIDC_MEMBER_GROUPS=<object ID or name of the assigned group>
   ```

Entra marks addresses of your own verified domains as verified; with
`OIDC_MEMBER_GROUPS` the e-mail check is not needed at all.

## Google Workspace

Google's ID tokens carry no groups. Use the e-mail domain instead:

1. In the Google Cloud console, open *APIs & Services → Credentials →
   Create credentials → OAuth client ID*, type *Web application*, with the
   redirect URI `https://home.example.org/api/auth/oidc/callback`.
2. On the *OAuth consent screen*, choose *Internal*, so only people of your
   Workspace can sign in.
3. Set:

   ```sh
   OIDC_ISSUER=https://accounts.google.com
   OIDC_CLIENT_ID=<client id>
   OIDC_CLIENT_SECRET=<client secret>
   OIDC_REDIRECT_URL=https://home.example.org/api/auth/oidc/callback
   OIDC_MEMBER_EMAILS=@your-workspace-domain.example
   ```

Groups are then maintained in Waylight itself. If you need provider
groups, put an identity provider that adds a groups claim, such as
Authentik or Keycloak, in front of Google.
