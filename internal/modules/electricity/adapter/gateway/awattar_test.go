package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trisse-home/internal/modules/electricity/domain"
)

// An answer as aWATTar sends it (shortened).
const recorded = `{"object":"list","data":[` +
	`{"start_timestamp":1794610800000,"end_timestamp":1794611700000,"marketprice":98.43,"unit":"Eur/MWh"},` +
	`{"start_timestamp":1794611700000,"end_timestamp":1794612600000,"marketprice":-3.1,"unit":"Eur/MWh"}],"url":"/at/v1/marketdata"}`

func TestAWATTar(t *testing.T) {
	var gotPath, gotStart string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotStart = r.URL.Path, r.URL.Query().Get("start")
		_, _ = io.WriteString(w, recorded)
	}))
	defer srv.Close()
	a := NewAWATTar(srv.Client(), srv.URL+"/{tld}/v1")
	from := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	slots, err := a.Slots(context.Background(), domain.ZoneAT, from, from.AddDate(0, 0, 2))
	if err != nil || len(slots) != 2 || slots[1].EurPerMWh != -3.1 || slots[0].End.Sub(slots[0].Start) != 15*time.Minute {
		t.Fatalf("%+v %v", slots, err)
	}
	if gotPath != "/at/v1/marketdata" || gotStart != "1794614400000" {
		t.Fatalf("request: %s start=%s", gotPath, gotStart)
	}
	if NewAWATTar(nil, "").url(domain.ZoneDELU) != "https://api.awattar.de/v1/marketdata" {
		t.Fatal("default URL")
	}
}
