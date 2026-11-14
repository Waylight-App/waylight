package gateway

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/shared/imap"
)

// IMAP reads shipping mails from one folder (#166). It is deliberately
// minimal and read-only: the folder is opened with EXAMINE (the server
// refuses any change), bodies are fetched with BODY.PEEK (not even the
// \Seen flag changes), and only headers and the first part of the text are
// read; attachments are ignored. Implicit TLS (port 993) only.
type IMAP struct {
	client imap.Client
}

// DialFunc opens a TCP connection; production passes the guarded dialer of
// the integration clients, so link-local and metadata addresses are refused.
type DialFunc = imap.DialFunc

// NewIMAP returns the mail source.
func NewIMAP(dial DialFunc) *IMAP { return &IMAP{client: imap.Client{Dial: dial}} }

// Limits of one run.
const (
	maxMailsPerRun = 50
	maxTextBytes   = 256 << 10
)

func account(s domain.MailSettings) imap.Account {
	return imap.Account{Host: s.Host, Port: s.Port, Username: s.Username, Password: s.Password}
}

// Test implements usecase.MailSource.
func (m *IMAP) Test(ctx context.Context, s domain.MailSettings) error {
	c, err := m.client.Open(ctx, account(s))
	if err != nil {
		return err
	}
	defer c.Close()
	_, _, err = c.Examine(s.Folder)
	return err
}

// Fetch implements usecase.MailSource.
func (m *IMAP) Fetch(ctx context.Context, s domain.MailSettings, st usecase.MailState, since time.Time) ([]domain.Mail, usecase.MailState, error) {
	c, err := m.client.Open(ctx, account(s))
	if err != nil {
		return nil, st, err
	}
	defer c.Close()
	validity, uidNext, err := c.Examine(s.Folder)
	if err != nil {
		return nil, st, err
	}
	next := usecase.MailState{UIDValidity: validity, LastUID: st.LastUID}
	var criteria string
	if st.UIDValidity != validity || st.LastUID == 0 {
		next.LastUID = 0
		criteria = "SINCE " + since.Format("2-Jan-2006")
	} else {
		criteria = fmt.Sprintf("UID %d:*", st.LastUID+1)
	}
	found, err := c.SearchUIDs(criteria)
	if err != nil {
		return nil, st, err
	}
	var uids []uint32
	for _, u := range found {
		// "n:*" always returns the last message, even an old one.
		if u > next.LastUID {
			uids = append(uids, u)
		}
	}
	if len(uids) > maxMailsPerRun {
		uids = uids[len(uids)-maxMailsPerRun:] // the newest
	}
	if uidNext > 0 && uidNext-1 > next.LastUID {
		next.LastUID = uidNext - 1
	}
	if len(uids) == 0 {
		return nil, next, nil
	}
	next.LastUID = max(next.LastUID, uids[len(uids)-1])
	set := make([]string, len(uids))
	for i, u := range uids {
		set[i] = strconv.FormatUint(uint64(u), 10)
	}
	lines, err := c.Cmd("UID FETCH " + strings.Join(set, ",") +
		" (UID BODY.PEEK[HEADER.FIELDS (FROM SUBJECT CONTENT-TYPE CONTENT-TRANSFER-ENCODING)] BODY.PEEK[TEXT]<0." + strconv.Itoa(maxTextBytes) + ">)")
	if err != nil {
		return nil, st, err
	}
	var mails []domain.Mail
	for _, l := range lines {
		if !fetchLine.MatchString(l.Text) {
			continue
		}
		items := l.Section(sectionRe)
		header, text := items["HEADER"], items["TEXT"]
		if header == nil {
			continue
		}
		mails = append(mails, ParseMail(header, text))
	}
	return mails, next, nil
}

var (
	fetchLine = regexp.MustCompile(`^\* \d+ FETCH `)
	sectionRe = regexp.MustCompile(`BODY\[(HEADER|TEXT)[^\]]*\](?:<\d+>)? `)
)
