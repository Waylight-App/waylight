// Package controller serves SCIM 2.0 under /scim/v2 (#396). Requests come
// from the identity provider with a bearer token, not from a browser session.
//
// Ring: Interface adapters (controller).
package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"waylight/ee/scim/domain"
	"waylight/ee/scim/usecase"
	"waylight/internal/shared/kernel"
)

const (
	schemaUser  = "urn:ietf:params:scim:schemas:core:2.0:User"
	schemaGroup = "urn:ietf:params:scim:schemas:core:2.0:Group"
	schemaList  = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	schemaError = "urn:ietf:params:scim:api:messages:2.0:Error"
	schemaSPC   = "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"
	contentType = "application/scim+json"
	maxBody     = 1 << 20
)

// Handlers serves SCIM.
type Handlers struct{ s *usecase.Service }

// NewHandlers returns the handlers.
func NewHandlers(s *usecase.Service) *Handlers { return &Handlers{s: s} }

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, scimType, detail string) {
	body := map[string]any{"schemas": []string{schemaError}, "status": strconv.Itoa(status), "detail": detail}
	if scimType != "" {
		body["scimType"] = scimType
	}
	write(w, status, body)
}

func fail(w http.ResponseWriter, err error) {
	var pe *kernel.PublicError
	detail := "internal error"
	if errors.As(err, &pe) {
		detail = pe.Error()
	}
	switch {
	case usecase.IsNotFound(err):
		writeError(w, http.StatusNotFound, "", detail)
	case errors.Is(err, kernel.ErrConflict):
		writeError(w, http.StatusConflict, "uniqueness", detail)
	case errors.Is(err, domain.ErrUnsupportedFilter):
		writeError(w, http.StatusBadRequest, "invalidFilter", err.Error())
	case errors.Is(err, kernel.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalidValue", detail)
	case errors.Is(err, kernel.ErrForbidden):
		writeError(w, http.StatusForbidden, "", detail)
	default:
		writeError(w, http.StatusInternalServerError, "", detail)
	}
}

// authorized checks the bearer token and answers 401 otherwise.
func (h *Handlers) authorized(w http.ResponseWriter, r *http.Request) bool {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || !h.s.Authorized(strings.TrimSpace(token)) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Waylight SCIM"`)
		writeError(w, http.StatusUnauthorized, "", usecase.ErrUnauthorized.Error())
		return false
	}
	return true
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalidSyntax", "The request body is not valid JSON.")
		return false
	}
	return true
}

// userDTO is the SCIM user resource, with the attributes Waylight keeps.
type userDTO struct {
	Schemas     []string `json:"schemas"`
	ID          string   `json:"id,omitempty"`
	ExternalID  string   `json:"externalId,omitempty"`
	UserName    string   `json:"userName"`
	DisplayName string   `json:"displayName,omitempty"`
	Active      *bool    `json:"active,omitempty"`
	Groups      []refDTO `json:"groups,omitempty"`
	Meta        *metaDTO `json:"meta,omitempty"`
}

type groupDTO struct {
	Schemas     []string `json:"schemas"`
	ID          string   `json:"id,omitempty"`
	ExternalID  string   `json:"externalId,omitempty"`
	DisplayName string   `json:"displayName"`
	Members     []refDTO `json:"members"`
	Meta        *metaDTO `json:"meta,omitempty"`
}

type refDTO struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
}

type metaDTO struct {
	ResourceType string `json:"resourceType"`
	Location     string `json:"location"`
}

type listDTO struct {
	Schemas      []string `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	StartIndex   int      `json:"startIndex"`
	ItemsPerPage int      `json:"itemsPerPage"`
	Resources    []any    `json:"Resources"`
}

type patchDTO struct {
	Operations []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value any    `json:"value"`
	} `json:"Operations"`
}

func toRefs(rs []usecase.Ref) []refDTO {
	out := []refDTO{}
	for _, r := range rs {
		out = append(out, refDTO{Value: r.ID, Display: r.Display})
	}
	return out
}

func fromRefs(rs []refDTO) []usecase.Ref {
	var out []usecase.Ref
	for _, r := range rs {
		out = append(out, usecase.Ref{ID: r.Value})
	}
	return out
}

func userOut(u usecase.User) userDTO {
	active := u.Active
	return userDTO{Schemas: []string{schemaUser}, ID: u.ID, ExternalID: u.ExternalID, UserName: u.UserName,
		DisplayName: u.DisplayName, Active: &active, Groups: toRefs(u.Groups),
		Meta: &metaDTO{ResourceType: "User", Location: "/scim/v2/Users/" + u.ID}}
}

func userIn(d userDTO) usecase.User {
	// A user without "active" is active (RFC 7643, 4.1.1 leaves the default to
	// the provider; every common one means active).
	active := d.Active == nil || *d.Active
	return usecase.User{UserName: d.UserName, ExternalID: d.ExternalID, DisplayName: d.DisplayName, Active: active}
}

func groupOut(g usecase.Group) groupDTO {
	return groupDTO{Schemas: []string{schemaGroup}, ID: g.ID, ExternalID: g.ExternalID, DisplayName: g.DisplayName,
		Members: toRefs(g.Members), Meta: &metaDTO{ResourceType: "Group", Location: "/scim/v2/Groups/" + g.ID}}
}

// page applies startIndex (1-based) and count to a list.
func page(r *http.Request, items []any) listDTO {
	start, _ := strconv.Atoi(r.URL.Query().Get("startIndex"))
	start = max(start, 1)
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil || count < 0 {
		count = len(items)
	}
	from := min(start-1, len(items))
	to := min(from+count, len(items))
	return listDTO{Schemas: []string{schemaList}, TotalResults: len(items), StartIndex: start, ItemsPerPage: to - from,
		Resources: append([]any{}, items[from:to]...)}
}

// ServiceProviderConfig handles GET /scim/v2/ServiceProviderConfig.
func (h *Handlers) ServiceProviderConfig(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	write(w, http.StatusOK, map[string]any{
		"schemas":        []string{schemaSPC},
		"patch":          map[string]bool{"supported": true},
		"bulk":           map[string]any{"supported": false, "maxOperations": 0, "maxPayloadSize": 0},
		"filter":         map[string]any{"supported": true, "maxResults": 200},
		"changePassword": map[string]bool{"supported": false},
		"sort":           map[string]bool{"supported": false},
		"etag":           map[string]bool{"supported": false},
		"authenticationSchemes": []map[string]string{{
			"type": "oauthbearertoken", "name": "Bearer token", "description": "The token set in SCIM_TOKEN",
		}},
	})
}

// ListUsers handles GET /scim/v2/Users.
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	f, err := domain.ParseFilter(r.URL.Query().Get("filter"))
	if err != nil {
		fail(w, err)
		return
	}
	users, err := h.s.ListUsers(r.Context(), f)
	if err != nil {
		fail(w, err)
		return
	}
	items := make([]any, 0, len(users))
	for _, u := range users {
		items = append(items, userOut(u))
	}
	write(w, http.StatusOK, page(r, items))
}

// GetUser handles GET /scim/v2/Users/{id}.
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	u, err := h.s.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, userOut(u))
}

// CreateUser handles POST /scim/v2/Users.
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d userDTO
	if !decode(w, r, &d) {
		return
	}
	u, err := h.s.CreateUser(r.Context(), userIn(d))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusCreated, userOut(u))
}

// ReplaceUser handles PUT /scim/v2/Users/{id}.
func (h *Handlers) ReplaceUser(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d userDTO
	if !decode(w, r, &d) {
		return
	}
	u, err := h.s.ReplaceUser(r.Context(), r.PathValue("id"), userIn(d))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, userOut(u))
}

func ops(d patchDTO) []usecase.PatchOp {
	out := make([]usecase.PatchOp, 0, len(d.Operations))
	for _, o := range d.Operations {
		out = append(out, usecase.PatchOp{Op: o.Op, Path: o.Path, Value: o.Value})
	}
	return out
}

// PatchUser handles PATCH /scim/v2/Users/{id}.
func (h *Handlers) PatchUser(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d patchDTO
	if !decode(w, r, &d) {
		return
	}
	u, err := h.s.PatchUser(r.Context(), r.PathValue("id"), ops(d))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, userOut(u))
}

// DeleteUser handles DELETE /scim/v2/Users/{id}.
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	if err := h.s.DeleteUser(r.Context(), r.PathValue("id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListGroups handles GET /scim/v2/Groups.
func (h *Handlers) ListGroups(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	f, err := domain.ParseFilter(r.URL.Query().Get("filter"))
	if err != nil {
		fail(w, err)
		return
	}
	groups, err := h.s.ListGroups(r.Context(), f)
	if err != nil {
		fail(w, err)
		return
	}
	items := make([]any, 0, len(groups))
	for _, g := range groups {
		items = append(items, groupOut(g))
	}
	write(w, http.StatusOK, page(r, items))
}

// GetGroup handles GET /scim/v2/Groups/{id}.
func (h *Handlers) GetGroup(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	g, err := h.s.GetGroup(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, groupOut(g))
}

// CreateGroup handles POST /scim/v2/Groups.
func (h *Handlers) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d groupDTO
	if !decode(w, r, &d) {
		return
	}
	g, err := h.s.CreateGroup(r.Context(), usecase.Group{DisplayName: d.DisplayName, ExternalID: d.ExternalID, Members: fromRefs(d.Members)})
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusCreated, groupOut(g))
}

// ReplaceGroup handles PUT /scim/v2/Groups/{id}.
func (h *Handlers) ReplaceGroup(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d groupDTO
	if !decode(w, r, &d) {
		return
	}
	g, err := h.s.ReplaceGroup(r.Context(), r.PathValue("id"),
		usecase.Group{DisplayName: d.DisplayName, ExternalID: d.ExternalID, Members: fromRefs(d.Members)})
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, groupOut(g))
}

// PatchGroup handles PATCH /scim/v2/Groups/{id}.
func (h *Handlers) PatchGroup(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	var d patchDTO
	if !decode(w, r, &d) {
		return
	}
	g, err := h.s.PatchGroup(r.Context(), r.PathValue("id"), ops(d))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, http.StatusOK, groupOut(g))
}

// DeleteGroup handles DELETE /scim/v2/Groups/{id}.
func (h *Handlers) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	if err := h.s.DeleteGroup(r.Context(), r.PathValue("id")); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
