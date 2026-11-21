package docsindex

import (
	"bufio"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// maxSummary bounds the summary of a page; the search only needs enough
// words to match a question.
const maxSummary = 240

// skipped are folders of the documentation that answer no question a user
// asks inside the app: decision records and JSON schemas.
var skipped = []string{"adr/", "schema/", "screenshots/"}

var (
	mdLink   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	mdMarks  = regexp.MustCompile("[*_`]")
	slugDrop = regexp.MustCompile(`[^\w\s-]`)
	slugJoin = regexp.MustCompile(`[-\s]+`)
)

// Build reads every Markdown page of docs and returns the index in path order.
func Build(docs fs.FS) ([]Page, error) {
	var pages []Page
	err := fs.WalkDir(docs, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".md" {
			return err
		}
		for _, s := range skipped {
			if strings.HasPrefix(p, s) {
				return nil
			}
		}
		f, err := docs.Open(p)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		page := parse(bufio.NewScanner(f))
		page.Path = pagePath(p)
		if page.Title != "" {
			pages = append(pages, page)
		}
		return nil
	})
	sort.Slice(pages, func(i, j int) bool { return pages[i].Path < pages[j].Path })
	return pages, err
}

// pagePath is the URL path the site generator gives a file: directory URLs,
// with index.md and README.md as the folder's own page.
func pagePath(file string) string {
	dir, name := path.Split(strings.TrimSuffix(file, ".md"))
	if name == "index" || name == "README" {
		return dir
	}
	return dir + name + "/"
}

func parse(sc *bufio.Scanner) Page {
	var page Page
	var summary []string
	inCode, summaryDone := false, false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		switch {
		case strings.HasPrefix(line, "# ") && page.Title == "":
			page.Title = plain(line[2:])
		case strings.HasPrefix(line, "## "), strings.HasPrefix(line, "### "):
			title := plain(strings.TrimLeft(line, "# "))
			page.Sections = append(page.Sections, Section{Title: title, Anchor: slug(title)})
			summaryDone = true
		case page.Title != "" && !summaryDone:
			switch {
			case line == "" && len(summary) > 0:
				summaryDone = true
			case line != "" && !strings.HasPrefix(line, "|") && !strings.HasPrefix(line, ">") &&
				!strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "<"):
				summary = append(summary, plain(line))
			}
		}
	}
	page.Summary = shorten(strings.Join(summary, " "))
	return page
}

// plain drops Markdown links and emphasis marks.
func plain(s string) string {
	return strings.TrimSpace(mdMarks.ReplaceAllString(mdLink.ReplaceAllString(s, "$1"), ""))
}

func shorten(s string) string {
	if len(s) <= maxSummary {
		return s
	}
	cut := strings.LastIndex(s[:maxSummary], " ")
	if cut < 0 {
		cut = maxSummary
	}
	return s[:cut] + " …"
}

// slug mirrors the default heading IDs of Python-Markdown's toc extension,
// which the documentation site uses.
func slug(s string) string {
	var b strings.Builder
	for _, r := range norm.NFKD.String(s) {
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		}
	}
	out := slugDrop.ReplaceAllString(b.String(), "")
	return slugJoin.ReplaceAllString(strings.ToLower(strings.TrimSpace(out)), "-")
}
