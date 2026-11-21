package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/places/adapter/presenter"
	"waylight/internal/modules/places/domain"
	"waylight/internal/modules/places/usecase"
)

type fakeUC struct{ usecase.Places }

func (fakeUC) Add(_ context.Context, in usecase.Input) (usecase.PlaceView, error) {
	p, err := domain.NewPlace("1", in.Name, in.Label, in.Lat, in.Lon, time.Unix(0, 0))
	return usecase.PlaceView{Place: p}, err
}

func TestAdd(t *testing.T) {
	for body, want := range map[string]int{
		`{"name":"Home","lat":53.1,"lon":8.2}`: http.StatusCreated,
		`{"name":"","lat":53.1,"lon":8.2}`:     http.StatusUnprocessableEntity,
		`{"name":"Home","lat":95,"lon":8.2}`:   http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Add(rec, httptest.NewRequest(http.MethodPost, "/api/places", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", body, rec.Code, want)
		}
		if want != http.StatusCreated {
			continue
		}
		var got presenter.PlaceDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got.Name != "Home" || got.UsedBy == nil {
			t.Errorf("%s: %+v %v", body, got, err)
		}
	}
}
