package controller

import (
	"net/http"
	"strings"
	"time"

	"waylight/internal/shared/answers"
	"waylight/internal/shared/httpapi"
)

// maxQuery bounds the query; longer input is no question for the search.
const maxQuery = 200

// AnswerEngine answers a query of the smart search (ADR 0026).
type AnswerEngine interface {
	Answer(r *http.Request, query string, lang answers.Lang) []answers.Answer
}

// AnswerHandlers serve the answers of the smart search.
type AnswerHandlers struct{ engine func() AnswerEngine }

// NewAnswerHandlers returns the handlers; engine is read per request, so
// the sources can be set after the routes.
func NewAnswerHandlers(engine func() AnswerEngine) *AnswerHandlers {
	return &AnswerHandlers{engine: engine}
}

// Answers handles GET /api/search/answers?q=…&lang=…: up to three answers in
// the given language, or the one of Accept-Language.
func (h *AnswerHandlers) Answers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = r.Header.Get("Accept-Language")
	}
	out := AnswersDTO{Answers: []AnswerDTO{}}
	if e := h.engine(); e != nil && q != "" && len(q) <= maxQuery {
		for _, a := range e.Answer(r, q, answers.ParseLang(lang)) {
			out.Answers = append(out.Answers, toAnswerDTO(a))
		}
	}
	httpapi.WriteJSON(w, http.StatusOK, out)
}

// AnswersDTO is the body of GET /api/search/answers.
type AnswersDTO struct {
	Answers []AnswerDTO `json:"answers"`
}

// AnswerDTO is one answer; fields depend on kind.
type AnswerDTO struct {
	Kind        string     `json:"kind"`
	Intent      string     `json:"intent"`
	Title       string     `json:"title"`
	WidgetKind  string     `json:"widgetKind,omitempty"`
	At          *time.Time `json:"at,omitempty"`
	Value       string     `json:"value,omitempty"`
	Detail      string     `json:"detail,omitempty"`
	Percent     *float64   `json:"percent,omitempty"`
	Severity    string     `json:"severity,omitempty"`
	State       string     `json:"state,omitempty"`
	Text        string     `json:"text,omitempty"`
	Suggestions []string   `json:"suggestions,omitempty"`
	Items       []ItemDTO  `json:"items,omitempty"`
	Empty       string     `json:"empty,omitempty"`
}

// ItemDTO is one row of a list answer.
type ItemDTO struct {
	Label   string   `json:"label"`
	Value   string   `json:"value"`
	State   string   `json:"state,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
}

func toAnswerDTO(a answers.Answer) AnswerDTO {
	d := AnswerDTO{
		Kind: string(a.Kind), Intent: a.Intent, Title: a.Title, WidgetKind: a.WidgetKind,
		Value: a.Value, Detail: a.Detail, Percent: a.Percent, Severity: string(a.Severity), State: string(a.State),
		Text: a.Text, Suggestions: a.Suggestions, Empty: a.Empty,
	}
	if !a.At.IsZero() {
		at := a.At.UTC()
		d.At = &at
	}
	if a.Kind == answers.List {
		d.Items = make([]ItemDTO, len(a.Items))
		for i, it := range a.Items {
			d.Items[i] = ItemDTO{Label: it.Label, Value: it.Value, State: string(it.State), Percent: it.Percent}
		}
	}
	return d
}
