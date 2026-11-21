package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/shared/kernel"
)

type fakeReauth struct{ err error }

func (f fakeReauth) ConfirmIdentity(context.Context, string, string, string) error { return f.err }

type fakeInspector struct {
	info    domain.BackupInfo
	err     error
	current []domain.Migration
}

func (f fakeInspector) Inspect(context.Context, string) (domain.BackupInfo, error) {
	return f.info, f.err
}
func (f fakeInspector) Current(context.Context) ([]domain.Migration, error) { return f.current, nil }

type fakeSwapper struct {
	staged, discarded, scheduled string
	content                      string
}

func (f *fakeSwapper) Stage(_ context.Context, r io.Reader) (string, error) {
	b, _ := io.ReadAll(r)
	f.content, f.staged = string(b), "/data/.restore-upload-1"
	return f.staged, nil
}
func (f *fakeSwapper) Discard(p string) { f.discarded = p }
func (f *fakeSwapper) Schedule(p string) error {
	f.scheduled = p
	return nil
}

type fakeSafety struct {
	err   error
	calls int
}

func (f *fakeSafety) SafetyCopy(context.Context) (FileView, error) {
	f.calls++
	return FileView{Name: "waylight-20261121-120000.db"}, f.err
}

func TestRestore(t *testing.T) {
	owner := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
	migrations := []domain.Migration{{Module: "auth", Version: "0001"}}
	good := domain.BackupInfo{Integrity: "ok", Migrations: migrations}
	name := "waylight-20261111-030000.db"
	store := &memStore{files: map[string]domain.File{name: {Name: name, CreatedAt: time.Now()}}}

	type outcome struct {
		sw        *fakeSwapper
		safety    *fakeSafety
		alerts    *alertRec
		restarted bool
		res       RestoreResult
		err       error
	}
	run := func(c context.Context, reauth error, info domain.BackupInfo, safetyErr error, req RestoreRequest) outcome {
		o := outcome{sw: &fakeSwapper{}, safety: &fakeSafety{err: safetyErr}, alerts: &alertRec{}}
		uc := NewRestorer(RestoreDeps{Store: store, Inspector: fakeInspector{info: info, current: migrations}, Swapper: o.sw,
			Reauth: fakeReauth{reauth}, Safety: o.safety, Alerts: o.alerts, Restart: func() { o.restarted = true }})
		o.res, o.err = uc.Restore(c, req)
		return o
	}
	upload := func() RestoreRequest {
		return RestoreRequest{Confirmed: true, Source: FromUpload, Upload: strings.NewReader("sqlite")}
	}
	// nothingChanged: no swap scheduled, no restart, the staged file is gone.
	nothingChanged := func(t *testing.T, label string, o outcome) {
		t.Helper()
		if o.err == nil || o.restarted || o.sw.scheduled != "" || o.sw.discarded != o.sw.staged {
			t.Fatalf("%s: %v restarted=%v %+v", label, o.err, o.restarted, o.sw)
		}
	}

	o := run(owner, nil, good, nil, upload())
	if o.err != nil || !o.restarted || o.sw.scheduled != o.sw.staged || o.sw.discarded != "" || o.sw.content != "sqlite" ||
		len(o.alerts.got) != 1 || o.safety.calls != 1 || o.res.SafetyCopy.Name != "waylight-20261121-120000.db" {
		t.Fatalf("upload: %+v", o)
	}
	if !strings.Contains(o.alerts.got[0].Body, o.res.SafetyCopy.Name) {
		t.Fatalf("the alert does not name the safety copy: %q", o.alerts.got[0].Body)
	}
	if o := run(owner, nil, good, nil, RestoreRequest{Confirmed: true, Source: FromLocal, Name: name}); o.err != nil || o.sw.content != "data" {
		t.Fatalf("local: %+v", o)
	}

	// Without the explicit confirmation nothing is read, staged or copied.
	o = run(owner, nil, good, nil, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("x")})
	if !errors.Is(o.err, domain.ErrNotConfirmed) || o.sw.staged != "" || o.safety.calls != 0 || o.restarted {
		t.Fatalf("unconfirmed: %+v", o)
	}
	// Wrong password: nothing is read or staged.
	o = run(owner, domain.ErrBackupNotFound, good, nil, upload())
	if o.err == nil || o.restarted || o.sw.staged != "" || o.safety.calls != 0 {
		t.Fatalf("reauth: %+v", o)
	}
	// A newer backup is rejected before the safety copy is taken.
	newer := domain.BackupInfo{Integrity: "ok", Migrations: append([]domain.Migration{{Module: "auth", Version: "0002"}}, migrations...)}
	o = run(owner, nil, newer, nil, upload())
	nothingChanged(t, "too new", o)
	if !errors.Is(o.err, domain.ErrTooNew) || o.safety.calls != 0 {
		t.Fatalf("too new: %+v", o)
	}
	nothingChanged(t, "corrupt", run(owner, nil, domain.BackupInfo{Integrity: "broken", Migrations: migrations}, nil, upload()))
	// Without a safety copy there is no restore.
	o = run(owner, nil, good, errors.New("disk full"), upload())
	nothingChanged(t, "safety copy failed", o)
	if !errors.Is(o.err, domain.ErrSafetyCopy) {
		t.Fatalf("safety copy: %v", o.err)
	}
	if o := run(owner, nil, good, nil, RestoreRequest{Confirmed: true, Source: FromLocal, Name: "../../etc/passwd"}); !errors.Is(o.err, domain.ErrBackupNotFound) {
		t.Fatalf("traversal: %v", o.err)
	}

	// Only the owner: every other role is refused before anything happens.
	for _, role := range []string{"admin", "member", "guest", "kiosk", "token", "system", ""} {
		c := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "x", Role: role})
		o := run(c, nil, good, nil, upload())
		if !errors.Is(o.err, kernel.ErrForbidden) || o.sw.staged != "" || o.safety.calls != 0 || o.restarted {
			t.Fatalf("%q: %+v", role, o)
		}
	}
}

func TestPreview(t *testing.T) {
	owner := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
	migrations := []domain.Migration{{Module: "auth", Version: "0001"}, {Module: "auth", Version: "0002"}}
	name := "waylight-20261111-030000.db"
	store := &memStore{files: map[string]domain.File{name: {Name: name, Size: 4096, CreatedAt: time.Now()}}}
	preview := func(c context.Context, insp fakeInspector, n string) (PreviewView, error) {
		insp.current = migrations
		return NewRestorer(RestoreDeps{Store: store, Inspector: insp}).Preview(c, n)
	}

	v, err := preview(owner, fakeInspector{info: domain.BackupInfo{Integrity: "ok", Migrations: migrations[:1], Tabs: 3, Widgets: 12}}, name)
	if err != nil || v.Problem != nil || v.File.Size != 4096 || v.Tabs != 3 || v.Widgets != 12 || v.Updates != 1 {
		t.Fatalf("preview: %+v %v", v, err)
	}
	if v, _ := preview(owner, fakeInspector{err: errors.New("not sqlite")}, name); !errors.Is(v.Problem, domain.ErrNotADatabase) {
		t.Fatalf("garbage: %+v", v)
	}
	if _, err := preview(owner, fakeInspector{}, "../waylight.db"); !errors.Is(err, domain.ErrBackupNotFound) {
		t.Fatalf("traversal: %v", err)
	}
	if _, err := preview(owner, fakeInspector{}, "waylight-20200101-000000.db"); !errors.Is(err, domain.ErrBackupNotFound) {
		t.Fatalf("missing: %v", err)
	}
	admin := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "a", Role: "admin"})
	if _, err := preview(admin, fakeInspector{}, name); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("admin: %v", err)
	}
}
