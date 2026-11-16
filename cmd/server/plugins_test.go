package main

import (
	"io"
	"net/http"
	"testing"
	"testing/fstest"
)

// TestFrontendPlugins checks ADR 0021's server side: /api/plugins is public
// and lists the build's plugins, whose files are served under their prefix.
func TestFrontendPlugins(t *testing.T) {
	saved := frontendPlugins
	t.Cleanup(func() { frontendPlugins = saved })

	frontendPlugins = nil
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	var list struct {
		Plugins []string `json:"plugins"`
	}
	newClient(t, srv.URL).json(http.MethodGet, "/api/plugins", nil, http.StatusOK, &list)
	if len(list.Plugins) != 0 {
		t.Fatalf("a build without plugins lists %v", list.Plugins)
	}

	frontendPlugins = []frontendPlugin{{prefix: "/ee/", entry: "/ee/plugin.js",
		files: fstest.MapFS{"plugin.js": {Data: []byte("export default {}")}}}}
	srv = newTestServer(t, up.URL)
	c := newClient(t, srv.URL)
	c.json(http.MethodGet, "/api/plugins", nil, http.StatusOK, &list)
	if len(list.Plugins) != 1 || list.Plugins[0] != "/ee/plugin.js" {
		t.Fatalf("plugins %v", list.Plugins)
	}
	res, err := http.Get(srv.URL + "/ee/plugin.js")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || string(body) != "export default {}" {
		t.Fatalf("entry: %d %q", res.StatusCode, body)
	}
	res, err = http.Get(srv.URL + "/ee/nope.js")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing file: %d (no app fallback under a plugin prefix)", res.StatusCode)
	}
}
