// Package usecase provisions members and groups from a directory over SCIM
// 2.0 (#396). Members and groups live in the auth module; this package only
// adds what SCIM needs on top: the directory's user names and external IDs.
//
// Ring: Use cases.
package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"slices"
	"strings"

	"trisse-home/ee/scim/domain"
	authdomain "trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/modules/auth/usecase"
	"trisse-home/internal/shared/kernel"
)

// Accounts is what the auth module offers for provisioning (output port).
type Accounts interface {
	CreateMember(ctx context.Context, name string) (string, error)
	SetDisabled(ctx context.Context, accountID string, disabled bool) error
	Member(ctx context.Context, accountID string) (usecase.ProvisionedMember, error)
	Members(ctx context.Context) ([]usecase.ProvisionedMember, error)
	DeleteMember(ctx context.Context, accountID string) error
	Groups(ctx context.Context) ([]authdomain.Group, error)
	SaveGroup(ctx context.Context, id, name string, accounts []string) (authdomain.Group, error)
	DeleteGroup(ctx context.Context, id string) error
}

// UserRecord is what the directory said about a user that the auth module
// does not keep: the user name it uses (an account name is shorter and
// lowercase), its own ID for the user, and a display name.
type UserRecord struct {
	AccountID   string
	UserName    string
	ExternalID  string
	DisplayName string
}

// Store keeps the records (output port). Get returns ok = false when the
// directory never named the account.
type Store interface {
	User(ctx context.Context, accountID string) (UserRecord, bool, error)
	SaveUser(ctx context.Context, r UserRecord) error
	DeleteUser(ctx context.Context, accountID string) error
	GroupExternalID(ctx context.Context, groupID string) (string, error)
	SaveGroupExternalID(ctx context.Context, groupID, externalID string) error
	DeleteGroup(ctx context.Context, groupID string) error
}

// User is a SCIM user: a member account.
type User struct {
	ID          string
	UserName    string
	ExternalID  string
	DisplayName string
	Active      bool
	Groups      []Ref
}

// Group is a SCIM group: a group of the auth module.
type Group struct {
	ID          string
	DisplayName string
	ExternalID  string
	Members     []Ref
}

// Ref points at a user or group, with a name to show.
type Ref struct {
	ID      string
	Display string
}

// PatchOp is one operation of a PATCH request (RFC 7644, 3.5.2).
type PatchOp struct {
	Op    string // add, remove, replace (any case)
	Path  string
	Value any
}

// Errors with the SCIM status they map to.
var (
	ErrUnauthorized = kernel.NewPublicError(kernel.ErrUnauthorized, "The SCIM token is missing or wrong.")
	ErrUserNameUsed = kernel.NewPublicError(kernel.ErrConflict, "A user with this userName exists.")
	ErrInvalidPatch = kernel.NewPublicError(kernel.ErrInvalid, "This PATCH operation is not supported.")
	ErrNoUserName   = kernel.NewPublicError(kernel.ErrInvalid, "userName is required.")
)

// Deps are the ports. Token is SCIM_TOKEN; empty disables SCIM.
type Deps struct {
	Accounts Accounts
	Store    Store
	Token    string
}

// Service implements the SCIM operations.
type Service struct{ d Deps }

// New returns the service.
func New(d Deps) *Service { return &Service{d: d} }

// Authorized checks the bearer token in constant time.
func (s *Service) Authorized(token string) bool {
	return s.d.Token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.d.Token)) == 1
}

func (s *Service) user(ctx context.Context, m usecase.ProvisionedMember, names map[string]string) (User, error) {
	rec, ok, err := s.d.Store.User(ctx, m.ID)
	if err != nil {
		return User{}, err
	}
	u := User{ID: m.ID, UserName: m.Name, Active: !m.Disabled, Groups: []Ref{}}
	if ok {
		u.UserName, u.ExternalID, u.DisplayName = rec.UserName, rec.ExternalID, rec.DisplayName
	}
	for _, g := range m.Groups {
		u.Groups = append(u.Groups, Ref{ID: g, Display: names[g]})
	}
	return u, nil
}

func (s *Service) groupNames(ctx context.Context) (map[string]string, error) {
	groups, err := s.d.Accounts.Groups(ctx)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	for _, g := range groups {
		names[g.ID()] = g.Name()
	}
	return names, nil
}

// GetUser returns a member.
func (s *Service) GetUser(ctx context.Context, id string) (User, error) {
	m, err := s.d.Accounts.Member(ctx, id)
	if err != nil {
		return User{}, err
	}
	names, err := s.groupNames(ctx)
	if err != nil {
		return User{}, err
	}
	return s.user(ctx, m, names)
}

// ListUsers returns the members matching a filter on userName or externalId.
func (s *Service) ListUsers(ctx context.Context, f *domain.Filter) ([]User, error) {
	members, err := s.d.Accounts.Members(ctx)
	if err != nil {
		return nil, err
	}
	names, err := s.groupNames(ctx)
	if err != nil {
		return nil, err
	}
	out := []User{}
	for _, m := range members {
		u, err := s.user(ctx, m, names)
		if err != nil {
			return nil, err
		}
		if f.Matches("username", u.UserName) || f.Matches("externalid", u.ExternalID) {
			out = append(out, u)
		}
	}
	return out, nil
}

// CreateUser adds a member for a directory user.
func (s *Service) CreateUser(ctx context.Context, in User) (User, error) {
	if strings.TrimSpace(in.UserName) == "" {
		return User{}, ErrNoUserName
	}
	existing, err := s.ListUsers(ctx, &domain.Filter{Attribute: "username", Value: in.UserName})
	if err != nil {
		return User{}, err
	}
	if len(existing) > 0 {
		return User{}, ErrUserNameUsed
	}
	id, err := s.d.Accounts.CreateMember(ctx, in.UserName)
	if err != nil {
		return User{}, err
	}
	if err := s.d.Store.SaveUser(ctx, UserRecord{AccountID: id, UserName: in.UserName, ExternalID: in.ExternalID, DisplayName: in.DisplayName}); err != nil {
		return User{}, err
	}
	if !in.Active {
		if err := s.d.Accounts.SetDisabled(ctx, id, true); err != nil {
			return User{}, err
		}
	}
	return s.GetUser(ctx, id)
}

// ReplaceUser takes over the directory's view of a user (PUT). The account
// name stays as created; userName is kept for the directory.
func (s *Service) ReplaceUser(ctx context.Context, id string, in User) (User, error) {
	if strings.TrimSpace(in.UserName) == "" {
		return User{}, ErrNoUserName
	}
	if _, err := s.d.Accounts.Member(ctx, id); err != nil {
		return User{}, err
	}
	if err := s.d.Store.SaveUser(ctx, UserRecord{AccountID: id, UserName: in.UserName, ExternalID: in.ExternalID, DisplayName: in.DisplayName}); err != nil {
		return User{}, err
	}
	if err := s.d.Accounts.SetDisabled(ctx, id, !in.Active); err != nil {
		return User{}, err
	}
	return s.GetUser(ctx, id)
}

// PatchUser applies replace/add operations on active, userName, displayName
// and externalId, with or without a path.
func (s *Service) PatchUser(ctx context.Context, id string, ops []PatchOp) (User, error) {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return User{}, err
	}
	set := func(attr string, v any) error {
		switch strings.ToLower(attr) {
		case "active":
			b, ok := domain.ParseBool(v)
			if !ok {
				return ErrInvalidPatch
			}
			u.Active = b
		case "username":
			u.UserName = fmt.Sprint(v)
		case "displayname":
			u.DisplayName = fmt.Sprint(v)
		case "externalid":
			u.ExternalID = fmt.Sprint(v)
		default:
			// Attributes trisse-home does not keep (name, emails, title …) are
			// accepted and ignored, so a provider's full update does not fail.
		}
		return nil
	}
	for _, op := range ops {
		if strings.EqualFold(op.Op, "remove") {
			continue
		}
		if op.Path != "" {
			if err := set(op.Path, op.Value); err != nil {
				return User{}, err
			}
			continue
		}
		values, ok := op.Value.(map[string]any)
		if !ok {
			return User{}, ErrInvalidPatch
		}
		for k, v := range values {
			if err := set(k, v); err != nil {
				return User{}, err
			}
		}
	}
	return s.ReplaceUser(ctx, id, u)
}

// DeleteUser removes the member like the settings do; their content is
// handled as for any deleted member.
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if err := s.d.Accounts.DeleteMember(ctx, id); err != nil {
		return err
	}
	return s.d.Store.DeleteUser(ctx, id)
}

func (s *Service) group(ctx context.Context, g authdomain.Group) (Group, error) {
	ext, err := s.d.Store.GroupExternalID(ctx, g.ID())
	if err != nil {
		return Group{}, err
	}
	out := Group{ID: g.ID(), DisplayName: g.Name(), ExternalID: ext, Members: []Ref{}}
	for _, a := range g.Accounts() {
		display := a
		if m, err := s.d.Accounts.Member(ctx, a); err == nil {
			display = m.Name
		}
		out.Members = append(out.Members, Ref{ID: a, Display: display})
	}
	return out, nil
}

func (s *Service) find(ctx context.Context, id string) (authdomain.Group, error) {
	groups, err := s.d.Accounts.Groups(ctx)
	if err != nil {
		return authdomain.Group{}, err
	}
	for _, g := range groups {
		if g.ID() == id {
			return g, nil
		}
	}
	return authdomain.Group{}, authdomain.ErrGroupNotFound
}

// GetGroup returns a group.
func (s *Service) GetGroup(ctx context.Context, id string) (Group, error) {
	g, err := s.find(ctx, id)
	if err != nil {
		return Group{}, err
	}
	return s.group(ctx, g)
}

// ListGroups returns the groups matching a filter on displayName or externalId.
func (s *Service) ListGroups(ctx context.Context, f *domain.Filter) ([]Group, error) {
	groups, err := s.d.Accounts.Groups(ctx)
	if err != nil {
		return nil, err
	}
	out := []Group{}
	for _, g := range groups {
		sg, err := s.group(ctx, g)
		if err != nil {
			return nil, err
		}
		if f.Matches("displayname", sg.DisplayName) || f.Matches("externalid", sg.ExternalID) {
			out = append(out, sg)
		}
	}
	return out, nil
}

func refIDs(refs []Ref) []string {
	ids := make([]string, 0, len(refs))
	for _, r := range refs {
		if !slices.Contains(ids, r.ID) {
			ids = append(ids, r.ID)
		}
	}
	return ids
}

// CreateGroup adds a group.
func (s *Service) CreateGroup(ctx context.Context, in Group) (Group, error) {
	g, err := s.d.Accounts.SaveGroup(ctx, "", in.DisplayName, refIDs(in.Members))
	if err != nil {
		return Group{}, err
	}
	if err := s.d.Store.SaveGroupExternalID(ctx, g.ID(), in.ExternalID); err != nil {
		return Group{}, err
	}
	return s.group(ctx, g)
}

// ReplaceGroup sets name and members (PUT).
func (s *Service) ReplaceGroup(ctx context.Context, id string, in Group) (Group, error) {
	g, err := s.d.Accounts.SaveGroup(ctx, id, in.DisplayName, refIDs(in.Members))
	if err != nil {
		return Group{}, err
	}
	if err := s.d.Store.SaveGroupExternalID(ctx, g.ID(), in.ExternalID); err != nil {
		return Group{}, err
	}
	return s.group(ctx, g)
}

// memberRefs reads a PATCH value with members: a list of {"value": id} or one.
func memberRefs(v any) ([]Ref, error) {
	var items []any
	switch x := v.(type) {
	case []any:
		items = x
	case map[string]any:
		items = []any{x}
	default:
		return nil, ErrInvalidPatch
	}
	var out []Ref
	for _, it := range items {
		m, ok := it.(map[string]any)
		id, _ := m["value"].(string)
		if !ok || id == "" {
			return nil, ErrInvalidPatch
		}
		out = append(out, Ref{ID: id})
	}
	return out, nil
}

// PatchGroup adds, removes or replaces members and renames the group.
func (s *Service) PatchGroup(ctx context.Context, id string, ops []PatchOp) (Group, error) {
	g, err := s.GetGroup(ctx, id)
	if err != nil {
		return Group{}, err
	}
	members := refIDs(g.Members)
	for _, op := range ops {
		path := strings.ToLower(strings.TrimSpace(op.Path))
		switch {
		case strings.EqualFold(op.Op, "remove") && strings.HasPrefix(path, "members["):
			who, ok := domain.MemberPath(op.Path)
			if !ok {
				return Group{}, ErrInvalidPatch
			}
			members = slices.DeleteFunc(members, func(m string) bool { return m == who })
		case path == "members":
			refs, err := memberRefs(op.Value)
			if err != nil && !strings.EqualFold(op.Op, "remove") {
				return Group{}, err
			}
			switch strings.ToLower(op.Op) {
			case "add":
				members = append(members, refIDs(refs)...)
			case "remove":
				if refs == nil { // no value: remove all members
					members = nil
				}
				for _, r := range refs {
					members = slices.DeleteFunc(members, func(m string) bool { return m == r.ID })
				}
			case "replace":
				members = refIDs(refs)
			default:
				return Group{}, ErrInvalidPatch
			}
		case path == "displayname" && !strings.EqualFold(op.Op, "remove"):
			g.DisplayName = fmt.Sprint(op.Value)
		case path == "externalid" && !strings.EqualFold(op.Op, "remove"):
			g.ExternalID = fmt.Sprint(op.Value)
		case path == "":
			values, ok := op.Value.(map[string]any)
			if !ok {
				return Group{}, ErrInvalidPatch
			}
			if name, ok := values["displayName"].(string); ok {
				g.DisplayName = name
			}
			if raw, ok := values["members"]; ok {
				refs, err := memberRefs(raw)
				if err != nil {
					return Group{}, err
				}
				members = refIDs(refs)
			}
		default:
			return Group{}, ErrInvalidPatch
		}
	}
	g.Members = nil
	for _, m := range members {
		g.Members = append(g.Members, Ref{ID: m})
	}
	return s.ReplaceGroup(ctx, id, g)
}

// DeleteGroup removes a group; its members stay.
func (s *Service) DeleteGroup(ctx context.Context, id string) error {
	if err := s.d.Accounts.DeleteGroup(ctx, id); err != nil {
		return err
	}
	return s.d.Store.DeleteGroup(ctx, id)
}

// IsNotFound reports errors that SCIM answers with 404.
func IsNotFound(err error) bool {
	return errors.Is(err, authdomain.ErrMemberNotFound) || errors.Is(err, authdomain.ErrGroupNotFound) ||
		errors.Is(err, kernel.ErrNotFound)
}
