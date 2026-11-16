// Package controller receives the Bot Connector's requests for the Teams
// command box search (#334). The route is reachable from the internet when
// an organisation switched it on: every request is checked by botauth
// before anything else happens.
//
// Ring: Interface adapters (controller).
package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"trisse-home/ee/teamsbot/adapter/botauth"
	"trisse-home/ee/teamsbot/domain"
)

// Verifier checks the Bot Connector token (port).
type Verifier interface {
	Verify(ctx context.Context, authorization string, a botauth.Activity) error
}

// Querier answers a search (port).
type Querier interface {
	Query(ctx context.Context, objectID, query string) domain.Response
}

// Handlers serve the route.
type Handlers struct {
	verify Verifier
	query  Querier
}

// NewHandlers returns the handlers.
func NewHandlers(v Verifier, q Querier) *Handlers { return &Handlers{verify: v, query: q} }

// maxBody bounds an activity; Teams sends a few kilobytes.
const maxBody = 256 << 10

// answerWithin leaves Teams' five seconds some room.
const answerWithin = 4 * time.Second

type activity struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	ServiceURL string `json:"serviceUrl"`
	ChannelID  string `json:"channelId"`
	From       struct {
		AADObjectID string `json:"aadObjectId"`
	} `json:"from"`
	Value struct {
		Parameters []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"parameters"`
	} `json:"value"`
}

// Messages handles POST /api/teams/messages.
func (h *Handlers) Messages(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}
	var a activity
	if err := json.Unmarshal(body, &a); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.verify.Verify(r.Context(), r.Header.Get("Authorization"), botauth.Activity{ServiceURL: a.ServiceURL, ChannelID: a.ChannelID}); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if a.Type != "invoke" || a.Name != "composeExtension/query" {
		// Messages, installs and other events need no answer.
		w.WriteHeader(http.StatusOK)
		return
	}
	query := ""
	for _, p := range a.Value.Parameters {
		if p.Name == "initialRun" {
			continue
		}
		query = p.Value
		break
	}
	ctx, cancel := context.WithTimeout(r.Context(), answerWithin)
	defer cancel()
	res := h.query.Query(ctx, a.From.AADObjectID, query)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}
