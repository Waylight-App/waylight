// Package gateway is a fixture for TestOwnershipRuleCatchesLeaks; it is
// parsed, never built (testdata is ignored by the go tool).
package gateway

func leakyList() {
	_ = `SELECT id, body FROM notes ORDER BY id`
}

func filteredList(p any) {
	cond, args := sqlutil.VisibleTo(p, "")
	_ = `SELECT id, body FROM notes WHERE ` + cond
	_ = args
}

func leakyDelete() {
	_ = "DELETE FROM tags WHERE id = ?"
}

func ownedDelete(p any) {
	cond, _ := sqlutil.OwnedBy(p, "")
	_ = "DELETE FROM tags WHERE id = ? AND " + cond
}

func insertWithoutOwner() {
	_ = `INSERT INTO notes (id, body, visibility) VALUES (?, ?, ?)`
}

func insertWithOwner() {
	_ = `INSERT INTO notes (id, body, owner_id, visibility) VALUES (?, ?, ?, ?)`
}

// exportAll dumps every account's notes for the instance export.
//
//archtest:all-owners the instance export contains all accounts (ADR 0018)
func exportAll() {
	_ = `SELECT id, body FROM notes`
}

func untouchedTable() {
	_ = `SELECT value FROM settings`
}

// visible wraps the filter, e.g. with a special case for guests.
func visible(p any) (string, []any) { return sqlutil.VisibleTo(p, "") }

func viaHelper(p any) {
	cond, _ := visible(p)
	_ = `SELECT id FROM notes WHERE ` + cond
}
