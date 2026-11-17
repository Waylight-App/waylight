package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/notifications/domain"
)

func TestSlackMessage(t *testing.T) {
	m := domain.Message{Title: "NAS is down", Body: "<!channel> & friends", Severity: domain.SeverityCritical, Type: "service"}
	var p struct {
		Text   string `json:"text"`
		Blocks []struct {
			Type     string                  `json:"type"`
			Text     struct{ Text string }   `json:"text"`
			Elements []struct{ Text string } `json:"elements"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(SlackMessage(m, "https://home.example.org"), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Blocks) != 3 || p.Blocks[0].Text.Text != "NAS is down" {
		t.Fatalf("%+v", p)
	}
	// A mention in a service name must not ping the channel.
	if p.Blocks[1].Text.Text != "&lt;!channel&gt; &amp; friends" || strings.Contains(p.Text, "<!channel>") {
		t.Fatalf("not escaped: %+v", p)
	}
	if f := p.Blocks[2].Elements[0].Text; !strings.Contains(f, "Critical") || !strings.Contains(f, "<https://home.example.org|waylight>") {
		t.Fatalf("footer %q", f)
	}
	_ = json.Unmarshal(SlackMessage(domain.Message{Title: strings.Repeat("x", 300)}, ""), &p)
	if n := len([]rune(p.Blocks[0].Text.Text)); n != slackHeaderLimit {
		t.Fatalf("the header must be cut to Slack's limit, got %d runes", n)
	}
}

func TestTeamsMessage(t *testing.T) {
	var p struct {
		Type        string `json:"type"`
		Attachments []struct {
			ContentType string          `json:"contentType"`
			ContentURL  *string         `json:"contentUrl"`
			Content     json.RawMessage `json:"content"`
		} `json:"attachments"`
	}
	raw := TeamsMessage(domain.Message{Title: "Backup failed", Body: "nightly", Severity: domain.SeverityWarning, Type: "backup"}, "https://home.example.org")
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != "message" || len(p.Attachments) != 1 || p.Attachments[0].ContentType != "application/vnd.microsoft.card.adaptive" || p.Attachments[0].ContentURL != nil {
		t.Fatalf("envelope: %s", raw)
	}
	card := string(p.Attachments[0].Content)
	for _, want := range []string{`"type":"AdaptiveCard"`, `"version":"1.4"`, `"color":"Warning"`, `"Action.OpenUrl"`, `"url":"https://home.example.org"`, "nightly"} {
		if !strings.Contains(card, want) {
			t.Errorf("card lacks %s: %s", want, card)
		}
	}
	if strings.Contains(string(TeamsMessage(domain.Message{Title: "x"}, "")), "Action.OpenUrl") {
		t.Fatal("no link without PUBLIC_URL")
	}
}

func TestChatPublishers(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = append(got, r.URL.Path+" "+r.Header.Get("Content-Type")+" "+string(b[:min(len(b), 20)]))
		if strings.Contains(r.URL.Path, "broken") {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	m := domain.Message{Title: "t", Severity: domain.SeverityInfo}
	if err := NewSlackPublisher(srv.Client(), "").Publish(context.Background(), domain.Channel{Secret: srv.URL + "/services/T/B/x"}, m); err != nil {
		t.Fatal(err)
	}
	if err := NewTeamsPublisher(srv.Client(), "").Publish(context.Background(), domain.Channel{Secret: srv.URL + "/workflows/1"}, m); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !strings.Contains(got[0], "application/json") {
		t.Fatalf("%v", got)
	}
	if err := NewTeamsPublisher(srv.Client(), "").Publish(context.Background(), domain.Channel{Secret: srv.URL + "/broken"}, m); err == nil {
		t.Fatal("a refused post must be an error")
	}
	// Transport errors carry the URL; it is the credential and must not leak.
	err := NewSlackPublisher(srv.Client(), "").Publish(context.Background(), domain.Channel{Secret: "http://127.0.0.1:1/services/SECRET"}, m)
	if err == nil || strings.Contains(err.Error(), "SECRET") {
		t.Fatalf("%v", err)
	}
}
