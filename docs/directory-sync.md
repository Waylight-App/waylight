# Directory sync with SCIM

With a licence key that includes `directory-sync` ([licence keys](licence-keys.md)),
an identity provider can create, change and remove members and groups in
trisse-home over SCIM 2.0. People removed or disabled in the directory lose
access without anyone touching trisse-home.

This needs a server built with the `ee` features and members switched on.
Sign-in itself is separate: combine SCIM with [single sign-on for
members](single-sign-on-members.md), so the people it creates can sign in.

## Set it up

1. Pick a long random token and set it:

   ```sh
   SCIM_TOKEN=<at least 32 random characters>
   ```

   Without `SCIM_TOKEN`, every SCIM request is refused.

2. Make `https://home.example.org/scim/v2` reachable for the provider. A
   cloud provider such as Microsoft Entra ID connects from the internet: let
   your reverse proxy forward `/scim/` and allow the provider's addresses in
   `ALLOWED_CIDRS`. Nothing else under `/scim/` exists.

3. In the provider, enter the URL as *Tenant URL* or *SCIM base URL* and the
   token as *Secret token* or *Bearer token*.

## What the provider can do

| SCIM | In trisse-home |
| --- | --- |
| Create a user | A member account. Its name comes from `userName` (lowercase, shortened, a number added when taken); `userName` itself is kept for the provider. |
| `active: false` | The member is disabled and signed out. `active: true` enables it again. |
| Delete a user | The member is deleted like under *Settings → Members and groups*. |
| Create, rename, delete a group | The same for a trisse-home group. |
| Add or remove group members | The same. Only members can be in a group set by the provider. |

The account name stays as it was created when `userName` changes later.
Names, e-mail addresses and other attributes the provider sends are accepted
and not stored. The owner and guests are never touched.

Lookups support the filter `attribute eq "value"` on `userName`,
`externalId` and `displayName`, which is what providers use before creating
a user or group. Bulk requests, sorting and other filters are not supported.

If the licence key expires or covers fewer accounts than the instance has,
the provider can still read users and groups, but changes are refused until
a valid key is entered.

## Microsoft Entra ID

1. In *Enterprise applications*, open your trisse-home app (or create a
   non-gallery app) and choose *Provisioning → Get started*.
2. Set *Provisioning Mode* to *Automatic*, *Tenant URL* to
   `https://home.example.org/scim/v2` and *Secret token* to `SCIM_TOKEN`, then
   *Test Connection*.
3. Under *Mappings*, keep *userName*, *active*, *displayName* and
   *externalId*; the other attributes are ignored by trisse-home.
4. Assign the users and groups that should be provisioned and start
   provisioning.

This walkthrough follows Entra's documented SCIM behaviour (for example
sending `"False"` as text to deactivate, which trisse-home understands) but
has not been run against a real tenant yet; please report what you find.

## Other providers

Okta, Authentik and Keycloak (with a SCIM plugin) work the same way: base
URL `https://home.example.org/scim/v2`, bearer token `SCIM_TOKEN`.
