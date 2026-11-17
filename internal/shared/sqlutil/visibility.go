package sqlutil

import (
	"strings"

	"waylight/internal/shared/kernel"
)

// VisibleTo returns a condition and its arguments that limit a query of a
// personal-content table to the records p may see (ADR 0018), for example
//
//	cond, args := sqlutil.VisibleTo(p, "")
//	rows, err := db.QueryContext(ctx, `SELECT id, body FROM notes WHERE `+cond, args...)
//
// It is the SQL form of kernel.Visibility.VisibleTo; a test proves both give
// the same answer. alias qualifies the columns when the query joins tables
// ("n" gives n.owner_id); leave it empty otherwise. Tables using it have the
// columns owner_id and visibility.
//
// The values are always bound as arguments, never formatted into the SQL.
func VisibleTo(p kernel.Principal, alias string) (string, []any) {
	if p.IsSystem() {
		return "1 = 1", nil
	}
	col := func(name string) string {
		if alias == "" {
			return name
		}
		return alias + "." + name
	}
	var b strings.Builder
	args := []any{p.AccountID, string(kernel.Everyone)}
	b.WriteString("(" + col("owner_id") + " = ? OR " + col("visibility") + " = ?")
	if len(p.Groups) > 0 {
		b.WriteString(" OR " + col("visibility") + " IN (")
		for i, g := range p.Groups {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("?")
			args = append(args, string(kernel.GroupVisibility(g)))
		}
		b.WriteString(")")
	}
	b.WriteString(")")
	return b.String(), args
}

// OwnedBy returns the condition for changing or deleting a record: only its
// owner may (ADR 0018), whatever its visibility. The system may change all.
func OwnedBy(p kernel.Principal, alias string) (string, []any) {
	if p.IsSystem() {
		return "1 = 1", nil
	}
	col := "owner_id"
	if alias != "" {
		col = alias + "." + col
	}
	return col + " = ?", []any{p.AccountID}
}
