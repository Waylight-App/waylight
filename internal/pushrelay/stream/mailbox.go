package stream

import (
	"context"
	"time"
)

// Event is one sealed message waiting for a device. IDs grow with every
// message, so a device acknowledges everything up to the last ID it saw.
type Event struct {
	ID      int64
	Payload []byte
}

// Mailbox stores sealed messages per device until they are acknowledged or
// expire (output port, implemented by the relay's store).
type Mailbox interface {
	// Put stores a message and keeps at most limit messages for the device,
	// dropping the oldest.
	Put(ctx context.Context, device string, payload []byte, now time.Time, limit int) (int64, error)
	// After returns the messages of device with an ID above after, oldest first.
	After(ctx context.Context, device string, after int64) ([]Event, error)
	// Ack deletes the messages of device up to and including id.
	Ack(ctx context.Context, device string, id int64) error
}
