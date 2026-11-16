# Directory sync: SCIM, LDAP and Active Directory

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

## Sign in with LDAP or Active Directory

Members can also sign in with their directory password on the normal sign-in
page (*Sign in as guest or member*). The first sign-in creates the member;
their groups follow `memberOf` at every sign-in, as with single sign-on.

```sh
# ldaps:// or ldap:// with StartTLS; unencrypted LDAP is never used.
LDAP_URL=ldaps://dc1.example.org
# A service account that may search users.
LDAP_BIND_DN=cn=trisse-home,ou=services,dc=example,dc=org
LDAP_BIND_PASSWORD=<password>
LDAP_BASE_DN=ou=people,dc=example,dc=org
# The attribute with the login name: uid (OpenLDAP) or sAMAccountName (AD).
LDAP_USER_ATTRIBUTE=uid
# Optional: only users in one of these groups (names or DNs) get in.
LDAP_MEMBER_GROUPS=trisse-home-users
# Optional: only these groups become trisse-home groups.
LDAP_SYNC_GROUPS=Office,Sales
# Optional: a CA certificate (PEM) for a directory with an internal CA.
LDAP_CA_FILE=/data/ldap-ca.pem
```

How names are checked:

- The owner always signs in with their local password. A directory that is
  down never locks the owner out.
- A name the directory knows is checked with a bind as that user. A name it
  does not know, such as a local guest, is checked locally as before.
- While the directory cannot be reached, local accounts still sign in;
  members created from the directory cannot, because only the directory knows
  their password.
- Someone removed from the directory, disabled in Active Directory or no
  longer in `LDAP_MEMBER_GROUPS` is refused and their member account is
  disabled; their data stays. When they are back, the account is enabled
  again, unless an admin disabled it.

### OpenLDAP

```sh
LDAP_URL=ldap://ldap.example.org        # StartTLS
LDAP_BIND_DN=cn=readonly,dc=example,dc=org
LDAP_BASE_DN=ou=people,dc=example,dc=org
LDAP_USER_ATTRIBUTE=uid
```

`memberOf` needs the memberof overlay; without it users sign in, but no
groups are synced.

### Active Directory

```sh
LDAP_URL=ldaps://dc1.example.org
LDAP_BIND_DN=CN=trisse-home,OU=Service Accounts,DC=example,DC=org
LDAP_BASE_DN=OU=Staff,DC=example,DC=org
LDAP_USER_ATTRIBUTE=sAMAccountName
LDAP_USER_CLASS=user
LDAP_CA_FILE=/data/ad-ca.pem             # the domain's root CA
```

Group names are taken from the first part of each group's DN
(`CN=Sales,OU=Groups,…` becomes *Sales*). Nested groups are not resolved:
only groups the user is a direct member of count.

The LDAP client is part of trisse-home (standard library only) and has been
tested against a test server, not yet against a production directory;
please report what you find.
