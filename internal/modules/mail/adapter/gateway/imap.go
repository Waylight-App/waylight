package gateway

import (
	"bytes"
	"context"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"waylight/internal/modules/mail/domain"
	"waylight/internal/shared/imap"
)

// IMAP implements usecase.Reader with the shared read-only IMAP client:
// EXAMINE, UID SEARCH UNSEEN and headers fetched with BODY.PEEK, so nothing
// is marked as read and no body is ever downloaded.
type IMAP struct{ client imap.Client }

// NewIMAP returns the reader.
func NewIMAP(dial imap.DialFunc) *IMAP { return &IMAP{client: imap.Client{Dial: dial}} }

func account(a domain.Account) imap.Account {
	return imap.Account{Host: a.Host, Port: a.Port, Username: a.Username, Password: a.Password}
}

// Test implements usecase.Reader.
func (m *IMAP) Test(ctx context.Context, a domain.Account) error {
	c, err := m.client.Open(ctx, account(a))
	if err != nil {
		return err
	}
	defer c.Close()
	_, _, err = c.Examine(a.Folder)
	return err
}

var (
	fetchLine    = regexp.MustCompile(`^\* \d+ FETCH `)
	headerItem   = regexp.MustCompile(`BODY\[(HEADER)[^\]]*\] `)
	internalDate = regexp.MustCompile(`INTERNALDATE "([^"]+)"`)
)

// Unread implements usecase.Reader.
func (m *IMAP) Unread(ctx context.Context, a domain.Account, newest int) (domain.Inbox, error) {
	c, err := m.client.Open(ctx, account(a))
	if err != nil {
		return domain.Inbox{}, err
	}
	defer c.Close()
	if _, _, err := c.Examine(a.Folder); err != nil {
		return domain.Inbox{}, err
	}
	uids, err := c.SearchUIDs("UNSEEN")
	if err != nil {
		return domain.Inbox{}, err
	}
	in := domain.Inbox{Unread: len(uids)}
	if len(uids) == 0 {
		return in, nil
	}
	if len(uids) > newest {
		uids = uids[len(uids)-newest:]
	}
	set := make([]string, len(uids))
	for i, u := range uids {
		set[i] = strconv.FormatUint(uint64(u), 10)
	}
	lines, err := c.Cmd("UID FETCH " + strings.Join(set, ",") + " (UID INTERNALDATE BODY.PEEK[HEADER.FIELDS (FROM SUBJECT DATE)])")
	if err != nil {
		return domain.Inbox{}, err
	}
	for _, l := range lines {
		if !fetchLine.MatchString(l.Text) {
			continue
		}
		header := l.Section(headerItem)["HEADER"]
		if header == nil {
			continue
		}
		msg := parseHeader(header)
		if m := internalDate.FindStringSubmatch(l.Text); m != nil {
			if t, err := time.Parse("2-Jan-2006 15:04:05 -0700", strings.TrimSpace(m[1])); err == nil {
				msg.Date = t
			}
		}
		in.Messages = append(in.Messages, msg)
	}
	return in, nil
}

func parseHeader(raw []byte) domain.Message {
	msg, err := mail.ReadMessage(bytes.NewReader(append(raw, '\r', '\n')))
	if err != nil {
		return domain.Message{}
	}
	from, _ := imap.Words.DecodeHeader(msg.Header.Get("From"))
	subject, _ := imap.Words.DecodeHeader(msg.Header.Get("Subject"))
	date, _ := msg.Header.Date()
	return domain.Message{From: strings.TrimSpace(from), Subject: strings.TrimSpace(subject), Date: date}
}
