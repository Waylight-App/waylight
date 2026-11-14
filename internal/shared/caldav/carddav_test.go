package caldav

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVCards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, _ := r.BasicAuth(); u != "u" || p != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if r.Method != "REPORT" || !strings.Contains(string(body), "addressbook-query") {
			t.Errorf("%s %s", r.Method, body)
		}
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = io.WriteString(w, `<?xml version="1.0"?><d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
<d:response><d:href>/ab/1.vcf</d:href><d:propstat><d:prop><card:address-data>BEGIN:VCARD
FN:Anna
BDAY:1990-05-17
END:VCARD</card:address-data></d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response>
<d:response><d:href>/ab/</d:href><d:propstat><d:prop/><d:status>HTTP/1.1 404 Not Found</d:status></d:propstat></d:response>
</d:multistatus>`)
	}))
	defer srv.Close()
	c := Client{HTTP: srv.Client()}
	cards, err := c.VCards(context.Background(), Account{URL: srv.URL + "/ab/", Username: "u", Password: "pw"})
	if err != nil || len(cards) != 1 || !strings.Contains(cards[0], "FN:Anna") {
		t.Fatal(cards, err)
	}
	if _, err := c.VCards(context.Background(), Account{URL: srv.URL + "/ab/", Username: "u", Password: "x"}); err == nil {
		t.Fatal("unauthorized")
	}
}
