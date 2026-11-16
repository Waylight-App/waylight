// Package gateway reads daily closes from the ECB and Stooq (#284).
//
// Ring: Interface Adapters.
package gateway

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/markets/domain"
)

// Default addresses (no API key).
const (
	DefaultECBURL   = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-hist-90d.xml"
	DefaultStooqURL = "https://stooq.com/q/d/l/"
)

func fetch(ctx context.Context, client *http.Client, u string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("source answered %s", res.Status)
	}
	return res.Body, nil
}

// ECB reads the euro reference rates of the last 90 days. One file holds
// every currency, so it is shared by all FX symbols of a request.
type ECB struct {
	client *http.Client
	url    string
}

// NewECB returns the ECB provider.
func NewECB(client *http.Client, u string) *ECB {
	if u == "" {
		u = DefaultECBURL
	}
	return &ECB{client: client, url: u}
}

// Series implements usecase.SeriesProvider: euro in the symbol's currency.
func (e *ECB) Series(ctx context.Context, sym domain.Symbol, from, to time.Time) ([]domain.Point, error) {
	body, err := fetch(ctx, e.client, e.url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var doc struct {
		Days []struct {
			Time  string `xml:"time,attr"`
			Rates []struct {
				Currency string  `xml:"currency,attr"`
				Rate     float64 `xml:"rate,attr"`
			} `xml:"Cube"`
		} `xml:"Cube>Cube"`
	}
	if err := xml.NewDecoder(io.LimitReader(body, 2<<20)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse ECB rates: %w", err)
	}
	var out []domain.Point
	for _, d := range doc.Days {
		day, err := time.Parse(time.DateOnly, d.Time)
		if err != nil || day.Before(from) || day.After(to) {
			continue
		}
		for _, r := range d.Rates {
			if r.Currency == sym.Code {
				out = append(out, domain.Point{Day: day, Value: r.Rate})
			}
		}
	}
	// The file is newest first.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// Latest implements usecase.RateSource: the newest day of euro reference
// rates in the file, for converting currencies in the search (#482).
func (e *ECB) Latest(ctx context.Context) (time.Time, map[string]float64, error) {
	body, err := fetch(ctx, e.client, e.url)
	if err != nil {
		return time.Time{}, nil, err
	}
	defer body.Close()
	var doc struct {
		Days []struct {
			Time  string `xml:"time,attr"`
			Rates []struct {
				Currency string  `xml:"currency,attr"`
				Rate     float64 `xml:"rate,attr"`
			} `xml:"Cube"`
		} `xml:"Cube>Cube"`
	}
	if err := xml.NewDecoder(io.LimitReader(body, 2<<20)).Decode(&doc); err != nil {
		return time.Time{}, nil, fmt.Errorf("parse ECB rates: %w", err)
	}
	var newest time.Time
	var rates map[string]float64
	for _, d := range doc.Days {
		day, err := time.Parse(time.DateOnly, d.Time)
		if err != nil || !day.After(newest) || len(d.Rates) == 0 {
			continue
		}
		newest, rates = day, map[string]float64{}
		for _, r := range d.Rates {
			if r.Rate > 0 && len(r.Currency) == 3 {
				rates[r.Currency] = r.Rate
			}
		}
	}
	if rates == nil {
		return time.Time{}, nil, fmt.Errorf("the ECB file holds no rates")
	}
	return newest, rates, nil
}

// Stooq reads daily closes as CSV (Date,Open,High,Low,Close,Volume).
type Stooq struct {
	client *http.Client
	url    string
}

// NewStooq returns the Stooq provider.
func NewStooq(client *http.Client, u string) *Stooq {
	if u == "" {
		u = DefaultStooqURL
	}
	return &Stooq{client: client, url: u}
}

// Series implements usecase.SeriesProvider.
func (s *Stooq) Series(ctx context.Context, sym domain.Symbol, from, to time.Time) ([]domain.Point, error) {
	q := url.Values{"s": {sym.Code}, "i": {"d"}, "d1": {from.Format("20060102")}, "d2": {to.Format("20060102")}}
	body, err := fetch(ctx, s.client, s.url+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var out []domain.Point
	sc := bufio.NewScanner(io.LimitReader(body, 1<<20))
	for first := true; sc.Scan(); first = false {
		line := strings.TrimSpace(sc.Text())
		if first {
			if !strings.HasPrefix(line, "Date,") {
				// "No data" or a limit notice instead of the CSV header.
				return nil, nil
			}
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 5 {
			continue
		}
		day, err1 := time.Parse(time.DateOnly, f[0])
		v, err2 := strconv.ParseFloat(f[4], 64)
		if err1 == nil && err2 == nil {
			out = append(out, domain.Point{Day: day, Value: v})
		}
	}
	return out, sc.Err()
}
