package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/containers/json" || r.URL.Query().Get("all") != "1" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"Id":"0123456789abcdef","Names":["/homarr"],"Image":"ghcr.io/x:latest","State":"running",
			"Status":"Up 3 hours (healthy)","Created":1700000000,"Labels":{"com.docker.compose.project":"homarr"},
			"Ports":[{"IP":"0.0.0.0","PrivatePort":7575,"PublicPort":7575,"Type":"tcp"},{"IP":"::","PrivatePort":7575,"PublicPort":7575,"Type":"tcp"},
				{"PrivatePort":53,"Type":"udp"}],
			"NetworkSettings":{"Networks":{"proxy":{"IPAddress":"172.18.0.5"},"none":{"IPAddress":""}}}}]`))
	}))
	defer srv.Close()
	got, err := NewAPI(srv.Client(), srv.URL+"/").Containers(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected %v %v", got, err)
	}
	c := got[0]
	if c.ID != "0123456789ab" || c.Name != "homarr" || c.Project != "homarr" || c.Created.Year() != 2023 {
		t.Fatalf("mapping wrong: %+v", c)
	}
	// TCP ports once each and the network addresses, for discovery (#590).
	if len(c.Ports) != 1 || c.Ports[0].Private != 7575 || c.Ports[0].Public != 7575 || len(c.Addresses) != 1 || c.Addresses[0] != "172.18.0.5" {
		t.Fatalf("ports and addresses: %+v %+v", c.Ports, c.Addresses)
	}
	if _, err := NewAPI(srv.Client(), srv.URL+"/nope").Containers(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
