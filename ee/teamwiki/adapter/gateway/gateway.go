// Package gateway asks Confluence, Microsoft Graph and Notion on behalf of a
// member (#327–#329). Each request carries the member's own token, so the
// service returns only what that person may see.
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const maxResponse = 2 << 20

// call sends a JSON request with a bearer token and decodes the answer.
func call(ctx context.Context, client *http.Client, method, url, token string, body any, headers map[string]string, dst any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", req.URL.Host, res.Status)
	}
	return json.NewDecoder(io.LimitReader(res.Body, maxResponse)).Decode(dst)
}

// excerpt removes highlight markup some services put into snippets and
// shortens the text.
func excerpt(s string) string {
	for _, tag := range []string{"@@@hl@@@", "@@@endhl@@@", "<c0>", "</c0>", "<ddd/>"} {
		s = strings.ReplaceAll(s, tag, "")
	}
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > 200 {
		s = string([]rune(s)[:199]) + "…"
	}
	return s
}

func parseTime(v string) time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}
	}
	return t
}
