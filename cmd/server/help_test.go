package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"testing"

	"waylight/internal/modules"
)

// TestWidgetHelpCoversEveryKind: every widget kind has exactly one help
// entry (#678), so a new widget cannot ship without one; the ee build runs
// it with its own modules too.
func TestWidgetHelpCoversEveryKind(t *testing.T) {
	up := fakeUpstream(t)
	a, err := build(context.Background(), newTestConfig(t, up.URL), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.db.Close()
	helped := map[string]int{}
	for _, h := range frontendOnlyHelp {
		helped[h.Kind]++
	}
	kinds := slices.Clone(frontendOnlyWidgetKinds)
	for _, m := range a.registry.All() {
		kinds = append(kinds, m.WidgetKinds()...)
		for _, h := range modules.HelpOf(m) {
			helped[h.Kind]++
			if !slices.Contains(m.WidgetKinds(), h.Kind) {
				t.Errorf("%s: help for %q, a kind it does not back", m.Name(), h.Kind)
			}
			if modules.HelpGerman(h.Shows) == "" || modules.HelpGerman(h.SetUp) == "" {
				t.Errorf("%s: no German help text", h.Kind)
			}
			if h.Shows.Template == "" || h.SetUp.Template == "" || h.Docs == "" || strings.HasPrefix(h.Docs, "/") {
				t.Errorf("%s: incomplete help %+v", h.Kind, h)
			}
		}
	}
	for _, h := range frontendOnlyHelp {
		if modules.HelpGerman(h.Shows) == "" || modules.HelpGerman(h.SetUp) == "" {
			t.Errorf("%s: no German help text", h.Kind)
		}
	}
	for _, k := range kinds {
		if helped[k] != 1 {
			t.Errorf("widget kind %q has %d help entries, want 1", k, helped[k])
		}
	}
}

// TestHelpGerman: ?lang=de and Accept-Language serve German texts under the
// same keys, English stays the default.
func TestHelpGerman(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)

	immich := func(path string) modules.WidgetHelpInfo {
		var help modules.HelpResponse
		c.json(http.MethodGet, path, nil, http.StatusOK, &help)
		i := slices.IndexFunc(help.Widgets, func(w modules.WidgetHelpInfo) bool { return w.Kind == "immich" })
		if i < 0 {
			t.Fatalf("no help for immich at %s", path)
		}
		return help.Widgets[i]
	}
	en := immich("/api/help/widgets")
	de := immich("/api/help/widgets?lang=de")
	if de.Shows.Key != en.Shows.Key || de.SetUp.Key != en.SetUp.Key {
		t.Errorf("keys differ: %+v vs %+v", de, en)
	}
	if de.Shows.Text == en.Shows.Text || de.SetUp.Text == en.SetUp.Text {
		t.Errorf("German help is still English: %+v", de)
	}
	if fr := immich("/api/help/widgets?lang=fr"); fr.Shows.Text != en.Shows.Text {
		t.Errorf("unsupported language did not fall back to English: %+v", fr)
	}
}

// TestHelpEndpoints: the widget help and the documentation index are served
// to a signed-in account.
func TestHelpEndpoints(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)

	var help modules.HelpResponse
	c.json(http.MethodGet, "/api/help/widgets", nil, http.StatusOK, &help)
	i := slices.IndexFunc(help.Widgets, func(w modules.WidgetHelpInfo) bool { return w.Kind == "immich" })
	if i < 0 {
		t.Fatalf("no help for immich in %d entries", len(help.Widgets))
	}
	w := help.Widgets[i]
	if w.Shows.Key == "" || w.SetUp.Text == "" || w.DocsURL != modules.DocsBaseURL+"modules/immich/" {
		t.Errorf("immich help = %+v", w)
	}

	var docs struct {
		BaseURL string `json:"baseUrl"`
		Pages   []struct {
			Path, Title string
		} `json:"pages"`
	}
	c.json(http.MethodGet, "/api/help/docs", nil, http.StatusOK, &docs)
	if docs.BaseURL != modules.DocsBaseURL || !slices.ContainsFunc(docs.Pages, func(p struct{ Path, Title string }) bool {
		return p.Path == "modules/immich/"
	}) {
		t.Errorf("docs index: base %q, %d pages", docs.BaseURL, len(docs.Pages))
	}
}
