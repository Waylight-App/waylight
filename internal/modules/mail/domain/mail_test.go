package domain

import "testing"

func TestNewAccount(t *testing.T) {
	a, err := NewAccount("a", " Privat ", "IMAP.GMX.NET", 0, " me@gmx.net ", "pw", "", "https://www.gmx.net/")
	if err != nil || a.Port != 993 || a.Folder != "INBOX" || a.Host != "imap.gmx.net" || a.Username != "me@gmx.net" || a.Name != "Privat" {
		t.Fatalf("%+v %v", a, err)
	}
	for _, bad := range [][4]string{
		{"", "imap.x.de", "u", "p"},
		{"n", "imap x", "u", "p"},
		{"n", "imap.x.de", "", "p"},
		{"n", "imap.x.de", "u", ""},
		{"n", "imap.x.de", "u\r\nA1 LOGOUT", "p"},
	} {
		if _, err := NewAccount("a", bad[0], bad[1], 993, bad[2], bad[3], "", ""); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if _, err := NewAccount("a", "n", "imap.x.de", 993, "u", "p", "", "javascript:alert(1)"); err == nil {
		t.Error("webmail must be http(s)")
	}
}

func TestSenderName(t *testing.T) {
	for in, want := range map[string]string{`"DHL Paket" <noreply@dhl.de>`: "DHL Paket", "<a@b.de>": "a@b.de", "a@b.de": "a@b.de", "Anna Muster <anna@x.de>": "Anna Muster"} {
		if got := SenderName(in); got != want {
			t.Errorf("%q: %q", in, got)
		}
	}
}
