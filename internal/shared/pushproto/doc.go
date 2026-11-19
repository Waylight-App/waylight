// Package pushproto is the wire protocol between a Waylight instance and the
// push relay (ADR 0025): request signatures with the instance's Ed25519 key
// and the end-to-end encryption of a message for a device. Both sides import
// it, so the format is defined once.
package pushproto
