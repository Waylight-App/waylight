// Package stream delivers sealed push messages to devices that keep a
// Server-Sent Events stream open to the relay (ADR 0027). Messages wait in
// a mailbox per device until the device acknowledges them, so a device
// that was offline catches up when it reconnects.
package stream
