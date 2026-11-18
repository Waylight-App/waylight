package demo

import (
	"net/http"
	"time"
)

// route answers the requests of one example service. Fixture files add
// their routes in init, so each service lives in its own file.
type route struct {
	match func(host, path string) bool
	serve func(req *http.Request, now time.Time) (status int, contentType string, body string)
}

var routes []route

// handle registers a fixture route; it is only called from init functions.
func handle(match func(host, path string) bool, serve func(req *http.Request, now time.Time) (int, string, string)) {
	routes = append(routes, route{match: match, serve: serve})
}

// hostIs matches one exact host name.
func hostIs(name string) func(host, path string) bool {
	return func(host, _ string) bool { return host == name }
}

// jsonOK is the common answer of a fixture.
func jsonOK(body string) (int, string, string) { return http.StatusOK, "application/json", body }
