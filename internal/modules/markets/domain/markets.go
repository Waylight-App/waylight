// Package domain holds the market watchlist rules (#284).
//
// Ring: Entities.
package domain

import (
	"regexp"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Source of a symbol.
type Source string

// Sources without an API key.
const (
	// SourceFX are the ECB reference rates: euro against a currency.
	SourceFX Source = "fx"
	// SourceStooq are shares, indices and crypto pairs in Stooq's notation
	// (aapl.us, ^dax, btcusd).
	SourceStooq Source = "stooq"
)

// MaxSymbols limits a watchlist.
const MaxSymbols = 12

// Symbol is one entry of a watchlist, e.g. "fx:USD" or "stooq:aapl.us".
type Symbol struct {
	Source Source
	Code   string
}

func (s Symbol) String() string { return string(s.Source) + ":" + s.Code }

var (
	currencyCode = regexp.MustCompile(`^[A-Z]{3}$`)
	stooqCode    = regexp.MustCompile(`^\^?[a-z0-9._-]{1,20}$`)
	// ErrInvalidSymbol rejects entries of an unknown form.
	ErrInvalidSymbol = kernel.NewValidationError("symbols", "Use fx:USD for currencies or stooq:aapl.us for shares, indices and crypto.")
	// ErrTooMany limits the list.
	ErrTooMany = kernel.NewValidationError("symbols", "At most 12 entries per widget.")
)

// ParseSymbol reads "fx:USD" or "stooq:aapl.us".
func ParseSymbol(s string) (Symbol, error) {
	src, code, ok := strings.Cut(strings.TrimSpace(s), ":")
	switch Source(strings.ToLower(src)) {
	case SourceFX:
		code = strings.ToUpper(code)
		if ok && currencyCode.MatchString(code) && code != "EUR" {
			return Symbol{Source: SourceFX, Code: code}, nil
		}
	case SourceStooq:
		code = strings.ToLower(code)
		if ok && stooqCode.MatchString(code) {
			return Symbol{Source: SourceStooq, Code: code}, nil
		}
	}
	return Symbol{}, ErrInvalidSymbol
}

// ParseSymbols reads a watchlist; duplicates are dropped.
func ParseSymbols(list []string) ([]Symbol, error) {
	var out []Symbol
	seen := map[Symbol]bool{}
	for _, s := range list {
		sym, err := ParseSymbol(s)
		if err != nil {
			return nil, err
		}
		if !seen[sym] {
			seen[sym] = true
			out = append(out, sym)
		}
	}
	if len(out) > MaxSymbols {
		return nil, ErrTooMany
	}
	return out, nil
}

// Point is one daily close.
type Point struct {
	Day   time.Time
	Value float64
}

// Quote is the latest value of a symbol with its change.
type Quote struct {
	Symbol Symbol
	Value  float64
	// ChangePct is the change against the previous close, in percent.
	ChangePct float64
	AsOf      time.Time
	// History are the closes of the last weeks, oldest first.
	History []Point
}

// NewQuote builds the quote from a series of closes (oldest first); ok is
// false for an empty series.
func NewQuote(sym Symbol, series []Point) (Quote, bool) {
	if len(series) == 0 {
		return Quote{}, false
	}
	last := series[len(series)-1]
	q := Quote{Symbol: sym, Value: last.Value, AsOf: last.Day, History: series}
	if len(series) > 1 && series[len(series)-2].Value != 0 {
		prev := series[len(series)-2].Value
		q.ChangePct = (last.Value - prev) / prev * 100
	}
	return q, true
}
