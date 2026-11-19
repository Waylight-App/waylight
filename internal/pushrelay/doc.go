// Package pushrelay forwards end-to-end encrypted messages from Waylight
// instances to Google FCM and Apple APNs (ADR 0025). The use case in this
// package is free of HTTP and storage; the senders, the binding store and
// the HTTP handler live in the sub-packages and are wired in cmd/push-relay.
package pushrelay
