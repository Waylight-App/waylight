package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/mail"
	"trisse-home/internal/platform/httpclient"
)

// The mail module registers itself; no shared file lists it (#193).
func init() {
	registerModule("mail", func(d moduleDeps) modules.Module {
		return mail.New(d.db, d.out.secrets, httpclient.IntegrationDialContext, d.ids, d.clock)
	})
}
