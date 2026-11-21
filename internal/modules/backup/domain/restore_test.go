package domain

import (
	"errors"
	"testing"
)

func TestValidateRestore(t *testing.T) {
	current := []Migration{{"auth", "0001"}, {"auth", "0002"}, {"todos", "0001"}}
	cases := []struct {
		name string
		b    BackupInfo
		want error
	}{
		{"same version", BackupInfo{Integrity: "ok", Migrations: current}, nil},
		{"older backup", BackupInfo{Integrity: "ok", Migrations: []Migration{{"auth", "0001"}}}, nil},
		{"newer backup", BackupInfo{Integrity: "ok", Migrations: append([]Migration{{"auth", "0003"}}, current...)}, ErrTooNew},
		{"foreign database", BackupInfo{Integrity: "ok"}, ErrNotADatabase},
		{"corrupt", BackupInfo{Integrity: "*** in database main ***", Migrations: current}, ErrCorrupt},
	}
	for _, tc := range cases {
		if err := ValidateRestore(tc.b, current); !errors.Is(err, tc.want) {
			t.Errorf("%s: %v, want %v", tc.name, err, tc.want)
		}
	}
}

func TestPreviewRestore(t *testing.T) {
	current := []Migration{{"auth", "0001"}, {"auth", "0002"}, {"todos", "0001"}}
	older := BackupInfo{Integrity: "ok", Migrations: []Migration{{"auth", "0001"}}, Tabs: 2, Widgets: 9}
	p := PreviewRestore(older, current)
	if p.Problem != nil || p.Modules != 1 || p.Updates != 2 || p.Tabs != 2 || p.Widgets != 9 {
		t.Fatalf("older backup: %+v", p)
	}
	if p := PreviewRestore(BackupInfo{Integrity: "ok", Migrations: current}, current); p.Problem != nil || p.Updates != 0 || p.Modules != 2 {
		t.Fatalf("same version: %+v", p)
	}
	if p := PreviewRestore(BackupInfo{Integrity: "broken", Migrations: current}, current); !errors.Is(p.Problem, ErrCorrupt) {
		t.Fatalf("corrupt: %+v", p)
	}
}
