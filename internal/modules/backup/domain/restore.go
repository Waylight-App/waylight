package domain

import "waylight/internal/shared/kernel"

// Migration identifies an applied schema migration.
type Migration struct{ Module, Version string }

// BackupInfo is what a backup file contains, as found by the inspector.
type BackupInfo struct {
	// Integrity is the result of PRAGMA integrity_check ("ok" when sound).
	Integrity  string
	Migrations []Migration
	// Tabs and Widgets count the dashboard content, so the owner can tell
	// backups apart before restoring one; zero without a dashboard schema.
	Tabs, Widgets int
}

// Preview is what the owner sees about a backup before restoring it (#677).
type Preview struct {
	// Problem is why the backup cannot be restored; nil when it can.
	Problem error
	// Modules counts the modules with data in the backup.
	Modules       int
	Tabs, Widgets int
	// Updates counts the schema updates the next start applies to an
	// older backup.
	Updates int
}

// PreviewRestore describes a backup for the confirmation step.
func PreviewRestore(b BackupInfo, current []Migration) Preview {
	p := Preview{Problem: ValidateRestore(b, current), Tabs: b.Tabs, Widgets: b.Widgets}
	modules := map[string]bool{}
	have := make(map[Migration]bool, len(b.Migrations))
	for _, m := range b.Migrations {
		modules[m.Module] = true
		have[m] = true
	}
	p.Modules = len(modules)
	for _, m := range current {
		if !have[m] {
			p.Updates++
		}
	}
	return p
}

// Restore errors.
var (
	ErrNotADatabase = kernel.NewPublicError(kernel.ErrInvalid, "The file is not a Waylight backup.")
	ErrCorrupt      = kernel.NewPublicError(kernel.ErrInvalid, "The backup is damaged (integrity check failed).")
	ErrTooNew       = kernel.NewPublicError(kernel.ErrInvalid,
		"The backup was made by a newer version of Waylight. Update first, then restore it.")
	ErrRestoreNotAllowed = kernel.NewPublicError(kernel.ErrForbidden, "Only the owner can restore backups in the browser.")
	ErrSafetyCopy        = kernel.NewPublicError(kernel.ErrUnavailable,
		"The current data could not be saved as a safety copy, so nothing was restored. Check the free disk space.")
	ErrNotConfirmed = kernel.NewValidationError("confirm", "Confirm that the current data will be replaced.")
)

// ValidateRestore checks a backup before it may replace the database: it
// must be sound, come from Waylight (auth schema present) and must not
// contain migrations unknown to the running version. Older backups are
// fine; the missing migrations run on the next start.
func ValidateRestore(b BackupInfo, current []Migration) error {
	if b.Integrity != "ok" {
		return ErrCorrupt
	}
	known := make(map[Migration]bool, len(current))
	for _, m := range current {
		known[m] = true
	}
	hasAuth := false
	for _, m := range b.Migrations {
		if m.Module == "auth" {
			hasAuth = true
		}
		if !known[m] {
			return ErrTooNew
		}
	}
	if !hasAuth {
		return ErrNotADatabase
	}
	return nil
}
