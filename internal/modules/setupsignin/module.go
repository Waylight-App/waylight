package setupsignin

import (
	"net/http"
	"net/url"

	"waylight/internal/modules"
	"waylight/internal/modules/setupsignin/adapter/controller"
	"waylight/internal/modules/setupsignin/adapter/gateway"
	"waylight/internal/modules/setupsignin/usecase"
	"waylight/internal/platform/oauth"
	"waylight/internal/shared/kernel"
)

// Module serves the sign-ins and the pastes from the phone of the setup
// assistant and hands the credentials to the modules that are set up.
type Module struct {
	modules.Base
	uc       usecase.SignIns
	handlers *controller.Handlers
	paste    *controller.PasteHandlers
}

// New builds the module. client must follow the integration URL rules;
// publicURL (PUBLIC_URL, may be nil) is Home Assistant's client ID and
// without it only Nextcloud can be signed in to.
func New(client *http.Client, publicURL *url.URL, clock kernel.Clock) *Module {
	uc := usecase.New(usecase.Deps{
		Nextcloud:     gateway.NewNextcloudLogin(client),
		HomeAssistant: newHomeAssistant(client, publicURL, clock.Now),
		Clock:         clock,
		Token:         oauth.NewVerifier,
	})
	pastes := usecase.NewPastes(usecase.PasteDeps{Clock: clock, Token: oauth.NewVerifier})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), paste: controller.NewPasteHandlers(pastes)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "setupsignin" }

// DataKind says the module stores nothing; sign-ins live in memory for
// minutes and belong to the account that started them.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// SignIns is what the modules that are set up redeem a sign-in with.
func (m *Module) SignIns() kernel.SignIns { return m.uc }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("POST /api/setup/signin", m.handlers.Start)
	r.Protected("GET /api/setup/signin/{id}", m.handlers.Status)
	r.Protected("DELETE /api/setup/signin/{id}", m.handlers.Cancel)
	r.Protected("POST /api/setup/paste", m.paste.Create)
	r.Protected("PUT /api/setup/paste/{code}", m.paste.Submit)
	r.Protected("GET /api/setup/paste/{code}", m.paste.Take)
	r.Protected("DELETE /api/setup/paste/{code}", m.paste.Cancel)
	r.Public("GET "+CallbackPath, m.handlers.Callback)
}
