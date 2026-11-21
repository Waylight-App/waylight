package docsindex

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

var update = flag.Bool("update", false, "rewrite index.json from docs/")

// TestIndexUpToDate fails when a documentation page changed without the
// index; `make generate` rewrites it.
func TestIndexUpToDate(t *testing.T) {
	pages, err := Build(os.DirFS("../../../docs"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := encode(pages)
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.WriteFile("index.json", want, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if !bytes.Equal(raw, want) {
		t.Fatal("index.json is out of date: run make generate")
	}
}

// encode writes one page per line, so a changed page shows as one changed
// line in a diff while the file stays small.
func encode(pages []Page) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("[\n")
	for i, p := range pages {
		line, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}
		b.Write(line)
		if i < len(pages)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("]\n")
	return b.Bytes(), nil
}

func TestBuild(t *testing.T) {
	docs := fstest.MapFS{
		"index.md":          {Data: []byte("# Waylight\n\nA start page.\n")},
		"modules/README.md": {Data: []byte("# Modules\n")},
		"modules/immich.md": {Data: []byte(
			"# Immich\n\nShows your [library](x.md) and *memories*.\nSecond line.\n\nMore text.\n\n" +
				"```sh\n# not a heading\n```\n\n## Set up the API key\n\n### Über Ähnliches?\n")},
		"adr/0001-x.md": {Data: []byte("# Decision\n")},
	}
	pages, err := Build(docs)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3 (adr skipped): %+v", len(pages), pages)
	}
	if pages[0].Path != "" || pages[1].Path != "modules/" || pages[2].Path != "modules/immich/" {
		t.Errorf("paths = %q, %q, %q", pages[0].Path, pages[1].Path, pages[2].Path)
	}
	im := pages[2]
	if im.Summary != "Shows your library and memories. Second line." {
		t.Errorf("summary = %q", im.Summary)
	}
	if len(im.Sections) != 2 || im.Sections[0].Anchor != "set-up-the-api-key" || im.Sections[1].Anchor != "uber-ahnliches" {
		t.Errorf("sections = %+v", im.Sections)
	}
}

func TestShortenCutsAtAWord(t *testing.T) {
	s := shorten(strings.Repeat("word ", 100))
	if len(s) > maxSummary+4 || !strings.HasSuffix(s, " …") {
		t.Errorf("shorten = %q", s)
	}
}

func TestHandler(t *testing.T) {
	h := Handler("https://example.org/docs/")
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/api/help/docs", nil))
	var got Index
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://example.org/docs/" || len(got.Pages) == 0 {
		t.Errorf("index = %+v", got.BaseURL)
	}
}
