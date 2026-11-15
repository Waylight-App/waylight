package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// ADR 0018: once a table carries owner_id, forgetting the visibility filter
// leaks other people's records. This rule makes the filter impossible to
// forget by hand: in a module's adapters, a function whose SQL reads,
// changes or deletes such a table must build its WHERE clause with
// sqlutil.VisibleTo or sqlutil.OwnedBy, and an INSERT into it must set
// owner_id.
//
// A function that deliberately works across all owners (the JSON export of
// the whole instance, a background job) says so with a comment line
//
//	//archtest:all-owners <reason>
//
// in its doc comment; the reason is required.

var (
	createTable = regexp.MustCompile(`(?is)CREATE TABLE (?:IF NOT EXISTS )?"?(\w+)"?\s*\((.*?)\);`)
	addOwner    = regexp.MustCompile(`(?i)ALTER TABLE "?(\w+)"? ADD COLUMN owner_id\b`)
	ownerColumn = regexp.MustCompile(`(?i)(^|[\s,(])owner_id\s`)
	tableRef    = regexp.MustCompile(`(?i)\b(FROM|UPDATE|INTO|JOIN)\s+"?(\w+)"?`)
	allOwners   = regexp.MustCompile(`^//archtest:all-owners\s+\S`)
)

// ownedTables returns the tables of one module that have an owner_id column,
// from its SQL migrations.
func ownedTables(moduleDir string) (map[string]bool, error) {
	owned := map[string]bool{}
	err := filepath.WalkDir(moduleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range createTable.FindAllStringSubmatch(string(b), -1) {
			if ownerColumn.MatchString(m[2]) {
				owned[strings.ToLower(m[1])] = true
			}
		}
		for _, m := range addOwner.FindAllStringSubmatch(string(b), -1) {
			owned[strings.ToLower(m[1])] = true
		}
		return nil
	})
	return owned, err
}

// ownershipViolations checks the adapters of every module under modulesDir.
func ownershipViolations(t *testing.T, modulesDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatal(err)
	}
	var problems []string
	fset := token.NewFileSet()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(modulesDir, e.Name())
		owned, err := ownedTables(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(owned) == 0 {
			continue
		}
		err = filepath.WalkDir(filepath.Join(dir, "adapter"), func(path string, d fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return filepath.SkipDir
			}
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil || exempt(fn) {
					continue
				}
				problems = append(problems, checkFunc(fset, fn, owned)...)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return problems
}

func exempt(fn *ast.FuncDecl) bool {
	if fn.Doc == nil {
		return false
	}
	return slices.ContainsFunc(fn.Doc.List, func(c *ast.Comment) bool { return allOwners.MatchString(c.Text) })
}

func checkFunc(fset *token.FileSet, fn *ast.FuncDecl, owned map[string]bool) []string {
	filtered := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "sqlutil" && (sel.Sel.Name == "VisibleTo" || sel.Sel.Name == "OwnedBy") {
				filtered = true
			}
		}
		return true
	})
	var problems []string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		sql, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		for _, m := range tableRef.FindAllStringSubmatch(sql, -1) {
			table := strings.ToLower(m[2])
			if !owned[table] {
				continue
			}
			pos := fset.Position(lit.Pos())
			where := pos.Filename + ":" + strconv.Itoa(pos.Line) + " " + fn.Name.Name
			if strings.EqualFold(m[1], "INTO") {
				if !strings.Contains(strings.ToLower(sql), "owner_id") {
					problems = append(problems, where+": INSERT INTO "+table+" must set owner_id from the principal")
				}
			} else if !filtered {
				problems = append(problems, where+": "+strings.ToUpper(m[1])+" "+table+" without sqlutil.VisibleTo/OwnedBy")
			}
		}
		return true
	})
	return problems
}

// TestOwnedTablesAreFiltered applies the rule to the real modules. Until the
// first module gains owner_id (ADR 0018 follow-up work) it finds nothing to
// check; TestOwnershipRuleCatchesLeaks proves that it works.
func TestOwnedTablesAreFiltered(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ownershipViolations(t, filepath.Join(root, "internal", "modules")) {
		t.Error(p)
	}
}

func TestOwnershipRuleCatchesLeaks(t *testing.T) {
	all := strings.Join(ownershipViolations(t, filepath.Join("testdata", "ownership")), "\n")
	for _, name := range []string{"leakyList", "leakyDelete", "insertWithoutOwner"} {
		if !strings.Contains(all, " "+name+":") {
			t.Errorf("rule missed %s; reported:\n%s", name, all)
		}
	}
	for _, name := range []string{"filteredList", "ownedDelete", "insertWithOwner", "exportAll", "untouchedTable"} {
		if strings.Contains(all, " "+name+":") {
			t.Errorf("rule reported %s, which is fine:\n%s", name, all)
		}
	}
}
