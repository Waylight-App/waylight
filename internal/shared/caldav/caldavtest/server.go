// Package caldavtest is an in-memory CalDAV server for gateway tests. Its
// answers follow responses recorded from Nextcloud 30 (namespaces, hrefs,
// propstat layout), trimmed to what the client reads.
package caldavtest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

// Paths of the fake account (Nextcloud layout).
const (
	Root      = "/remote.php/dav/"
	Principal = "/remote.php/dav/principals/users/anna/"
	Home      = "/remote.php/dav/calendars/anna/"
	TaskList  = "/remote.php/dav/calendars/anna/tasks/"
	EventList = "/remote.php/dav/calendars/anna/personal/"
)

// Server is a fake CalDAV server with one task list.
type Server struct {
	*httptest.Server
	User, Password string

	mu      sync.Mutex
	objects map[string]object
	ctag    int
	seq     int
	// NoCTag hides getctag and sync-token, as some servers do.
	NoCTag bool
	// Requests counts requests by "METHOD kind" (e.g. "REPORT multiget").
	Requests map[string]int
}

type object struct{ etag, data string }

// New starts a server; it is closed when the test ends.
func New(t testing.TB) *Server {
	s := &Server{User: "anna", Password: "app-pass", objects: map[string]object{}, Requests: map[string]int{}}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// ListURL is the URL of the task list.
func (s *Server) ListURL() string { return s.URL + TaskList }

// Put stores an object as another app would and returns its ETag.
func (s *Server) Put(name, data string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store(name, data)
}

// Data returns the stored iCalendar text of an object ("" if missing).
func (s *Server) Data(name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.objects[name].data
}

// Names lists the stored objects.
func (s *Server) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for n := range s.objects {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Count returns how often a request kind was served.
func (s *Server) Count(kind string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Requests[kind]
}

func (s *Server) store(name, data string) string {
	s.seq++
	s.ctag++
	etag := fmt.Sprintf(`"e%d"`, s.seq)
	s.objects[name] = object{etag: etag, data: data}
	return etag
}

const taskListProps = `<d:resourcetype><d:collection/><cal:calendar/></d:resourcetype><d:displayname>Aufgaben</d:displayname><cal:supported-calendar-component-set><cal:comp name="VTODO"/></cal:supported-calendar-component-set>`

var hrefPattern = regexp.MustCompile(`<(?:D:|d:)?href>([^<]+)</(?:D:|d:)?href>`)

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, p, ok := r.BasicAuth(); !ok || u != s.User || p != s.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, _ := io.ReadAll(r.Body)
	b := string(body)
	name := strings.TrimPrefix(r.URL.Path, TaskList)
	switch {
	case r.Method == "PROPFIND" && strings.Contains(b, "current-user-principal"):
		s.Requests["PROPFIND principal"]++
		switch r.URL.Path {
		case Root:
			multi(w, response(Root, "<d:current-user-principal><d:href>"+Principal+"</d:href></d:current-user-principal>"))
		case Principal:
			multi(w, response(Principal, "<cal:calendar-home-set><d:href>"+Home+"</d:href></cal:calendar-home-set>"))
		default:
			multi(w, response(r.URL.Path, ""))
		}
	case r.Method == "PROPFIND" && strings.Contains(b, "getctag"):
		s.Requests["PROPFIND ctag"]++
		prop := fmt.Sprintf(`<cs:getctag>http://sabre.io/ns/sync/%d</cs:getctag><d:sync-token>http://sabre.io/ns/sync/%d</d:sync-token>`, s.ctag, s.ctag)
		if s.NoCTag {
			prop = ""
		}
		multi(w, response(TaskList, prop))
	case r.Method == "PROPFIND" && r.URL.Path == TaskList:
		s.Requests["PROPFIND collections"]++
		multi(w, response(TaskList, taskListProps))
	case r.Method == "PROPFIND" && r.URL.Path == Home:
		s.Requests["PROPFIND collections"]++
		multi(w, response(Home, "<d:resourcetype><d:collection/></d:resourcetype>")+
			response(TaskList, taskListProps)+
			response(EventList, `<d:resourcetype><d:collection/><cal:calendar/></d:resourcetype><d:displayname>Persönlich</d:displayname><cal:supported-calendar-component-set><cal:comp name="VEVENT"/></cal:supported-calendar-component-set>`))
	case r.Method == "REPORT" && strings.Contains(b, "calendar-multiget"):
		s.Requests["REPORT multiget"]++
		var out strings.Builder
		for _, m := range hrefPattern.FindAllStringSubmatch(b, -1) {
			href, _ := url.PathUnescape(m[1])
			n := strings.TrimPrefix(href, TaskList)
			if o, ok := s.objects[n]; ok {
				out.WriteString(response(TaskList+n, "<d:getetag>"+xmlEscape(o.etag)+"</d:getetag><cal:calendar-data>"+xmlEscape(o.data)+"</cal:calendar-data>"))
			}
		}
		multi(w, out.String())
	case r.Method == "REPORT":
		withData := strings.Contains(b, "calendar-data")
		if withData {
			s.Requests["REPORT data"]++
		} else {
			s.Requests["REPORT etags"]++
		}
		var out strings.Builder
		for _, n := range sortedKeys(s.objects) {
			o := s.objects[n]
			if !strings.Contains(o.data, "BEGIN:VTODO") {
				continue
			}
			prop := "<d:getetag>" + xmlEscape(o.etag) + "</d:getetag>"
			if withData {
				prop += "<cal:calendar-data>" + xmlEscape(o.data) + "</cal:calendar-data>"
			}
			out.WriteString(response(TaskList+n, prop))
		}
		multi(w, out.String())
	case r.Method == http.MethodGet:
		s.Requests["GET"]++
		o, ok := s.objects[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", o.etag)
		_, _ = io.WriteString(w, o.data)
	case r.Method == http.MethodPut:
		s.Requests["PUT"]++
		o, exists := s.objects[name]
		if r.Header.Get("If-None-Match") == "*" && exists || r.Header.Get("If-Match") != "" && (!exists || r.Header.Get("If-Match") != o.etag) {
			w.WriteHeader(http.StatusPreconditionFailed)
			return
		}
		w.Header().Set("ETag", s.store(name, b))
		if exists {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	case r.Method == http.MethodDelete:
		s.Requests["DELETE"]++
		if _, ok := s.objects[name]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		delete(s.objects, name)
		s.ctag++
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func sortedKeys(m map[string]object) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

func response(href, prop string) string {
	return "<d:response><d:href>" + href + "</d:href><d:propstat><d:prop>" + prop + "</d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response>"
}

func multi(w http.ResponseWriter, responses string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)
	_, _ = io.WriteString(w, `<?xml version="1.0"?>
<d:multistatus xmlns:d="DAV:" xmlns:s="http://sabredav.org/ns" xmlns:cal="urn:ietf:params:xml:ns:caldav" xmlns:cs="http://calendarserver.org/ns/" xmlns:oc="http://owncloud.org/ns" xmlns:nc="http://nextcloud.org/ns">`+responses+`</d:multistatus>`)
}
