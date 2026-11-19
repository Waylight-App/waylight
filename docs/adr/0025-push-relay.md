# ADR 0025: Push relay for the mobile app

- Status: Proposed
- Date: 2026-11-19
- Issues: [#563](https://github.com/Waylight-App/waylight/issues/563), [#564](https://github.com/Waylight-App/waylight/issues/564)

## Context

The mobile app shows notifications from Waylight while it is closed. On
Android and iOS that only works through Google Firebase Cloud Messaging
(FCM) and the Apple Push Notification service (APNs). Both accept messages
only when they are signed with keys of the app's publisher, and those keys
cannot ship with self-hosted servers. Every Waylight instance therefore
hands its messages to one relay run by the project, which signs them for
FCM and APNs.

## Decision

1. **A separate program, `cmd/push-relay`.** It shares no state with a
   Waylight instance and has no access to one. It runs at
   `push.getwaylight.app`; FCM service account and APNs key exist only in
   its environment.

2. **One endpoint.** `POST /v1/push` with
   `{"platform": "fcm" | "apns", "token": "…", "payload": "<base64>"}`.
   Answers: `202` sent, `410` the device token is no longer valid (the
   instance removes the device), `401` bad signature, `403` token bound to
   another instance, `413` payload too large (more than 3,000 bytes after
   decoding), `429` rate limited, `502` FCM or APNs failed.

3. **Instances sign their requests.** Each instance creates an Ed25519 key
   pair on first use and keeps it. A request carries
   `Authorization: Waylight-Push key=<public key>, ts=<unix seconds>, sig=<signature>`
   (base64url without padding); the signature covers
   `waylight-push-v1\n<ts>\n<hex SHA-256 of the body>`. The relay accepts a
   time stamp within five minutes. There is no registration: the public key
   is the instance's identity.

4. **Device tokens are bound to the first instance.** The first accepted
   request for a device token binds the token's SHA-256 hash to the
   instance key; requests from other keys for that token get `403`. A
   leaked token cannot be used by anyone else to send to the device.
   Bindings that are unused for 90 days are deleted, and a `410` from FCM
   or APNs deletes the binding at once.

5. **The relay cannot read messages.** The Waylight instance encrypts each
   message for the device: ephemeral X25519 key agreement with the device's
   public key, HKDF-SHA256 (salt empty, info `waylight-push-v1`) to a
   32-byte key, AES-256-GCM with a random 12-byte nonce. The payload is
   `0x01 ‖ ephemeral public key (32) ‖ nonce (12) ‖ ciphertext and tag`;
   the plaintext is JSON `{"title", "body", "link"}`. FCM gets a data
   message `{"p": payload}` with high priority; APNs gets an alert with a
   generic text, `mutable-content: 1` and `p`, which the notification
   service extension decrypts and replaces before it is shown.

6. **Rate limits in memory.** Per instance key 600 messages an hour, per
   device token 120 an hour, per client address 1,200 an hour. Limits are
   not persisted; a restart resets them.

7. **What is stored and logged.** Only the bindings of point 4 (token hash,
   instance key, last use) in SQLite. Logs contain the status code and the
   duration, never tokens, keys or payloads.

## Consequences

- Google and Apple see when a device gets a message, not its content.
- The relay must be running for push messages; instances that turn it off
  (`PUSH_RELAY_URL` empty) still reach people through ntfy, Gotify, e-mail
  and the other channels.
- Rotating an instance key means the devices register again; the bindings
  of the old key expire after 90 days.
- The privacy policy of getwaylight.app gets a section on the relay before
  it goes live.
