package smarthome

import (
	"net/http"
	"testing"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/modules/smarthome/usecase"
)

func TestCompactAction(t *testing.T) {
	off := usecase.EntityView{Entity: domain.Entity{ID: "switch.coffee", State: "off"}, Action: domain.ActionToggle}
	a := compactAction(off)
	if a == nil || a.Label.String() != "Turn on" || a.Method != http.MethodPost || a.Path != "/api/smarthome/entities/switch.coffee/action" {
		t.Fatalf("toggle %+v", a)
	}
	on := off
	on.State = "on"
	if a := compactAction(on); a == nil || a.Label.String() != "Turn off" {
		t.Fatalf("toggle on %+v", a)
	}
	press := usecase.EntityView{Entity: domain.Entity{ID: "script.morning"}, Action: domain.ActionPress}
	if a := compactAction(press); a == nil || a.Label.String() != "Run" || a.ID != "smarthome.press" {
		t.Fatalf("press %+v", a)
	}
	if a := compactAction(usecase.EntityView{Entity: domain.Entity{ID: "sensor.power"}}); a != nil {
		t.Fatalf("read-only entity got %+v", a)
	}
}
