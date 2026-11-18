package usecase

import (
	"context"
	"fmt"
)

// demoNote is a second pinned note of the public demo: something a
// household actually keeps on its start page.
const demoNote = `# House notes 🏠

**Wi-Fi for guests:** Waylight-Guest (password on the fridge)

**Bins:** organic every week, the rest every second week – see the waste widget.

**Plumber:** Hinrichs & Sohn, +49 441 555 0177

**Server:** backups run at 3:00, updates on Sunday evening.
`

// SeedDemo adds the house notes next to the welcome note.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list notes: %w", err)
	}
	if len(existing) > 1 {
		return nil
	}
	_, err = NewSaveNote(d).Execute(ctx, "", NoteInput{Body: demoNote, Pinned: true})
	return err
}
