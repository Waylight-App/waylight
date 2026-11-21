package modules

import (
	"errors"
	"testing"

	"waylight/internal/shared/kernel"
)

func TestStatusOf(t *testing.T) {
	notConfigured := kernel.NewPublicError(kernel.ErrNotFound, "Immich is not configured yet.")
	if s := StatusOf(nil, notConfigured); s.State != StateOK {
		t.Fatalf("ok: %+v", s)
	}
	if s := StatusOf(notConfigured, notConfigured); s.State != StateSetup || s.Action != ActionConnect || s.Reason != notConfigured {
		t.Fatalf("setup: %+v", s)
	}
	refused := kernel.NewPublicError(kernel.ErrUnauthorized, "The API key is not accepted.")
	if s := StatusOf(refused, notConfigured); s.State != StateFailing || s.Action != ActionFix {
		t.Fatalf("refused: %+v", s)
	}
	if s := StatusOf(errors.New("dial tcp: refused"), notConfigured); s.State != StateFailing || s.Action != ActionCheck || s.Reason.Template != "The service does not answer." {
		t.Fatalf("down: %+v", s)
	}
}
