package gateway

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/notes/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// Highlight markers produced by the SQL below (char(2) and char(3)).
const (
	hitStart = "\x02"
	hitEnd   = "\x03"
)

// Search implements usecase.SearchIndex with FTS5: best matches first,
// title matches weigh more than body matches. Only notes the principal may
// see are found.
func (r *Repository) Search(ctx context.Context, ftsQuery string, limit int) ([]domain.SearchHit, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "n")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT n.id,
			highlight(notes_fts, 0, char(2), char(3)),
			snippet(notes_fts, 1, char(2), char(3), '…', 12)
		FROM notes_fts JOIN notes n ON n.rowid = notes_fts.rowid
		WHERE notes_fts MATCH ? AND `+cond+`
		ORDER BY bm25(notes_fts, 5.0, 1.0)
		LIMIT ?`, append(append([]any{ftsQuery}, args...), limit)...)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	defer rows.Close()
	out := []domain.SearchHit{}
	for rows.Next() {
		var id, title, snippet string
		if err := rows.Scan(&id, &title, &snippet); err != nil {
			return nil, fmt.Errorf("scan hit: %w", err)
		}
		out = append(out, domain.SearchHit{ID: id, Title: segments(title), Snippet: segments(snippet)})
	}
	return out, rows.Err()
}

// segments splits marked-up FTS output into plain and highlighted parts.
// Stray control characters from note text are dropped.
func segments(s string) []domain.Segment {
	out := []domain.Segment{}
	hit := false
	for s != "" {
		marker := hitStart
		if hit {
			marker = hitEnd
		}
		i := strings.Index(s, marker)
		if i < 0 {
			i = len(s)
		}
		if text := strings.NewReplacer(hitStart, "", hitEnd, "").Replace(s[:i]); text != "" {
			out = append(out, domain.Segment{Text: text, Hit: hit})
		}
		if i == len(s) {
			break
		}
		s, hit = s[i+1:], !hit
	}
	return out
}

// ListRevisions implements usecase.RevisionRepository: the history of a
// note the principal may see.
func (r *Repository) ListRevisions(ctx context.Context, noteID string) ([]domain.Revision, error) {
	cond, args := sqlutil.VisibleTo(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, note_id, title, body, created_at FROM note_revisions
		WHERE note_id = ? AND note_id IN (SELECT id FROM notes WHERE `+cond+`) ORDER BY created_at DESC, rowid DESC`, append([]any{noteID}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query revisions: %w", err)
	}
	defer rows.Close()
	var out []domain.Revision
	for rows.Next() {
		var rev domain.Revision
		var created string
		if err := rows.Scan(&rev.ID, &rev.NoteID, &rev.Title, &rev.Body, &created); err != nil {
			return nil, fmt.Errorf("scan revision: %w", err)
		}
		if rev.CreatedAt, err = sqlutil.ParseTime(created); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}

// AddRevision implements usecase.RevisionRepository. Only the note's owner
// creates revisions: a refused edit of someone else's note must not leave
// one behind (the use case keeps the revision before saving).
func (r *Repository) AddRevision(ctx context.Context, rev domain.Revision) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	//nolint:gosec // owned comes from sqlutil with bound arguments
	_, err := r.db.ExecContext(ctx, `INSERT INTO note_revisions (id, note_id, title, body, created_at) SELECT ?, ?, ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM notes WHERE id = ? AND `+owned+`)`,
		append([]any{rev.ID, rev.NoteID, rev.Title, rev.Body, sqlutil.FormatTime(rev.CreatedAt), rev.NoteID}, args...)...)
	if err != nil {
		return fmt.Errorf("insert revision: %w", err)
	}
	return nil
}

// DeleteRevisions implements usecase.RevisionRepository (pruning old
// revisions of an own note).
func (r *Repository) DeleteRevisions(ctx context.Context, ids ...string) error {
	owned, args := sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
	for _, id := range ids {
		//nolint:gosec // owned comes from sqlutil with bound arguments
		if _, err := r.db.ExecContext(ctx, `DELETE FROM note_revisions WHERE id = ? AND note_id IN (SELECT id FROM notes WHERE `+owned+`)`,
			append([]any{id}, args...)...); err != nil {
			return fmt.Errorf("delete revision: %w", err)
		}
	}
	return nil
}
