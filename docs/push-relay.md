# Push relay

The mobile app gets notifications through Google Firebase Cloud Messaging
(FCM) on Android and the Apple Push Notification service (APNs) on iOS.
Both only accept messages signed with the keys of the app's publisher, so
Waylight instances hand their messages to the push relay run by the
project at `push.getwaylight.app`. [ADR 0025](adr/0025-push-relay.md)
describes the protocol.

This page is for running the relay. Instances need nothing from it: they
sign their requests with a key they create themselves.

## What the relay sees

- The device token and the platform, to forward the message.
- A payload encrypted for the device. The relay cannot read it, and neither
  can Google or Apple.
- The public key of the sending instance and its network address, for the
  rate limits.

It stores only which instance key sent to a device token first (as a
SHA-256 hash of the token), so nobody else can send to that device, and
deletes that record after 90 days without use. Logs contain the status code
and the duration of each request.

## Run it

Build the image from the repository root:

```sh
docker build -f deploy/push-relay/Dockerfile -t waylight-push-relay .
```

Configure it through the environment:

| Variable | Meaning |
| --- | --- |
| `FCM_SERVICE_ACCOUNT_FILE` | Service account JSON of the Firebase project, with the role *Firebase Cloud Messaging API Admin*. |
| `APNS_KEY_FILE` | The `.p8` key for Apple push notifications. |
| `APNS_KEY_ID`, `APNS_TEAM_ID` | Key ID and team ID from the Apple Developer account. |
| `APNS_TOPIC` | Bundle ID of the iOS app. |
| `APNS_SANDBOX` | `true` for development builds of the app. |
| `TRUSTED_PROXIES` | Addresses of the reverse proxy in CIDR notation, so rate limits use the client's address. |
| `PUSH_RELAY_ADDR` | Listen address, default `:8080`. |
| `PUSH_RELAY_DB` | SQLite database of the token bindings, default `/data/push-relay.db`. |

At least one of FCM and APNs must be configured; a relay with FCM only
rejects messages for iOS devices. Mount the key files read-only, keep
`/data` on a volume, and put the relay behind a reverse proxy with TLS.
`GET /healthz` answers `204` while it runs.

## Limits

Per hour: 600 messages per instance key, 120 per device token and 1,200 per
client address. A request above a limit gets `429`; the instance tries again
later. The counters live in memory and start over after a restart.
