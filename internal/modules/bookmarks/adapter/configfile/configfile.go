// Package configfile maps the bookmarks to their section of
// waylight.yaml (#88) and back.
//
// Ring: Interface Adapters.
package configfile

import (
	"waylight/internal/modules/bookmarks/usecase"
	"waylight/internal/shared/kernel"
)

// Bookmark is one link.
type Bookmark struct {
	Title       string   `yaml:"title,omitempty"`
	URL         string   `yaml:"url"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty,flow"`
}

// Entry is either a bookmark without folder or a folder with bookmarks,
// so the file reads like the bookmark list in the app.
type Entry struct {
	Folder      string     `yaml:"folder,omitempty"`
	Bookmarks   []Bookmark `yaml:"bookmarks,omitempty"`
	Title       string     `yaml:"title,omitempty"`
	URL         string     `yaml:"url,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Tags        []string   `yaml:"tags,omitempty,flow"`
}

// Section is the bookmarks section.
type Section []Entry

var errEntry = kernel.NewValidationError("bookmarks", "Each bookmarks entry is either a bookmark with url or a folder with bookmarks, not both.")

// FromSection maps the use case section: unfiled bookmarks first, then the folders.
func FromSection(s usecase.ConfigSection) *Section {
	out := Section{}
	for _, b := range s.Unfiled {
		out = append(out, Entry{Title: b.Title, URL: b.URL, Description: b.Description, Tags: b.Tags})
	}
	for _, f := range s.Folders {
		e := Entry{Folder: f.Name}
		for _, b := range f.Bookmarks {
			e.Bookmarks = append(e.Bookmarks, Bookmark(b))
		}
		out = append(out, e)
	}
	return &out
}

// Section maps back to the use case section.
func (s Section) Section() (usecase.ConfigSection, error) {
	var out usecase.ConfigSection
	for _, e := range s {
		isFolder := e.Folder != "" || len(e.Bookmarks) > 0
		if isFolder == (e.URL != "") || (isFolder && (e.Title != "" || e.Description != "" || len(e.Tags) > 0)) {
			return usecase.ConfigSection{}, errEntry
		}
		if !isFolder {
			out.Unfiled = append(out.Unfiled, usecase.ConfigBookmark{Title: e.Title, URL: e.URL, Description: e.Description, Tags: e.Tags})
			continue
		}
		f := usecase.ConfigFolder{Name: e.Folder}
		for _, b := range e.Bookmarks {
			f.Bookmarks = append(f.Bookmarks, usecase.ConfigBookmark(b))
		}
		out.Folders = append(out.Folders, f)
	}
	return out, nil
}
