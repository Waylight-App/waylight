# ADR 0027: Push to Android over Server-Sent Events from the relay

- Status: Accepted
- Date: 2026-11-20
- Issues: [#582](https://github.com/Waylight-App/waylight/issues/582), [waylight-mobile#205](https://github.com/Waylight-App/waylight-mobile/issues/205)
- Supersedes: the FCM parts of [ADR 0025](0025-push-relay.md)

## Context

ADR 0025 delivers Android notifications through Google Firebase Cloud
Messaging. That puts Google code into the app and Google into every
delivery. The project runs its own server at Hetzner
(`waylight-eu-central-1`), which already hosts the relay. Android has no
system push service besides FCM; without it, an app receives in real time
only through a connection it keeps open itself.

## Decision

1. **The relay delivers to Android itself.** The app keeps one
   Server-Sent Events stream open to `GET /v1/stream` on
   `push.getwaylight.app`. FCM, its sender and its configuration are
   removed. APNs stays for the iOS app.
2. **Server-Sent Events, not WebSocket or MQTT.** Notifications flow from
   the relay to the phone only. SSE is plain HTTP through the existing
   reverse proxy, reconnects by itself and resumes with `Last-Event-ID`; it
   needs no second protocol and no broker.
3. **Devices prove their key.** The device signs the stream request with an
   ECDSA P-256 key from the Android Keystore
   (`Authorization: Waylight-Device key=…, ts=…, sig=…`, signature over
   `waylight-stream-v1\n<ts>`, one minute of allowed clock skew). The
   device token the instance sends to is the base64url SHA-256 of that key,
   so only the holder of the key can open the stream for it. Instances keep
   signing their requests as in ADR 0025, and the first instance stays bound
   to the token.
4. **Mailbox per device.** The relay stores each sealed message in SQLite
   with a growing event ID until the device acknowledges it through
   `Last-Event-ID` on its next connection; at most 100 messages per device,
   kept for 72 hours. A device that was offline catches up when it
   reconnects.
5. **One stream per device, bounded resources.** A new stream replaces the
   older one of the same device. A comment line every 25 seconds keeps the
   connection alive; every write has its own deadline. Streams are limited
   per client address (20 open, 600 openings an hour) and in total
   (20,000).
6. **Content stays end-to-end encrypted.** Payloads are sealed for the
   device's X25519 key as in ADR 0025. The relay never learns that key, so
   it can neither read a message nor make up one the device would accept.
7. **Platform `sse`.** Instances register Android devices with the
   platform `sse` instead of `fcm`; devices registered for FCM are deleted
   and register again.

## Consequences

- No Google service takes part in delivering to Android; the app needs no
  Google code.
- Android requires a foreground service with a permanent notification to
  keep the stream open; the app explains it and lets the user turn push
  off.
- The relay holds open connections and stores sealed messages for a short
  time; it learns the network address and online times of phones, which it
  does not log.
- The web interface does not use Web Push, since Chrome and Edge deliver it
  through Google; it gets live notifications from its own instance over
  Server-Sent Events (#583).

## Alternatives considered

- **FCM (ADR 0025):** no connection of our own, but Google code in the app
  and Google in every delivery.
- **WebSocket:** two-way, which push does not need; more protocol to secure.
- **MQTT:** efficient, but a broker as another service to run and secure.
- **UnifiedPush:** open and battery friendly, but users must install a
  distributor app; it can be added later for those who have one.
- **Polling with WorkManager:** no permanent notification, but delays of 15
  minutes and more.
