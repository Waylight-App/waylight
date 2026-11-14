package caldav

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const cardReport = `<?xml version="1.0" encoding="utf-8"?>
<C:addressbook-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:carddav">
  <D:prop><D:getetag/><C:address-data/></D:prop>
</C:addressbook-query>`

type cardMultistatus struct {
	Responses []struct {
		Href      string `xml:"href"`
		Propstats []struct {
			Status      string `xml:"status"`
			AddressData string `xml:"prop>address-data"`
		} `xml:"propstat"`
	} `xml:"response"`
}

// VCards reads every vCard of a CardDAV address book (#287) with one
// addressbook-query REPORT; a.URL is the address book collection.
func (c Client) VCards(ctx context.Context, a Account) ([]string, error) {
	res, err := c.Do(ctx, a, "REPORT", a.URL, strings.NewReader(cardReport), map[string]string{"Depth": "1", "Content-Type": xmlHeaders})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, errNoCollection
	}
	if res.StatusCode != http.StatusMultiStatus {
		return nil, StatusError(res)
	}
	var ms cardMultistatus
	if err := xml.NewDecoder(io.LimitReader(res.Body, MaxResponseBytes)).Decode(&ms); err != nil {
		return nil, fmt.Errorf("parse CardDAV response: %w", err)
	}
	var out []string
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if (ps.Status == "" || strings.Contains(ps.Status, " 200 ")) && ps.AddressData != "" {
				out = append(out, ps.AddressData)
			}
		}
	}
	return out, nil
}
