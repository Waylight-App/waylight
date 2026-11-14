package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/markets/domain"
)

// Shortened answers as the ECB and Stooq send them.
const ecbXML = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
<gesmes:subject>Reference rates</gesmes:subject>
<Cube><Cube time="2026-11-13"><Cube currency="USD" rate="1.0854"/><Cube currency="CHF" rate="0.9312"/></Cube>
<Cube time="2026-11-12"><Cube currency="USD" rate="1.08"/><Cube currency="CHF" rate="0.93"/></Cube></Cube></gesmes:Envelope>`

const stooqCSV = "Date,Open,High,Low,Close,Volume\n2026-11-12,180,182,179,181.5,1000\n2026-11-13,181.5,185,181,184.25,1200\n"

func TestSources(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ecb.xml":
			_, _ = io.WriteString(w, ecbXML)
		case "/q/d/l/":
			query = r.URL.RawQuery
			if r.URL.Query().Get("s") == "nope.us" {
				_, _ = io.WriteString(w, "No data")
				return
			}
			_, _ = io.WriteString(w, stooqCSV)
		}
	}))
	defer srv.Close()
	ctx := context.Background()
	from, to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	fx, err := NewECB(srv.Client(), srv.URL+"/ecb.xml").Series(ctx, domain.Symbol{Source: domain.SourceFX, Code: "USD"}, from, to)
	if err != nil || len(fx) != 2 || fx[0].Value != 1.08 || fx[1].Value != 1.0854 {
		t.Fatalf("oldest first: %+v %v", fx, err)
	}
	st := NewStooq(srv.Client(), srv.URL+"/q/d/l/")
	pts, err := st.Series(ctx, domain.Symbol{Source: domain.SourceStooq, Code: "aapl.us"}, from, to)
	if err != nil || len(pts) != 2 || pts[1].Value != 184.25 {
		t.Fatalf("%+v %v", pts, err)
	}
	if query != "d1=20261001&d2=20261114&i=d&s=aapl.us" {
		t.Fatal(query)
	}
	if pts, err := st.Series(ctx, domain.Symbol{Source: domain.SourceStooq, Code: "nope.us"}, from, to); err != nil || pts != nil {
		t.Fatal("no data must be empty, not an error", pts, err)
	}
}
