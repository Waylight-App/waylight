package gateway

import (
	"context"
	"net/http"

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/shared/caldav"
)

// CardDAV implements usecase.AddressBook with the shared CardDAV client.
type CardDAV struct{ client caldav.Client }

// NewCardDAV returns the address book reader.
func NewCardDAV(client *http.Client) *CardDAV { return &CardDAV{client: caldav.Client{HTTP: client}} }

// Cards implements usecase.AddressBook.
func (c *CardDAV) Cards(ctx context.Context, url, username, password string) ([]domain.Card, error) {
	vcards, err := c.client.VCards(ctx, caldav.Account{URL: url, Username: username, Password: password})
	if err != nil {
		return nil, err
	}
	var out []domain.Card
	for _, v := range vcards {
		out = append(out, domain.ParseVCards(v)...)
	}
	return out, nil
}
