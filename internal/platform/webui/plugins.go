package webui

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// PluginsHandler answers GET /api/plugins with the frontend plugins of this
// build (ADR 0021): the app loads each listed module after start. A core
// build lists none, so the app never asks for a bundle that is not there.
func PluginsHandler(urls []string) http.HandlerFunc {
	list := append([]string{}, urls...)
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		_ = json.NewEncoder(w).Encode(struct {
			Plugins []string `json:"plugins"`
		}{Plugins: list})
	}
}

// PluginFiles serves a plugin bundle mounted at prefix (for example "/ee/").
// Unlike the app it has no fallback page: a missing file is a 404, and
// directories are never listed. The entry module keeps its name across
// releases, so it is revalidated; files under assets/ carry a content hash
// and are cached for good.
func PluginFiles(prefix string, fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.StripPrefix(strings.TrimSuffix(prefix, "/"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		info, err := fs.Stat(fsys, name)
		if name == "" || name == "." || err != nil || info.IsDir() || strings.HasPrefix(path.Base(name), ".") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		files.ServeHTTP(w, r)
	}))
}
