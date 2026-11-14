// Package gateway reads day-ahead prices from aWATTar (#288).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/electricity/domain"
)

// DefaultAWATTarURLs are the public market data APIs of aWATTar (no API key;
// EPEX day-ahead prices of the bidding zone). "{tld}" is de or at.
const DefaultAWATTarURL = "https://api.awattar.{tld}/v1"

// AWATTar implements usecase.PriceProvider.
type AWATTar struct {
	client *http.Client
	base   string // with {tld}
}

// NewAWATTar returns the provider; base may contain "{tld}".
func NewAWATTar(client *http.Client, base string) *AWATTar {
	if base == "" {
		base = DefaultAWATTarURL
	}
	return &AWATTar{client: client, base: strings.TrimSuffix(base, "/")}
}

func (a *AWATTar) url(zone domain.Zone) string {
	tld := "de"
	if zone == domain.ZoneAT {
		tld = "at"
	}
	return strings.ReplaceAll(a.base, "{tld}", tld) + "/marketdata"
}

// Slots implements usecase.PriceProvider.
func (a *AWATTar) Slots(ctx context.Context, zone domain.Zone, from, to time.Time) ([]domain.Slot, error) {
	u := a.url(zone) + "?start=" + strconv.FormatInt(from.UnixMilli(), 10) + "&end=" + strconv.FormatInt(to.UnixMilli(), 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aWATTar answered %s", res.Status)
	}
	var body struct {
		Data []struct {
			Start       int64   `json:"start_timestamp"`
			End         int64   `json:"end_timestamp"`
			MarketPrice float64 `json:"marketprice"`
			Unit        string  `json:"unit"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("parse aWATTar answer: %w", err)
	}
	out := make([]domain.Slot, 0, len(body.Data))
	for _, d := range body.Data {
		if !strings.EqualFold(d.Unit, "Eur/MWh") && d.Unit != "" {
			continue
		}
		out = append(out, domain.Slot{Start: time.UnixMilli(d.Start).UTC(), End: time.UnixMilli(d.End).UTC(), EurPerMWh: d.MarketPrice})
	}
	return out, nil
}
