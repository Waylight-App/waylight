package configfile

import (
	"errors"
	"testing"

	"trisse-home/internal/modules/bookmarks/usecase"
	"trisse-home/internal/shared/kernel"
)

func TestSectionRoundTrip(t *testing.T) {
	in := usecase.ConfigSection{
		Unfiled: []usecase.ConfigBookmark{{Title: "Docs", URL: "https://docs", Tags: []string{"work"}}},
		Folders: []usecase.ConfigFolder{{Name: "Media", Bookmarks: []usecase.ConfigBookmark{{URL: "https://films"}}}},
	}
	out, err := FromSection(in).Section()
	if err != nil || len(out.Unfiled) != 1 || out.Folders[0].Name != "Media" || out.Folders[0].Bookmarks[0].URL != "https://films" {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestEntryIsBookmarkOrFolder(t *testing.T) {
	for _, s := range []Section{
		{{Folder: "Media", URL: "https://x"}},
		{{Title: "No URL"}},
		{{Folder: "Media", Tags: []string{"x"}}},
	} {
		if _, err := s.Section(); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v accepted", s)
		}
	}
}
