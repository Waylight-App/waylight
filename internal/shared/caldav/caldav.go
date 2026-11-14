// Package caldav is a small CalDAV client for task lists (VTODO): listing
// with ETags, multiget of changed objects, conditional writes and the
// discovery of a user's task lists. It is shared by the shopping list and
// the to-do providers; modules map its errors to their domain errors.
//
// Ring: Interface Adapters (helpers for gateways).
package caldav

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// MaxResponseBytes limits every response body that is read.
const MaxResponseBytes = 4 << 20

// Errors of conditional requests and missing objects.
var (
	ErrPreconditionFailed = errors.New("caldav: the object was changed meanwhile")
	ErrNotFound           = errors.New("caldav: the object does not exist")
	// ErrUnauthorized wraps kernel.ErrTargetAuth, so the user is told to
	// check the login.
	ErrUnauthorized = fmt.Errorf("caldav: %w", kernel.ErrTargetAuth)
	// errNoCollection wraps kernel.ErrTargetNotFound: the list address is wrong.
	errNoCollection = fmt.Errorf("caldav: %w", kernel.ErrTargetNotFound)
)

// Account is a task list (a calendar collection that holds VTODOs) with
// its credentials. URL ends with a slash.
type Account struct {
	URL      string
	Username string
	Password string
}

// Object is one calendar object resource of the collection.
type Object struct {
	Name string // last path segment, e.g. "abc.ics"
	ETag string
	Data string // iCalendar text (empty when only ETags were asked for)
}

// Client talks to CalDAV servers through an injected HTTP client.
type Client struct{ HTTP *http.Client }

// Do sends one authenticated request.
func (c Client) Do(ctx context.Context, a Account, method, target string, body io.Reader, headers map[string]string) (*http.Response, error) {
	// Every address goes through the integration URL rules again: list
	// addresses are owner-configured, but hrefs of principals and calendar
	// homes come from the server's answers.
	u, err := kernel.ParseTarget("url", target, kernel.WithTargetQuery())
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(strings.SplitN(target, "?", 2)[0], "/") {
		u.Path += "/" // collections end with a slash; ParseTarget trims it
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(a.Username, a.Password)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return res, nil
}

// StatusError maps an unexpected status to ErrPreconditionFailed,
// ErrNotFound, ErrUnauthorized or a generic error.
func StatusError(res *http.Response) error {
	switch res.StatusCode {
	case http.StatusPreconditionFailed:
		return ErrPreconditionFailed
	case http.StatusNotFound, http.StatusGone:
		return ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	default:
		return fmt.Errorf("server answered %s", res.Status)
	}
}

const xmlHeaders = "application/xml; charset=utf-8"

type multistatus struct {
	Responses []struct {
		Href      string `xml:"href"`
		Propstats []struct {
			Status       string `xml:"status"`
			ETag         string `xml:"prop>getetag"`
			CTag         string `xml:"prop>getctag"`
			SyncToken    string `xml:"prop>sync-token"`
			CalendarData string `xml:"prop>calendar-data"`
			DisplayName  string `xml:"prop>displayname"`
			ResourceType struct {
				Calendar *struct{} `xml:"calendar"`
			} `xml:"prop>resourcetype"`
			Components struct {
				Comps []struct {
					Name string `xml:"name,attr"`
				} `xml:"comp"`
			} `xml:"prop>supported-calendar-component-set"`
			Principal string `xml:"prop>current-user-principal>href"`
			Home      string `xml:"prop>calendar-home-set>href"`
		} `xml:"propstat"`
	} `xml:"response"`
}

func (c Client) multistatus(ctx context.Context, a Account, method, target, depth, body string) (multistatus, error) {
	var ms multistatus
	res, err := c.Do(ctx, a, method, target, strings.NewReader(body), map[string]string{"Depth": depth, "Content-Type": xmlHeaders})
	if err != nil {
		return ms, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return ms, errNoCollection
	}
	if res.StatusCode != http.StatusMultiStatus {
		return ms, StatusError(res)
	}
	if err := xml.NewDecoder(io.LimitReader(res.Body, MaxResponseBytes)).Decode(&ms); err != nil {
		return ms, fmt.Errorf("parse CalDAV response: %w", err)
	}
	return ms, nil
}

const todoReport = `<?xml version="1.0" encoding="utf-8"?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop><D:getetag/>%s</D:prop>
  <C:filter><C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO"/></C:comp-filter></C:filter>
</C:calendar-query>`

// objects collects the objects of a multistatus; names are the decoded
// last path segments.
func objects(ms multistatus) []Object {
	var out []Object
	for _, r := range ms.Responses {
		name := objectName(r.Href)
		if name == "" {
			continue
		}
		for _, ps := range r.Propstats {
			if ps.Status != "" && !strings.Contains(ps.Status, " 200 ") {
				continue
			}
			out = append(out, Object{Name: name, ETag: ps.ETag, Data: ps.CalendarData})
		}
	}
	return out
}

func objectName(href string) string {
	p := href
	if u, err := url.Parse(href); err == nil {
		p = u.Path
	}
	if strings.HasSuffix(p, "/") {
		return "" // the collection itself
	}
	return path.Base(p)
}

// Todos lists every VTODO object with its ETag and data.
func (c Client) Todos(ctx context.Context, a Account) ([]Object, error) {
	ms, err := c.multistatus(ctx, a, "REPORT", a.URL, "1", fmt.Sprintf(todoReport, "<C:calendar-data/>"))
	if err != nil {
		return nil, err
	}
	return objects(ms), nil
}

// ETags lists the VTODO objects with their ETags only (no data).
func (c Client) ETags(ctx context.Context, a Account) ([]Object, error) {
	ms, err := c.multistatus(ctx, a, "REPORT", a.URL, "1", fmt.Sprintf(todoReport, ""))
	if err != nil {
		return nil, err
	}
	return objects(ms), nil
}

// Multiget fetches the data of some objects in one request
// (calendar-multiget, RFC 4791 §7.9).
func (c Client) Multiget(ctx context.Context, a Account, names []string) ([]Object, error) {
	if len(names) == 0 {
		return nil, nil
	}
	base, err := url.Parse(a.URL)
	if err != nil {
		return nil, fmt.Errorf("parse list URL: %w", err)
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?>
<C:calendar-multiget xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><D:prop><D:getetag/><C:calendar-data/></D:prop>`)
	for _, n := range names {
		b.WriteString("<D:href>")
		_ = xml.EscapeText(&b, []byte(base.JoinPath(n).EscapedPath()))
		b.WriteString("</D:href>")
	}
	b.WriteString("</C:calendar-multiget>")
	ms, err := c.multistatus(ctx, a, "REPORT", a.URL, "1", b.String())
	if err != nil {
		return nil, err
	}
	return objects(ms), nil
}

// CTag returns a token that changes whenever the collection changes
// (getctag, or the sync-token of RFC 6578); "" if the server has neither.
func (c Client) CTag(ctx context.Context, a Account) (string, error) {
	ms, err := c.multistatus(ctx, a, "PROPFIND", a.URL, "0", `<?xml version="1.0" encoding="utf-8"?>
<D:propfind xmlns:D="DAV:" xmlns:CS="http://calendarserver.org/ns/"><D:prop><CS:getctag/><D:sync-token/></D:prop></D:propfind>`)
	if err != nil {
		return "", err
	}
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if ps.CTag != "" {
				return ps.CTag, nil
			}
			if ps.SyncToken != "" {
				return ps.SyncToken, nil
			}
		}
	}
	return "", nil
}

// Get reads one object.
func (c Client) Get(ctx context.Context, a Account, name string) (Object, error) {
	res, err := c.Do(ctx, a, http.MethodGet, a.URL+url.PathEscape(name), nil, nil)
	if err != nil {
		return Object{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Object{}, StatusError(res)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, MaxResponseBytes))
	if err != nil {
		return Object{}, fmt.Errorf("read object: %w", err)
	}
	return Object{Name: name, ETag: res.Header.Get("ETag"), Data: string(raw)}, nil
}

// Put writes one object. ifMatch "" with create set uses If-None-Match: *
// (never overwrites); otherwise If-Match makes the write fail with
// ErrPreconditionFailed when the object changed since that ETag. It returns
// the new ETag if the server sends one.
func (c Client) Put(ctx context.Context, a Account, name, data, ifMatch string, create bool) (string, error) {
	headers := map[string]string{"Content-Type": "text/calendar; charset=utf-8"}
	switch {
	case create:
		headers["If-None-Match"] = "*"
	case ifMatch != "":
		headers["If-Match"] = ifMatch
	}
	res, err := c.Do(ctx, a, http.MethodPut, a.URL+url.PathEscape(name), bytes.NewBufferString(data), headers)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusCreated, http.StatusNoContent, http.StatusOK:
		return res.Header.Get("ETag"), nil
	}
	return "", StatusError(res)
}

// Delete removes one object; a missing object is not an error.
func (c Client) Delete(ctx context.Context, a Account, name string) error {
	res, err := c.Do(ctx, a, http.MethodDelete, a.URL+url.PathEscape(name), nil, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusNoContent, http.StatusOK, http.StatusNotFound, http.StatusGone:
		return nil
	}
	return StatusError(res)
}

// Collection is a task list found by Discover.
type Collection struct {
	URL  string
	Name string
}

const propfindPrincipal = `<?xml version="1.0" encoding="utf-8"?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><D:prop><D:current-user-principal/><C:calendar-home-set/></D:prop></D:propfind>`

const propfindCollections = `<?xml version="1.0" encoding="utf-8"?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><D:prop><D:resourcetype/><D:displayname/><C:supported-calendar-component-set/></D:prop></D:propfind>`

// Discover finds the task lists of the account, starting at a.URL: the
// server root or DAV endpoint (via current-user-principal and
// calendar-home-set), a calendar home, or a list itself.
func (c Client) Discover(ctx context.Context, a Account) ([]Collection, error) {
	base, err := url.Parse(a.URL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	home := a.URL
	if ms, err := c.multistatus(ctx, a, "PROPFIND", a.URL, "0", propfindPrincipal); err == nil {
		h, principal := homeAndPrincipal(ms)
		if h == "" && principal != "" {
			if pm, err := c.multistatus(ctx, a, "PROPFIND", base.ResolveReference(&url.URL{Path: principal}).String(), "0", propfindPrincipal); err == nil {
				h, _ = homeAndPrincipal(pm)
			}
		}
		if h != "" {
			home = base.ResolveReference(&url.URL{Path: h}).String()
		}
	} else if errors.Is(err, ErrUnauthorized) {
		return nil, err
	}
	ms, err := c.multistatus(ctx, a, "PROPFIND", home, "1", propfindCollections)
	if err != nil {
		return nil, err
	}
	var out []Collection
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if ps.ResourceType.Calendar == nil || !supportsTodos(ps.Components.Comps) {
				continue
			}
			u := base.ResolveReference(&url.URL{Path: hrefPath(r.Href)}).String()
			if !strings.HasSuffix(u, "/") {
				u += "/"
			}
			name := strings.TrimSpace(ps.DisplayName)
			if name == "" {
				name = path.Base(strings.TrimSuffix(hrefPath(r.Href), "/"))
			}
			out = append(out, Collection{URL: u, Name: name})
		}
	}
	return out, nil
}

func hrefPath(href string) string {
	if u, err := url.Parse(href); err == nil {
		return u.Path
	}
	return href
}

func homeAndPrincipal(ms multistatus) (home, principal string) {
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if ps.Home != "" {
				home = hrefPath(ps.Home)
			}
			if ps.Principal != "" {
				principal = hrefPath(ps.Principal)
			}
		}
	}
	return home, principal
}

// supportsTodos is true when the list holds VTODOs; a server that does
// not say which components it supports allows all of them.
func supportsTodos(comps []struct {
	Name string `xml:"name,attr"`
}) bool {
	if len(comps) == 0 {
		return true
	}
	for _, c := range comps {
		if strings.EqualFold(c.Name, "VTODO") {
			return true
		}
	}
	return false
}
