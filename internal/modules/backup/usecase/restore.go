package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/shared/kernel"
)

// Reauthenticator re-checks the password and 2FA code of the signed-in
// account (output port, implemented by the auth module).
type Reauthenticator interface {
	ConfirmIdentity(ctx context.Context, password, code, clientKey string) error
}

// BackupInspector checks a backup file and reads the live schema (output port).
type BackupInspector interface {
	Inspect(ctx context.Context, path string) (domain.BackupInfo, error)
	Current(ctx context.Context) ([]domain.Migration, error)
}

// DatabaseSwapper stages a file and schedules it to replace the database
// on the next start (output port, implemented by the platform).
type DatabaseSwapper interface {
	Stage(ctx context.Context, r io.Reader) (string, error)
	Discard(path string)
	Schedule(path string) error
}

// RestoreSource selects where the backup comes from.
type RestoreSource string

// Restore sources.
const (
	FromLocal  RestoreSource = "local"
	FromRemote RestoreSource = "remote"
	FromUpload RestoreSource = "upload"
)

// SafetyCopier saves the current database as an ordinary backup before a
// restore replaces it (output port, implemented by the Backups use case).
type SafetyCopier interface {
	SafetyCopy(ctx context.Context) (FileView, error)
}

// RestoreRequest is the input of Restore.
type RestoreRequest struct {
	Password, Code, ClientKey string
	// Confirmed is the owner's explicit yes to replacing the current data.
	Confirmed bool
	Source    RestoreSource
	Name      string    // local or remote backup name
	Upload    io.Reader // uploaded file
}

// RestoreResult tells the owner where the previous state went.
type RestoreResult struct {
	// SafetyCopy is the backup of the state before the restore; restoring
	// it the same way undoes the restore.
	SafetyCopy FileView
}

// PreviewView describes a stored backup before it is restored (#677).
type PreviewView struct {
	File FileView
	domain.Preview
}

// RestoreDeps bundles the ports of the restore.
type RestoreDeps struct {
	Store     Store
	Remote    RemoteStore
	RemoteCfg RemoteSettingsRepository
	Inspector BackupInspector
	Swapper   DatabaseSwapper
	Reauth    Reauthenticator
	Safety    SafetyCopier
	Alerts    kernel.AlertSink
	// Restart shuts the server down gracefully so it starts with the
	// restored database (Docker restarts the container).
	Restart func()
}

// Restorer is the input port of RestoreBackup (#46, #677).
type Restorer interface {
	// Preview checks a local backup and says what it contains.
	Preview(ctx context.Context, name string) (PreviewView, error)
	Restore(ctx context.Context, req RestoreRequest) (RestoreResult, error)
}

type restorer struct{ d RestoreDeps }

// NewRestorer returns the restore use case.
func NewRestorer(d RestoreDeps) Restorer { return &restorer{d: d} }

// ownerOnly is an allow-list: a restore replaces everyone's data, so only
// the owner's own session may start it, never an admin, a token or a
// guest, whatever the route guard lets through.
func ownerOnly(ctx context.Context) error {
	if kernel.PrincipalFrom(ctx).Role != "owner" {
		return domain.ErrRestoreNotAllowed
	}
	return nil
}

// Preview implements Restorer. The stored file is inspected in place and
// read-only; nothing is staged.
func (uc *restorer) Preview(ctx context.Context, name string) (PreviewView, error) {
	if err := ownerOnly(ctx); err != nil {
		return PreviewView{}, err
	}
	if !domain.ValidName(name) {
		return PreviewView{}, domain.ErrBackupNotFound
	}
	rc, f, err := uc.d.Store.Open(name)
	if err != nil {
		return PreviewView{}, err
	}
	_ = rc.Close()
	current, err := uc.d.Inspector.Current(ctx)
	if err != nil {
		return PreviewView{}, fmt.Errorf("read current schema: %w", err)
	}
	view := PreviewView{File: toView(f)}
	info, err := uc.d.Inspector.Inspect(ctx, uc.d.Store.Path(name))
	if err != nil {
		view.Problem = domain.ErrNotADatabase
		return view, nil
	}
	view.Preview = domain.PreviewRestore(info, current)
	return view, nil
}

// Restore implements Restorer: confirm identity, stage, validate, save the
// current state as a safety copy, schedule the swap and restart. Nothing
// changes when any step fails.
func (uc *restorer) Restore(ctx context.Context, req RestoreRequest) (RestoreResult, error) {
	if err := ownerOnly(ctx); err != nil {
		return RestoreResult{}, err
	}
	if !req.Confirmed {
		return RestoreResult{}, domain.ErrNotConfirmed
	}
	if err := uc.d.Reauth.ConfirmIdentity(ctx, req.Password, req.Code, req.ClientKey); err != nil {
		return RestoreResult{}, err
	}
	src, err := uc.open(ctx, req)
	if err != nil {
		return RestoreResult{}, err
	}
	defer src.Close()
	path, err := uc.d.Swapper.Stage(ctx, src)
	if err != nil {
		return RestoreResult{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "The backup could not be received: {detail}", "detail", err.Error())
	}
	scheduled := false
	defer func() {
		if !scheduled {
			uc.d.Swapper.Discard(path)
		}
	}()
	info, err := uc.d.Inspector.Inspect(ctx, path)
	if err != nil {
		return RestoreResult{}, domain.ErrNotADatabase
	}
	current, err := uc.d.Inspector.Current(ctx)
	if err != nil {
		return RestoreResult{}, fmt.Errorf("read current schema: %w", err)
	}
	if err := domain.ValidateRestore(info, current); err != nil {
		return RestoreResult{}, err
	}
	// The safety copy is taken after the source was staged: it is a new
	// backup and must not push the chosen one out of the retention first.
	safety, err := uc.d.Safety.SafetyCopy(ctx)
	if err != nil {
		return RestoreResult{}, domain.ErrSafetyCopy
	}
	if err := uc.d.Swapper.Schedule(path); err != nil {
		return RestoreResult{}, fmt.Errorf("schedule restore: %w", err)
	}
	scheduled = true
	uc.d.Alerts.Alert(ctx, kernel.Alert{Kind: "security", Title: "Database restore started",
		Body: "Waylight restarts with a restored backup. The state before is kept as the backup " + safety.Name +
			" and as waylight.db.before-restore.", Urgent: true})
	uc.d.Restart()
	return RestoreResult{SafetyCopy: safety}, nil
}

func (uc *restorer) open(ctx context.Context, req RestoreRequest) (io.ReadCloser, error) {
	switch req.Source {
	case FromUpload:
		if req.Upload == nil {
			return nil, kernel.NewValidationError("file", "Choose a backup file.")
		}
		return io.NopCloser(req.Upload), nil
	case FromLocal:
		if !domain.ValidName(req.Name) {
			return nil, domain.ErrBackupNotFound
		}
		rc, _, err := uc.d.Store.Open(req.Name)
		return rc, err
	case FromRemote:
		if !domain.ValidName(req.Name) {
			return nil, domain.ErrBackupNotFound
		}
		s, err := uc.d.RemoteCfg.Load(ctx)
		if err != nil {
			return nil, err
		}
		rc, err := uc.d.Remote.Get(ctx, s, req.Name)
		if err != nil {
			if errors.Is(err, kernel.ErrNotFound) {
				return nil, err
			}
			return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
		}
		return rc, nil
	}
	return nil, kernel.NewValidationError("source", "The source must be local, remote or upload.")
}
