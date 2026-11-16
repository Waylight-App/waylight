// Package socket keeps a Slack Socket Mode connection (#333, ADR 0023): the
// instance connects out to Slack, so Slack never needs to reach it.
//
// Ring: Interface adapters (controller: it receives commands and hands them
// to the use case).
package socket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/websocket"
)

// Opener returns a fresh Socket Mode address (apps.connections.open).
type Opener func(ctx context.Context) (string, error)

// Command is a slash command from the envelope's payload.
type Command struct {
	TeamID  string `json:"team_id"`
	UserID  string `json:"user_id"`
	Command string `json:"command"`
	Text    string `json:"text"`
}

// Handler answers a command with the text only the asking person sees.
type Handler func(ctx context.Context, c Command) string

type envelope struct {
	Type       string          `json:"type"`
	EnvelopeID string          `json:"envelope_id"`
	Reason     string          `json:"reason"`
	Payload    json.RawMessage `json:"payload"`
}

type ack struct {
	EnvelopeID string `json:"envelope_id"`
	Payload    any    `json:"payload,omitempty"`
}

type message struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// AnswerWithin is the time a command gets before the acknowledgement must
// leave; Slack waits three seconds.
const AnswerWithin = 2500 * time.Millisecond

// Run connects and serves until ctx ends, reconnecting with a growing pause
// when the connection drops or Slack asks for a refresh.
func Run(ctx context.Context, open Opener, handle Handler, onConnect func(context.Context)) error {
	pause := time.Second
	for ctx.Err() == nil {
		err := serve(ctx, open, handle, onConnect)
		if ctx.Err() != nil {
			return nil
		}
		if errors.Is(err, errRefresh) {
			pause = time.Second
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(pause):
		}
		pause = min(pause*2, 2*time.Minute)
	}
	return nil
}

// errRefresh is a disconnect Slack announced; reconnect at once.
var errRefresh = errors.New("slack asked to reconnect")

func serve(ctx context.Context, open Opener, handle Handler, onConnect func(context.Context)) error {
	addr, err := open(ctx)
	if err != nil {
		return err
	}
	cfg, err := websocket.NewConfig(addr, "https://slack.com")
	if err != nil {
		return err
	}
	conn, err := cfg.DialContext(ctx)
	if err != nil {
		return fmt.Errorf("connect to slack: %w", err)
	}
	defer conn.Close()
	// Closing the connection ends the blocking read when ctx ends.
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if onConnect != nil {
		go onConnect(ctx)
	}
	for {
		var env envelope
		if err := websocket.JSON.Receive(conn, &env); err != nil {
			return fmt.Errorf("read from slack: %w", err)
		}
		switch env.Type {
		case "hello":
		case "disconnect":
			return errRefresh
		case "slash_commands":
			var c Command
			if err := json.Unmarshal(env.Payload, &c); err != nil {
				_ = websocket.JSON.Send(conn, ack{EnvelopeID: env.EnvelopeID})
				continue
			}
			cctx, cancel := context.WithTimeout(ctx, AnswerWithin)
			text := handle(cctx, c)
			cancel()
			if err := websocket.JSON.Send(conn, ack{EnvelopeID: env.EnvelopeID,
				Payload: message{ResponseType: "ephemeral", Text: text}}); err != nil {
				return fmt.Errorf("answer slack: %w", err)
			}
		default:
			// Events and interactions are not used; acknowledge so Slack does
			// not retry them.
			if env.EnvelopeID != "" {
				_ = websocket.JSON.Send(conn, ack{EnvelopeID: env.EnvelopeID})
			}
		}
	}
}
