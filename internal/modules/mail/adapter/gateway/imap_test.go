package gateway

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/mail/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/imap"
)

var writes = regexp.MustCompile(`(?i)^(?:SELECT|STORE|UID STORE|COPY|UID COPY|MOVE|UID MOVE|EXPUNGE|DELETE|APPEND)\b`)

// fakeServer answers like Dovecot: EXAMINE, UID SEARCH UNSEEN, UID FETCH of
// headers with INTERNALDATE; any command that could change the mailbox
// fails the test.
func fakeServer(t *testing.T, unseen []int) (string, int) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	header := func(uid int) string {
		return fmt.Sprintf("From: =?UTF-8?Q?J=C3=BCrgen?= <j@x.de>\r\nSubject: Nachricht %d\r\nDate: Sat, 14 Nov 2026 08:0%d:00 +0100\r\n\r\n", uid, uid%10)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				r := bufio.NewReader(c)
				fmt.Fprint(c, "* OK ready\r\n")
				for {
					line, err := r.ReadString('\n')
					if err != nil {
						return
					}
					tag, cmd, _ := strings.Cut(strings.TrimRight(line, "\r\n"), " ")
					switch {
					case writes.MatchString(cmd):
						t.Errorf("write command: %s", cmd)
						fmt.Fprintf(c, "%s NO\r\n", tag)
					case strings.HasPrefix(cmd, "LOGIN"):
						if !strings.Contains(cmd, `"pw"`) {
							fmt.Fprintf(c, "%s NO [AUTHENTICATIONFAILED] wrong\r\n", tag)
							continue
						}
						fmt.Fprintf(c, "%s OK\r\n", tag)
					case strings.HasPrefix(cmd, "EXAMINE"):
						fmt.Fprintf(c, "* OK [UIDVALIDITY 3]\r\n* OK [UIDNEXT 50]\r\n%s OK [READ-ONLY]\r\n", tag)
					case cmd == "UID SEARCH UNSEEN":
						s := make([]string, len(unseen))
						for i, u := range unseen {
							s[i] = fmt.Sprint(u)
						}
						fmt.Fprintf(c, "* SEARCH %s\r\n%s OK\r\n", strings.Join(s, " "), tag)
					case strings.HasPrefix(cmd, "UID FETCH"):
						if !strings.Contains(cmd, "BODY.PEEK[HEADER.FIELDS") || strings.Contains(cmd, "TEXT") {
							t.Errorf("must fetch headers only, with PEEK: %s", cmd)
						}
						set := strings.Fields(cmd)[2]
						for i, u := range strings.Split(set, ",") {
							var uid int
							fmt.Sscan(u, &uid)
							h := header(uid)
							fmt.Fprintf(c, "* %d FETCH (UID %d INTERNALDATE \"14-Nov-2026 0%d:00:00 +0000\" BODY[HEADER.FIELDS (FROM SUBJECT DATE)] {%d}\r\n%s)\r\n", i+1, uid, uid%10, len(h), h)
						}
						fmt.Fprintf(c, "%s OK\r\n", tag)
					case cmd == "LOGOUT":
						fmt.Fprintf(c, "* BYE\r\n%s OK\r\n", tag)
						return
					default:
						fmt.Fprintf(c, "%s BAD\r\n", tag)
					}
				}
			}(c)
		}
	}()
	addr := ln.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port
}

func TestIMAPUnread(t *testing.T) {
	host, port := fakeServer(t, []int{3, 5, 8})
	m := &IMAP{client: imap.Client{Dial: (&net.Dialer{}).DialContext, Plain: true}}
	a := domain.Account{Host: host, Port: port, Username: "u", Password: "pw", Folder: "INBOX"}
	ctx := context.Background()
	if err := m.Test(ctx, a); err != nil {
		t.Fatal(err)
	}
	in, err := m.Unread(ctx, a, 2)
	if err != nil || in.Unread != 3 || len(in.Messages) != 2 {
		t.Fatalf("count all, list the newest two: %+v %v", in, err)
	}
	if in.Messages[1].From != "Jürgen <j@x.de>" || in.Messages[1].Subject != "Nachricht 8" || !in.Messages[1].Date.Equal(time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)) {
		t.Fatalf("decoded header and INTERNALDATE: %+v", in.Messages[1])
	}
	a.Password = "nope"
	if err := m.Test(ctx, a); err == nil {
		t.Fatal("wrong password accepted")
	}
	empty := &IMAP{client: m.client}
	host2, port2 := fakeServer(t, nil)
	if in, err := empty.Unread(ctx, domain.Account{Host: host2, Port: port2, Username: "u", Password: "pw", Folder: "INBOX"}, 10); err != nil || in.Unread != 0 {
		t.Fatal(in, err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRepository(t *testing.T) {
	ctx := context.Background()
	r := NewRepository(dbtest.Open(t, "mail", Migrations()), prefixCipher{})
	a := domain.Account{ID: "a", Name: "Privat", Host: "imap.x.de", Port: 993, Username: "u", Password: "geheim", Folder: "INBOX"}
	if err := r.Save(ctx, a); err != nil {
		t.Fatal(err)
	}
	var enc string
	_ = r.db.QueryRowContext(ctx, `SELECT password_encrypted FROM mail_accounts`).Scan(&enc)
	if enc != "enc:geheim" {
		t.Fatal(enc)
	}
	if got, err := r.Get(ctx, "a"); err != nil || got != a {
		t.Fatalf("%+v %v", got, err)
	}
	_ = r.Delete(ctx, "a")
	if _, err := r.Get(ctx, "a"); err == nil {
		t.Fatal("deleted")
	}
}
