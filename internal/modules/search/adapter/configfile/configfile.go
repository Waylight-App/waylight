// Package configfile maps the search settings to their section of
// waylight.yaml (#88) and back.
//
// Ring: Interface Adapters.
package configfile

import (
	"waylight/internal/modules/search/domain"
	"waylight/internal/modules/search/usecase"
)

// Section is the search section.
type Section struct {
	Provider        string     `yaml:"provider,omitempty"`
	OpenInNewTab    bool       `yaml:"openInNewTab,omitempty"`
	CapturePrefixes *Prefixes  `yaml:"capturePrefixes,omitempty"`
	Shortcuts       []Shortcut `yaml:"shortcuts,omitempty"`
	Keywords        []Keyword  `yaml:"keywords,omitempty"`
}

// Prefixes are the quick capture prefixes of the search field.
type Prefixes struct {
	Todo     string `yaml:"todo"`
	Note     string `yaml:"note"`
	Bookmark string `yaml:"bookmark"`
}

// Shortcut is a !bang; {query} in the URL is replaced by the search terms.
type Shortcut struct {
	Keyword string `yaml:"keyword"`
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
}

// Keyword is an own keyword (#478): kind is service, bookmark, widget, url
// or query; target is the id, the web address or the search text.
type Keyword struct {
	Keyword string `yaml:"keyword"`
	Kind    string `yaml:"kind"`
	Target  string `yaml:"target"`
	Label   string `yaml:"label,omitempty"`
}

// FromSection maps the use case section.
func FromSection(s usecase.ConfigSection) *Section {
	out := &Section{Provider: s.Provider, OpenInNewTab: s.OpenInNewTab}
	if s.Prefixes != domain.DefaultCapturePrefixes {
		out.CapturePrefixes = &Prefixes{Todo: s.Prefixes.Todo, Note: s.Prefixes.Note, Bookmark: s.Prefixes.Bookmark}
	}
	for _, sc := range s.Shortcuts {
		out.Shortcuts = append(out.Shortcuts, Shortcut{Keyword: sc.Keyword, Name: sc.Name, URL: sc.URLTemplate})
	}
	for _, k := range s.Keywords {
		out.Keywords = append(out.Keywords, Keyword{Keyword: k.Keyword, Kind: k.Kind, Target: k.Target, Label: k.Label})
	}
	return out
}

// Section maps back to the use case section.
func (s Section) Section() usecase.ConfigSection {
	out := usecase.ConfigSection{Provider: s.Provider, OpenInNewTab: s.OpenInNewTab}
	if p := s.CapturePrefixes; p != nil {
		out.Prefixes = domain.CapturePrefixes{Todo: p.Todo, Note: p.Note, Bookmark: p.Bookmark}
	}
	for _, sc := range s.Shortcuts {
		out.Shortcuts = append(out.Shortcuts, usecase.ShortcutView{Keyword: sc.Keyword, Name: sc.Name, URLTemplate: sc.URL})
	}
	for _, k := range s.Keywords {
		out.Keywords = append(out.Keywords, usecase.KeywordView{Keyword: k.Keyword, Kind: k.Kind, Target: k.Target, Label: k.Label})
	}
	return out
}
