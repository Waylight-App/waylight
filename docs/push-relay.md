# Push relay

The mobile app gets its notifications from the push relay run by the
project at `push.getwaylight.app`. Waylight instances hand their encrypted
messages to it; Android phones keep a Server-Sent Events stream open to the
relay and receive them directly, and iOS phones will be reached through the
Apple Push Notification service (APNs), which only accepts messages signed
with the publisher's key. [ADR 0025](adr/0025-push-relay.md) describes the
protocol for instances, [ADR 0027](adr/0027-push-over-server-sent-events.md)
the stream.

This page is for running the relay. Instances need nothing from it: they
sign their requests with a key they create themselves.

## What the relay sees

- The device token and the platform, to forward the message.
- A payload encrypted for the device. The relay cannot read it, and neither
  can Apple for iOS devices.
- The public key of the sending instance and its network address, for the
  rate limits.
- For Android, the network address of each phone while its stream is open.

It stores which instance key sent to a device token first (as a SHA-256
hash of the token), so nobody else can send to that device, and deletes
that record after 90 days without use. Messages for Android phones wait in
a mailbox until the phone has received them, at most 100 per phone and
72 hours. Logs contain the status code and the duration of each request,
never device tokens, keys, addresses of phones or payloads.

## Run it

Build the image from the repository root:

```sh
docker build -f deploy/push-relay/Dockerfile -t waylight-push-relay .
```

Configure it through the environment:

| Variable | Meaning |
| --- | --- |
| `APNS_KEY_FILE` | The `.p8` key for Apple push notifications. |
| `APNS_KEY_ID`, `APNS_TEAM_ID` | Key ID and team ID from the Apple Developer account. |
| `APNS_TOPIC` | Bundle ID of the iOS app. |
| `APNS_SANDBOX` | `true` for development builds of the app. |
| `TRUSTED_PROXIES` | Addresses of the reverse proxy in CIDR notation, so rate limits use the client's address. |
| `PUSH_RELAY_ADDR` | Listen address, default `:8080`. |
| `PUSH_RELAY_DB` | SQLite database of the token bindings and the mailbox, default `/data/push-relay.db`. |

The stream for Android needs no configuration. APNs is optional; without
it the relay rejects messages for iOS devices. Mount the APNs key read-only, keep
`/data` on a volume, and put the relay behind a reverse proxy with TLS.
`GET /healthz` answers `204` while it runs.

## The stream for Android

`GET /v1/stream` is a Server-Sent Events stream for one phone. The phone
signs the request with an ECDSA P-256 key from the Android Keystore:

```
Authorization: Waylight-Device key=<public key>, ts=<unix seconds>, sig=<signature>
```

The key is the SubjectPublicKeyInfo in base64url, the signature (ASN.1,
base64url) covers `waylight-stream-v1\n<ts>`, and the time stamp must lie
within a minute of the relay's clock. The phone's device token, which it
gives its Waylight instance, is the base64url SHA-256 of the key.

Each message is one event: `id` grows, `data` is the sealed payload in
base64. A comment line every 25 seconds keeps the connection alive. When the
phone reconnects it sends `Last-Event-ID`; the relay deletes the messages up
to that ID and sends the rest. A phone has one stream; a new one replaces
the old.

## Limits

Per hour: 600 messages per instance key, 120 per device token and 1,200 per
client address; 600 stream openings per client address. At most 20 open
streams per client address and 20,000 in total. A request above a limit gets
`429`; instances and phones try again later. The counters live in memory and
start over after a restart.
