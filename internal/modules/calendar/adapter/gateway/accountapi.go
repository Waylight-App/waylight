package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"waylight/internal/modules/calendar/domain"
)

// Limits of the provider APIs (#282).
const (
	maxAPIResponseBytes = 8 << 20
	// maxPages bounds pagination; 10 pages of 250–500 events cover any
	// realistic month.
	maxPages       = 10
	googlePageSize = "250"
	graphPageSize  = "500"
)

// AccountAPI implements usecase.AccountAPI for Google Calendar (API v3) and
// Microsoft 365 (Graph). The base URLs are fields so tests can point them
// at a fake server.
type AccountAPI struct {
	client    *http.Client
	GoogleURL string
	GraphURL  string
}

// NewAccountAPI returns the API client with the public endpoints.
func NewAccountAPI(client *http.Client) *AccountAPI {
	return &AccountAPI{client: client, GoogleURL: "https://www.googleapis.com/calendar/v3", GraphURL: "https://graph.microsoft.com/v1.0"}
}

// statusError keeps the provider's status code readable in the agenda.
type statusError struct{ code int }

func (e statusError) Error() string { return fmt.Sprintf("the provider answered %d", e.code) }

func (a *AccountAPI) do(ctx context.Context, method, rawURL, token string, body any, header map[string]string, dst any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, r)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return statusError{res.StatusCode}
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(res.Body, maxAPIResponseBytes)).Decode(dst)
}

// Profile implements usecase.AccountAPI.
func (a *AccountAPI) Profile(ctx context.Context, p domain.Provider, token string) (string, error) {
	if p == domain.ProviderMicrosoft {
		var me struct {
			Mail string `json:"mail"`
			UPN  string `json:"userPrincipalName"`
			Name string `json:"displayName"`
		}
		if err := a.do(ctx, http.MethodGet, a.GraphURL+"/me?$select=mail,userPrincipalName,displayName", token, nil, nil, &me); err != nil {
			return "", err
		}
		return firstOf(me.Mail, me.UPN, me.Name), nil
	}
	// The ID of the primary calendar is the account's e-mail address.
	var cal struct {
		ID string `json:"id"`
	}
	if err := a.do(ctx, http.MethodGet, a.GoogleURL+"/calendars/primary", token, nil, nil, &cal); err != nil {
		return "", err
	}
	return cal.ID, nil
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Calendars implements usecase.AccountAPI.
func (a *AccountAPI) Calendars(ctx context.Context, p domain.Provider, token string) ([]domain.AccountCalendar, error) {
	var out []domain.AccountCalendar
	if p == domain.ProviderMicrosoft {
		var res struct {
			Value []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				HexColor string `json:"hexColor"`
				CanEdit  bool   `json:"canEdit"`
			} `json:"value"`
		}
		if err := a.do(ctx, http.MethodGet, a.GraphURL+"/me/calendars?$top=50&$select=id,name,hexColor,canEdit", token, nil, nil, &res); err != nil {
			return nil, err
		}
		for _, c := range res.Value {
			out = append(out, domain.AccountCalendar{ID: c.ID, Name: c.Name, Color: domain.SafeColor(c.HexColor), Writable: c.CanEdit})
		}
		return out, nil
	}
	var res struct {
		Items []struct {
			ID              string `json:"id"`
			Summary         string `json:"summary"`
			SummaryOverride string `json:"summaryOverride"`
			BackgroundColor string `json:"backgroundColor"`
			AccessRole      string `json:"accessRole"`
		} `json:"items"`
	}
	if err := a.do(ctx, http.MethodGet, a.GoogleURL+"/users/me/calendarList?maxResults=250", token, nil, nil, &res); err != nil {
		return nil, err
	}
	for _, c := range res.Items {
		out = append(out, domain.AccountCalendar{ID: c.ID, Name: firstOf(c.SummaryOverride, c.Summary, c.ID), Color: domain.SafeColor(c.BackgroundColor),
			Writable: c.AccessRole == "owner" || c.AccessRole == "writer"})
	}
	return out, nil
}

// Events implements usecase.AccountAPI; recurring events come expanded.
func (a *AccountAPI) Events(ctx context.Context, p domain.Provider, token, calendarID string, from, to time.Time) ([]domain.Event, error) {
	if p == domain.ProviderMicrosoft {
		return a.graphEvents(ctx, token, calendarID, from, to)
	}
	return a.googleEvents(ctx, token, calendarID, from, to)
}

type googleTime struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"`
}

// parse returns the time and whether it is a date (all-day).
func (t googleTime) parse() (time.Time, bool, error) {
	if t.Date != "" {
		d, err := time.Parse(time.DateOnly, t.Date)
		return d, true, err
	}
	d, err := time.Parse(time.RFC3339, t.DateTime)
	return d, false, err
}

func (a *AccountAPI) googleEvents(ctx context.Context, token, calendarID string, from, to time.Time) ([]domain.Event, error) {
	q := url.Values{
		"timeMin": {from.UTC().Format(time.RFC3339)}, "timeMax": {to.UTC().Format(time.RFC3339)},
		"singleEvents": {"true"}, "orderBy": {"startTime"}, "maxResults": {googlePageSize},
	}
	var out []domain.Event
	for range maxPages {
		var res struct {
			Items []struct {
				ID          string     `json:"id"`
				Status      string     `json:"status"`
				Summary     string     `json:"summary"`
				Location    string     `json:"location"`
				Description string     `json:"description"`
				HTMLLink    string     `json:"htmlLink"`
				Start       googleTime `json:"start"`
				End         googleTime `json:"end"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}
		u := a.GoogleURL + "/calendars/" + url.PathEscape(calendarID) + "/events?" + q.Encode()
		if err := a.do(ctx, http.MethodGet, u, token, nil, nil, &res); err != nil {
			return nil, err
		}
		for _, it := range res.Items {
			if it.Status == "cancelled" {
				continue
			}
			start, allDay, err1 := it.Start.parse()
			end, _, err2 := it.End.parse()
			if err1 != nil || err2 != nil {
				continue // an event the API describes oddly is left out, not the whole calendar
			}
			out = append(out, domain.Event{UID: "google:" + it.ID, Summary: it.Summary, Location: it.Location, Description: it.Description,
				URL: domain.SafeLink(it.HTMLLink), Start: start, End: end, AllDay: allDay})
		}
		if res.NextPageToken == "" {
			break
		}
		q.Set("pageToken", res.NextPageToken)
	}
	return out, nil
}

// graphTimeLayout is how Graph writes times when asked for UTC.
const graphTimeLayout = "2006-01-02T15:04:05.9999999"

type graphTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

func (a *AccountAPI) graphEvents(ctx context.Context, token, calendarID string, from, to time.Time) ([]domain.Event, error) {
	q := url.Values{
		"startDateTime": {from.UTC().Format(time.RFC3339)}, "endDateTime": {to.UTC().Format(time.RFC3339)},
		"$top":    {graphPageSize},
		"$select": {"id,subject,location,bodyPreview,webLink,start,end,isAllDay,isCancelled"},
	}
	next := a.GraphURL + "/me/calendars/" + url.PathEscape(calendarID) + "/calendarView?" + q.Encode()
	// Times come back in UTC, so no time zone table is needed here.
	header := map[string]string{"Prefer": `outlook.timezone="UTC"`}
	var out []domain.Event
	for page := 0; page < maxPages && next != ""; page++ {
		var res struct {
			Value []struct {
				ID          string    `json:"id"`
				Subject     string    `json:"subject"`
				BodyPreview string    `json:"bodyPreview"`
				WebLink     string    `json:"webLink"`
				IsAllDay    bool      `json:"isAllDay"`
				IsCancelled bool      `json:"isCancelled"`
				Start       graphTime `json:"start"`
				End         graphTime `json:"end"`
				Location    struct {
					DisplayName string `json:"displayName"`
				} `json:"location"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}
		if err := a.do(ctx, http.MethodGet, next, token, nil, header, &res); err != nil {
			return nil, err
		}
		for _, it := range res.Value {
			if it.IsCancelled {
				continue
			}
			start, err1 := time.ParseInLocation(graphTimeLayout, it.Start.DateTime, time.UTC)
			end, err2 := time.ParseInLocation(graphTimeLayout, it.End.DateTime, time.UTC)
			if err1 != nil || err2 != nil {
				continue
			}
			out = append(out, domain.Event{UID: "microsoft:" + it.ID, Summary: it.Subject, Location: it.Location.DisplayName,
				Description: it.BodyPreview, URL: domain.SafeLink(it.WebLink), Start: start, End: end, AllDay: it.IsAllDay})
		}
		// Only follow links to the same API, never to a host the response names.
		next = ""
		if strings.HasPrefix(res.NextLink, a.GraphURL+"/") {
			next = res.NextLink
		}
	}
	return out, nil
}

// Create implements usecase.AccountAPI.
func (a *AccountAPI) Create(ctx context.Context, p domain.Provider, token, calendarID string, e domain.Event) error {
	if p == domain.ProviderMicrosoft {
		layout := "2006-01-02T15:04:05"
		body := map[string]any{
			"subject":  e.Summary,
			"start":    graphTime{DateTime: e.Start.UTC().Format(layout), TimeZone: "UTC"},
			"end":      graphTime{DateTime: e.End.UTC().Format(layout), TimeZone: "UTC"},
			"isAllDay": e.AllDay,
		}
		if e.Location != "" {
			body["location"] = map[string]string{"displayName": e.Location}
		}
		return a.do(ctx, http.MethodPost, a.GraphURL+"/me/calendars/"+url.PathEscape(calendarID)+"/events", token, body, nil, nil)
	}
	start, end := googleTime{DateTime: e.Start.UTC().Format(time.RFC3339)}, googleTime{DateTime: e.End.UTC().Format(time.RFC3339)}
	if e.AllDay {
		start, end = googleTime{Date: e.Start.Format(time.DateOnly)}, googleTime{Date: e.End.Format(time.DateOnly)}
	}
	body := map[string]any{"summary": e.Summary, "start": start, "end": end}
	if e.Location != "" {
		body["location"] = e.Location
	}
	return a.do(ctx, http.MethodPost, a.GoogleURL+"/calendars/"+url.PathEscape(calendarID)+"/events", token, body, nil, nil)
}
