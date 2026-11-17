package archtest

import (
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// eePath is the import path prefix of the separately licensed team features
// (ADR 0017).
const eePath = modulePath + "/ee"

// importsEE reports whether an import path points into ee/.
func importsEE(imp string) bool {
	return imp == eePath || strings.HasPrefix(imp, eePath+"/")
}

// onlyBuiltWithEETag reports whether the build constraint among the comment
// lines above the package clause excludes the file from a build without the
// "ee" tag. A file without a constraint is always built and therefore not
// guarded.
func onlyBuiltWithEETag(comments []string) bool {
	for _, line := range comments {
		if !constraint.IsGoBuild(line) {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			return false
		}
		// Every tag except "ee" is assumed to be set: the file is guarded only
		// if leaving out "ee" alone switches it off.
		return !expr.Eval(func(tag string) bool { return tag != "ee" })
	}
	return false
}

// TestCoreDoesNotDependOnEE keeps the MIT core independent of ee/: packages
// below internal/ never import it, and the composition root does so only in
// files that are left out of a build without the "ee" tag.
func TestCoreDoesNotDependOnEE(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, dir := range []string{"internal", "cmd"} {
		err = filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly|parser.ParseComments)
			if err != nil {
				return err
			}
			for _, spec := range f.Imports {
				if !importsEE(strings.Trim(spec.Path.Value, `"`)) {
					continue
				}
				rel := strings.TrimPrefix(path, root+string(os.PathSeparator))
				if dir == "internal" {
					t.Errorf("%s imports ee/: the core must not depend on it (ADR 0017)", rel)
					continue
				}
				var header []string
				for _, group := range f.Comments {
					if group.Pos() > f.Package {
						break
					}
					for _, c := range group.List {
						header = append(header, c.Text)
					}
				}
				if !onlyBuiltWithEETag(header) {
					t.Errorf("%s imports ee/ without the build constraint //go:build ee (ADR 0017)", rel)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestEEImportRules(t *testing.T) {
	if !importsEE("waylight/ee/teams") || !importsEE("waylight/ee") {
		t.Error("imports below ee/ must be recognised")
	}
	if importsEE("waylight/eex") || importsEE("waylight/internal/modules/notes") {
		t.Error("imports outside ee/ must not be recognised")
	}
	cases := []struct {
		header []string
		want   bool
	}{
		{[]string{"//go:build ee"}, true},
		{[]string{"//go:build ee && linux"}, true},
		{[]string{"// Comment.", "//go:build ee"}, true},
		{nil, false},
		{[]string{"// Comment."}, false},
		{[]string{"//go:build !ee"}, false},
		{[]string{"//go:build ee || linux"}, false},
	}
	for _, c := range cases {
		if got := onlyBuiltWithEETag(c.header); got != c.want {
			t.Errorf("onlyBuiltWithEETag(%q) = %v, want %v", c.header, got, c.want)
		}
	}
}
