package controller

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// pageStyle is inline so the page loads nothing; the CSP allows exactly
// this text by its hash.
const pageStyle = `:root{color-scheme:light dark;font:18px/1.5 system-ui,sans-serif}` +
	`body{margin:0 auto;padding:1.5rem 1rem;max-width:26rem}` +
	`input,button{box-sizing:border-box;width:100%;font:inherit;padding:.7rem;margin:.5rem 0}` +
	`button{cursor:pointer}.err{color:#b00020}@media(prefers-color-scheme:dark){.err{color:#ff8a80}}`

var pageCSP = "default-src 'none'; style-src 'sha256-" + styleHash() + "'; form-action 'self'; " +
	"base-uri 'none'; frame-ancestors 'none'"

func styleHash() string {
	sum := sha256.Sum256([]byte(pageStyle))
	return base64.StdEncoding.EncodeToString(sum[:])
}

var pageTemplate = template.Must(template.New("paste").Parse(`<!doctype html>
<html lang="{{.Lang}}"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex"><title>{{.Title}}</title>
<style>{{.Style}}</style></head><body>
<h1>{{.Title}}</h1>
{{if .Message}}<p{{if .Failed}} class="err" role="alert"{{end}}>{{.Message}}</p>{{end}}
{{if .Form}}<form method="post" autocomplete="off">
<label for="v">{{.Label}}</label>
<input id="v" name="value" type="password" required maxlength="4096" autocomplete="off" autocapitalize="off" spellcheck="false" autofocus>
<button type="submit">{{.Send}}</button></form>
{{if .AppLink}}<p><a href="{{.AppLink}}">{{.OpenApp}}</a></p>{{end}}{{end}}
</body></html>
`))

type pageText struct {
	Title, Label, Send, OpenApp, Done, Expired, Many, Value string
}

var pageTexts = map[string]pageText{
	"en": {"Paste from phone", "Key", "Send", "Open in the Waylight app",
		"Sent. You can close this page and go back to the setup assistant.",
		"This code has expired or was already used. Show a new code in the setup assistant and scan it again.",
		"Too many wrong codes. Please wait a few minutes.",
		"Enter the key; it must not be empty or longer than 4096 characters."},
	"de": {"Vom Handy einfügen", "Schlüssel", "Senden", "In der Waylight-App öffnen",
		"Gesendet. Du kannst diese Seite schließen und zum Einrichtungsassistenten zurückgehen.",
		"Dieser Code ist abgelaufen oder wurde schon benutzt. Zeige im Einrichtungsassistenten einen neuen Code an und scanne ihn erneut.",
		"Zu viele falsche Codes. Bitte warte ein paar Minuten.",
		"Gib den Schlüssel ein; er darf nicht leer und nicht länger als 4096 Zeichen sein."},
}

type pageData struct {
	Lang, Title, Label, Send, OpenApp, Message string
	Style                                      template.CSS
	AppLink                                    template.URL
	Form, Failed                               bool
}

// hostForm accepts what a Host header of a server looks like and nothing
// that could break out of the link built from it.
var hostForm = regexp.MustCompile(`^[A-Za-z0-9.-]+(:[0-9]{1,5})?$|^\[[0-9A-Fa-f:.]+\](:[0-9]{1,5})?$`)

func pageLang(r *http.Request) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Accept-Language"))), "de") {
		return "de"
	}
	return "en"
}

// appLink is the link the app opens (waylight://paste); the page offers it
// because the QR code carries the address of this page instead.
func appLink(r *http.Request, code string) template.URL {
	if !hostForm.MatchString(r.Host) {
		return ""
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	q := url.Values{"server": {scheme + "://" + r.Host}, "code": {code}}
	// html/template rejects unknown schemes; the query is fully escaped and the prefix is fixed.
	return template.URL("waylight://paste?" + q.Encode()) //nolint:gosec // see above
}

func (h *PasteHandlers) render(w http.ResponseWriter, r *http.Request, status int, d pageData, code string) {
	t := pageTexts[pageLang(r)]
	d.Lang, d.Title, d.Label, d.Send, d.OpenApp, d.Style = pageLang(r), t.Title, t.Label, t.Send, t.OpenApp, pageStyle
	if d.Form {
		d.AppLink = appLink(r, code)
	}
	hd := w.Header()
	hd.Set("Content-Type", "text/html; charset=utf-8")
	hd.Set("Content-Security-Policy", pageCSP)
	hd.Set("Cache-Control", "no-store")
	hd.Set("Referrer-Policy", "no-referrer")
	hd.Set("X-Robots-Tag", "noindex")
	w.WriteHeader(status)
	// A failed write means the phone went away; nothing to report.
	_ = pageTemplate.Execute(w, d)
}

// failure maps an error of the paste use cases to a status and a text.
func failure(t pageText, err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrPasteValue):
		return http.StatusBadRequest, t.Value
	case errors.Is(err, kernel.ErrTooManyRequests):
		return http.StatusTooManyRequests, t.Many
	default:
		return http.StatusNotFound, t.Expired
	}
}

// Page handles GET /paste/{code}: the page the QR code opens in the browser
// of a phone, with one field for the key.
func (h *PasteHandlers) Page(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := h.uc.Check(clientKey(r), code); err != nil {
		status, msg := failure(pageTexts[pageLang(r)], err)
		h.render(w, r, status, pageData{Message: msg, Failed: true}, code)
		return
	}
	h.render(w, r, http.StatusOK, pageData{Form: true}, code)
}

// maxPageForm bounds the form: the key plus its field name.
const maxPageForm = domain.MaxPasteValue*4 + 64

// Fill handles POST /paste/{code}: the key from the page's form. The key is
// never logged or echoed.
func (h *PasteHandlers) Fill(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	t := pageTexts[pageLang(r)]
	r.Body = http.MaxBytesReader(w, r.Body, maxPageForm)
	if err := r.ParseForm(); err != nil {
		h.render(w, r, http.StatusBadRequest, pageData{Message: t.Value, Failed: true, Form: true}, code)
		return
	}
	if err := h.uc.Fill(clientKey(r), code, r.PostForm.Get("value")); err != nil {
		status, msg := failure(t, err)
		h.render(w, r, status, pageData{Message: msg, Failed: true, Form: status == http.StatusBadRequest}, code)
		return
	}
	h.render(w, r, http.StatusOK, pageData{Message: t.Done}, code)
}

func clientKey(r *http.Request) string { return httpapi.ClientIP(r.Context()).String() }
