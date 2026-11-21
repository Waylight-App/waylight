package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"waylight/internal/modules/calendar/domain"
	"waylight/internal/shared/kernel"
)

// Bounds for the number of days requested by the widget.
const (
	DefaultDays = 7
	MaxDays     = 31
)

// SettingsView never contains the password.
type SettingsView struct {
	Configured   bool
	CalendarURLs []string
	Username     string
	HasPassword  bool
}

// SettingsInput is the input of SaveSettings; an empty password keeps the stored one.
type SettingsInput struct {
	CalendarURLs []string
	Username     string
	Password     string
}

// ConnectInput is what the setup assistant asks for (#588): the server
// address instead of each calendar's URL.
type ConnectInput struct {
	Server, Username, Password string
}

// ConnectView is the outcome of Connect: the saved settings, the calendars
// found and the next few events for the preview.
type ConnectView struct {
	Settings  SettingsView
	Calendars []domain.FoundCalendar
	Next      []EventView
}

// connectPreview is how many upcoming events the setup preview shows.
const connectPreview = 3

// EventView is an event at the use case boundary.
type EventView struct {
	UID, Summary, Location, Calendar, Color string
	// Description (plain text) and URL (http(s) only) for the detail view (#163).
	Description, URL string
	Start, End       time.Time
	AllDay           bool
}

// AgendaResponse lists upcoming events.
type AgendaResponse struct {
	Events    []EventView
	UpdatedAt time.Time
	Error     string
}

// Calendar is the input port of the module.
type Calendar interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Connect finds the event calendars of an account from the server
	// address, checks them and saves them, for the setup assistant (#588).
	Connect(ctx context.Context, in ConnectInput) (ConnectView, error)
	Upcoming(ctx context.Context, days int) (AgendaResponse, error)
	// Range returns the events overlapping [from, to) (at most 42 days).
	Range(ctx context.Context, from, to time.Time) (AgendaResponse, error)
	// CreateEvent writes a new event to one of the configured calendars.
	CreateEvent(ctx context.Context, in NewEventInput) (EventView, error)
	Refresh(ctx context.Context) error
}

// NewEventInput describes an event to create. For all-day events Start and
// End are dates and End is inclusive (a one-day event has End == Start).
type NewEventInput struct {
	CalendarURL, Summary, Location string
	AllDay                         bool
	Start, End                     time.Time
}

// EventWriter creates events in a calendar (output port).
type EventWriter interface {
	Create(ctx context.Context, s domain.Settings, calendarURL string, e domain.Event) error
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// CalendarFinder lists the event calendars of a CalDAV account from the
// server address (output port, #588).
type CalendarFinder interface {
	Discover(ctx context.Context, server, username, password string) ([]domain.FoundCalendar, error)
}

// EventSource loads events of one CalDAV calendar in a time range (output port).
type EventSource interface {
	Events(ctx context.Context, s domain.Settings, calendarURL string, from, to time.Time) ([]domain.Event, error)
}

// FeedSource loads the events of an ICS subscription in a time range,
// recurrences expanded (output port, #98). The adapter decides when to
// download the feed again (refresh interval, conditional requests).
type FeedSource interface {
	Events(ctx context.Context, s domain.Subscription, from, to time.Time) ([]domain.Event, error)
}

// SubscriptionRepository persists ICS subscriptions (output port).
type SubscriptionRepository interface {
	List(ctx context.Context) ([]domain.Subscription, error)
	Get(ctx context.Context, id string) (domain.Subscription, error)
	Save(ctx context.Context, s domain.Subscription) error
	Delete(ctx context.Context, id string) error
}

// SharedAccount is a connected account another person shared, with the
// owner whose tokens it uses.
type SharedAccount struct {
	Account domain.Account
	OwnerID string
}

// SharedAccounts lists the connected accounts others shared with the
// principal (output port, team calendar #299).
type SharedAccounts interface {
	ListSharedAccounts(ctx context.Context) ([]SharedAccount, error)
}

// SharedSubscriptions lists the subscriptions other accounts shared with
// the principal (output port, team calendar #299).
type SharedSubscriptions interface {
	ListShared(ctx context.Context) ([]domain.Subscription, error)
}

// Deps bundles the output ports of the calendar module.
type Deps struct {
	// Shared and SharingOn add calendars others shared to a person's agenda
	// while SharingOn says so (the ee/ team calendar); nil adds none.
	Shared         SharedSubscriptions
	SharedAccounts SharedAccounts
	SharingOn      func(ctx context.Context) bool
	Repo           SettingsRepository
	Source         EventSource
	Writer         EventWriter
	// Finder lets the setup assistant connect by server address; nil
	// leaves only the form with calendar addresses.
	Finder        CalendarFinder
	Subscriptions SubscriptionRepository
	Feeds         FeedSource
	IDs           kernel.IDGenerator
	Clock         kernel.Clock
	// Accounts, API and OAuth connect Google and Microsoft calendars
	// (#282); without them no account can be connected.
	Accounts AccountRepository
	API      AccountAPI
	OAuth    OAuth
}

// Service bundles both input ports; one value implements them so CalDAV
// calendars and subscriptions share the agenda cache.
type Service interface {
	Calendar
	Subscriptions
	Accounts
}

type calendar struct {
	repo   SettingsRepository
	source EventSource
	writer EventWriter
	finder CalendarFinder
	subs   SubscriptionRepository
	feeds  FeedSource
	ids    kernel.IDGenerator
	clock  kernel.Clock
	// connected accounts (#282)
	accounts AccountRepository
	api      AccountAPI
	oauth    OAuth
	acc      accountState
	mu       sync.RWMutex
	// agendas caches the agenda per account whose calendars apply
	// (kernel.Principal.SettingsAccount, ADR 0018).
	agendas map[string]*agendaCache
	// shared and sharingOn: see Deps.
	shared         SharedSubscriptions
	sharedAccounts SharedAccounts
	sharingOn      func(ctx context.Context) bool
}

// agendaCache is the cached agenda of one account and the period it covers.
type agendaCache struct {
	cached               *AgendaResponse
	windowFrom, windowTo time.Time
}

// cacheOf returns the agenda cache of the principal's account, creating it;
// the caller holds uc.mu for writing.
func (uc *calendar) cacheOf(ctx context.Context) *agendaCache {
	key := kernel.PrincipalFrom(ctx).SettingsAccount()
	c := uc.agendas[key]
	if c == nil {
		c = &agendaCache{}
		uc.agendas[key] = c
	}
	return c
}

// snapshot reads the principal's cached agenda and its window.
func (uc *calendar) snapshot(ctx context.Context) (*AgendaResponse, time.Time, time.Time) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	c := uc.agendas[kernel.PrincipalFrom(ctx).SettingsAccount()]
	if c == nil {
		return nil, time.Time{}, time.Time{}
	}
	return c.cached, c.windowFrom, c.windowTo
}

// NewCalendar returns the calendar use cases.
func NewCalendar(d Deps) Service {
	return &calendar{repo: d.Repo, source: d.Source, writer: d.Writer, finder: d.Finder, subs: d.Subscriptions, feeds: d.Feeds, ids: d.IDs, clock: d.Clock,
		accounts: d.Accounts, api: d.API, oauth: d.OAuth, acc: accountState{pending: map[string]*pendingSignIn{}}, agendas: map[string]*agendaCache{},
		shared: d.Shared, sharedAccounts: d.SharedAccounts, sharingOn: d.SharingOn}
}

// sourceSet is everything the agenda is built from.
type sourceSet struct {
	caldav   *domain.Settings // nil when not configured
	subs     []domain.Subscription
	accounts []domain.Account
}

// listAccounts returns the connected accounts; none when the module was
// built without account support.
func (uc *calendar) listAccounts(ctx context.Context) ([]domain.Account, error) {
	if uc.accounts == nil {
		return nil, nil
	}
	list, err := uc.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return list, nil
}

// sources returns every configured source; domain.ErrNotConfigured when
// there is none.
func (uc *calendar) sources(ctx context.Context) (sourceSet, error) {
	caldav, subs, err := uc.feedSources(ctx)
	if err != nil && !errors.Is(err, domain.ErrNotConfigured) {
		return sourceSet{}, err
	}
	accounts, aerr := uc.listAccounts(ctx)
	if aerr != nil {
		return sourceSet{}, aerr
	}
	if err != nil && len(accounts) == 0 {
		return sourceSet{}, err
	}
	return sourceSet{caldav: caldav, subs: subs, accounts: accounts}, nil
}

// sharedSubscriptions returns the calendars others shared with a person
// while the team calendar is on. Guests, kiosk displays and background work
// see the owner's calendars as before and nothing shared. They are never
// part of a person's cached agenda: whether they show depends on the
// person's groups and the licence at the time of the request, so they are
// merged in per request (the feed itself is cached by the gateway).
func (uc *calendar) sharedSubscriptions(ctx context.Context) (sharedSet, error) {
	p := kernel.PrincipalFrom(ctx)
	if uc.sharingOn == nil || p.TabScoped() || p.IsSystem() || p.IsToken() || !uc.sharingOn(ctx) {
		return sharedSet{}, nil
	}
	var set sharedSet
	if uc.shared != nil {
		list, err := uc.shared.ListShared(ctx)
		if err != nil {
			return sharedSet{}, fmt.Errorf("list shared subscriptions: %w", err)
		}
		set.subs = list
	}
	if uc.sharedAccounts != nil {
		list, err := uc.sharedAccounts.ListSharedAccounts(ctx)
		if err != nil {
			return sharedSet{}, fmt.Errorf("list shared accounts: %w", err)
		}
		set.accounts = list
	}
	return set, nil
}

// sharedSet is what others shared with a person.
type sharedSet struct {
	subs     []domain.Subscription
	accounts []SharedAccount
}

func (s sharedSet) empty() bool { return len(s.subs) == 0 && len(s.accounts) == 0 }

// withShared adds the events of shared calendars between from and to to
// res, sorted with the rest; a shared feed that fails is named in Error.
func (uc *calendar) withShared(ctx context.Context, res AgendaResponse, shared sharedSet, from, to time.Time) AgendaResponse {
	if shared.empty() {
		return res
	}
	var failed []string
	add := func(e domain.Event, calendar, color string) {
		if !e.Start.Before(to) || !(domain.Event{End: e.End, AllDay: e.AllDay}).EndsAfter(from) {
			return
		}
		res.Events = append(res.Events, EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Description: e.Description,
			URL: e.URL, Calendar: calendar, Color: color, Start: e.Start, End: e.End, AllDay: e.AllDay})
	}
	for _, sub := range shared.subs {
		events, err := uc.feeds.Events(ctx, sub, from, to)
		if err != nil {
			failed = append(failed, sub.Name)
			continue
		}
		for _, e := range events {
			add(e, sub.Name, sub.Color)
		}
	}
	for _, sa := range shared.accounts {
		// The account is read as its owner: a token refresh is stored with
		// the owner's account, and providers that rotate refresh tokens
		// (Microsoft) keep working for the owner. accessToken serialises
		// refreshes and re-reads the stored token first.
		owner := kernel.WithPrincipal(ctx, kernel.PrincipalFor(sa.OwnerID))
		events, errs, _ := uc.accountEvents(owner, []domain.Account{sa.Account}, from, to)
		failed = append(failed, errs...)
		for _, e := range events {
			add(e, e.Calendar, e.Color)
		}
	}
	slices.SortStableFunc(res.Events, func(a, b EventView) int { return a.Start.Compare(b.Start) })
	if len(failed) > 0 {
		msg := "Some shared calendars could not be loaded: " + strings.Join(failed, ", ")
		if res.Error != "" {
			msg = res.Error + " " + msg
		}
		res.Error = msg
	}
	return res
}

// feedSources returns the CalDAV settings (nil when not configured) and the
// subscriptions; domain.ErrNotConfigured when there is neither.
func (uc *calendar) feedSources(ctx context.Context) (*domain.Settings, []domain.Subscription, error) {
	var caldav *domain.Settings
	s, err := uc.repo.Load(ctx)
	switch {
	case err == nil:
		caldav = &s
	case !errors.Is(err, domain.ErrNotConfigured):
		return nil, nil, fmt.Errorf("load settings: %w", err)
	}
	subs, err := uc.subs.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list subscriptions: %w", err)
	}
	if caldav == nil && len(subs) == 0 {
		return nil, nil, domain.ErrNotConfigured
	}
	return caldav, subs, nil
}

// Settings implements Calendar.
func (uc *calendar) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{CalendarURLs: []string{}}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, CalendarURLs: s.CalendarURLs, Username: s.Username, HasPassword: true}, nil
}

// SaveSettings implements Calendar; every calendar is queried once so wrong
// URLs or credentials are reported immediately.
func (uc *calendar) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	password := in.Password
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			wanted := []string{in.Username}
			for _, u := range in.CalendarURLs {
				if strings.TrimSpace(u) != "" {
					wanted = append(wanted, u)
				}
			}
			if kernel.TargetChanged(append([]string{old.Username}, old.CalendarURLs...), wanted) {
				return SettingsView{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	s, err := domain.NewSettings(in.CalendarURLs, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	now := uc.clock.Now()
	if _, err := uc.fetchRange(ctx, sourceSet{caldav: &s}, now, now.Add(domain.Horizon)); err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	_ = uc.Refresh(ctx) // errors are reported through AgendaResponse.Error
	return SettingsView{Configured: true, CalendarURLs: s.CalendarURLs, Username: s.Username, HasPassword: true}, nil
}

// Connect implements Calendar. Errors name the cause in the setup
// assistant's words (kernel.SetupFailure); the calendars are then checked
// and saved like the form does.
func (uc *calendar) Connect(ctx context.Context, in ConnectInput) (ConnectView, error) {
	server, user := strings.TrimSpace(in.Server), strings.TrimSpace(in.Username)
	// The server address follows the same rules as a calendar address.
	if _, err := domain.NewSettings([]string{server}, user, in.Password); err != nil {
		return ConnectView{}, err
	}
	if uc.finder == nil {
		return ConnectView{}, domain.ErrNoCalendars
	}
	found, err := uc.finder.Discover(ctx, server, user, in.Password)
	if err != nil {
		return ConnectView{}, kernel.SetupFailure("Nextcloud", err)
	}
	if len(found) == 0 {
		return ConnectView{}, domain.ErrNoCalendars
	}
	urls := make([]string, 0, len(found))
	for _, f := range found {
		urls = append(urls, f.URL)
	}
	view, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: user, Password: in.Password})
	if err != nil {
		return ConnectView{}, err
	}
	out := ConnectView{Settings: view, Calendars: found}
	if agenda, err := uc.Upcoming(ctx, 14); err == nil {
		out.Next = agenda.Events[:min(len(agenda.Events), connectPreview)]
	}
	return out, nil
}

// Upcoming implements Calendar (cache-aside).
func (uc *calendar) Upcoming(ctx context.Context, days int) (AgendaResponse, error) {
	if days <= 0 {
		days = DefaultDays
	}
	days = min(days, MaxDays)
	now := uc.clock.Now()
	until := now.Add(time.Duration(days) * 24 * time.Hour)
	shared, err := uc.sharedSubscriptions(ctx)
	if err != nil {
		return AgendaResponse{}, err
	}
	cached, _, _ := uc.snapshot(ctx)
	if cached == nil {
		if _, err := uc.sources(ctx); err != nil {
			if errors.Is(err, domain.ErrNotConfigured) && !shared.empty() {
				return uc.withShared(ctx, AgendaResponse{UpdatedAt: now, Events: []EventView{}}, shared, now, until), nil
			}
			return AgendaResponse{}, err
		}
		_ = uc.Refresh(ctx) // errors are reported through AgendaResponse.Error
		cached, _, _ = uc.snapshot(ctx)
		if cached == nil {
			return AgendaResponse{}, kernel.NewPublicError(kernel.ErrUnavailable, "The calendar server is not reachable.")
		}
	}
	res := AgendaResponse{UpdatedAt: cached.UpdatedAt, Error: cached.Error, Events: []EventView{}}
	for _, e := range cached.Events {
		if e.Start.Before(until) && (domain.Event{End: e.End, AllDay: e.AllDay}).EndsAfter(now) {
			res.Events = append(res.Events, e)
		}
	}
	return uc.withShared(ctx, res, shared, now, until), nil
}

// Range implements Calendar. Periods inside the cached window are answered
// from the cache, others are loaded directly (and not cached).
func (uc *calendar) Range(ctx context.Context, from, to time.Time) (AgendaResponse, error) {
	if err := domain.CheckRange(from, to); err != nil {
		return AgendaResponse{}, err
	}
	shared, err := uc.sharedSubscriptions(ctx)
	if err != nil {
		return AgendaResponse{}, err
	}
	src, err := uc.sources(ctx)
	if errors.Is(err, domain.ErrNotConfigured) && !shared.empty() {
		return uc.withShared(ctx, AgendaResponse{UpdatedAt: uc.clock.Now(), Events: []EventView{}}, shared, from, to), nil
	}
	if err != nil {
		return AgendaResponse{}, err
	}
	cached, wFrom, wTo := uc.snapshot(ctx)
	agenda := cached
	if cached == nil || from.Before(wFrom) || to.After(wTo) {
		live, err := uc.fetchRange(ctx, src, from, to)
		if err != nil {
			return AgendaResponse{}, err
		}
		agenda = &live
	}
	res := AgendaResponse{UpdatedAt: agenda.UpdatedAt, Error: agenda.Error, Events: []EventView{}}
	for _, e := range agenda.Events {
		if e.Start.Before(to) && e.End.After(from) {
			res.Events = append(res.Events, e)
		}
	}
	return uc.withShared(ctx, res, shared, from, to), nil
}

// CreateEvent implements Calendar. CalendarURL is a CalDAV collection or a
// calendar of a connected account (domain.CalendarRef).
func (uc *calendar) CreateEvent(ctx context.Context, in NewEventInput) (EventView, error) {
	end := in.End
	if in.AllDay {
		end = end.AddDate(0, 0, 1) // inclusive date → exclusive end
	}
	e, err := domain.NewEvent(uc.ids.NewID()+"@waylight", in.Summary, in.Location, in.Start, end, in.AllDay)
	if err != nil {
		return EventView{}, err
	}
	var name string
	if accountID, calendarID, ok := domain.ParseCalendarRef(in.CalendarURL); ok && uc.accounts != nil {
		if name, err = uc.createInAccount(ctx, accountID, calendarID, e); err != nil {
			return EventView{}, err
		}
	} else {
		s, err := uc.repo.Load(ctx)
		if errors.Is(err, domain.ErrNotConfigured) {
			return EventView{}, domain.ErrUnknownCalendar
		}
		if err != nil {
			return EventView{}, err
		}
		if !s.Contains(in.CalendarURL) {
			return EventView{}, domain.ErrUnknownCalendar
		}
		if err := uc.writer.Create(ctx, s, in.CalendarURL, e); err != nil {
			return EventView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
		}
		name = domain.CalendarName(in.CalendarURL)
	}
	_ = uc.Refresh(ctx) // the new event shows up in the agenda right away
	return EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Calendar: name,
		Start: e.Start, End: e.End, AllDay: e.AllDay}, nil
}

// Refresh implements Calendar. For a request it reloads the principal's
// agenda; the background job (the system) reloads the admin's and every
// account's agenda that was asked for since the start.
func (uc *calendar) Refresh(ctx context.Context) error {
	if !kernel.PrincipalFrom(ctx).IsSystem() {
		return uc.refreshOne(ctx)
	}
	uc.mu.RLock()
	accounts := []string{kernel.SystemAccountID}
	for key := range uc.agendas {
		if key != kernel.SystemAccountID {
			accounts = append(accounts, key)
		}
	}
	uc.mu.RUnlock()
	var errs []error
	for _, a := range accounts {
		errs = append(errs, uc.refreshOne(kernel.WithPrincipal(ctx, kernel.PrincipalFor(a))))
	}
	return errors.Join(errs...)
}

// refreshOne reloads the agenda of the principal's account.
func (uc *calendar) refreshOne(ctx context.Context) error {
	src, err := uc.sources(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		uc.mu.Lock()
		uc.cacheOf(ctx).cached = nil // the last source was removed
		uc.mu.Unlock()
		return nil
	}
	if err != nil {
		return err
	}
	agenda, err := uc.fetch(ctx, src)
	if err != nil {
		uc.mu.Lock()
		c := uc.cacheOf(ctx)
		res := AgendaResponse{}
		if c.cached != nil {
			res = *c.cached
		}
		res.Error = "The calendar server is not reachable; showing the last known events."
		c.cached = &res
		uc.mu.Unlock()
		return err
	}
	uc.store(ctx, agenda)
	return nil
}

func (uc *calendar) fetch(ctx context.Context, src sourceSet) (AgendaResponse, error) {
	now := uc.clock.Now()
	from := now.Add(-24 * time.Hour) // include today's all-day events in every time zone
	res, err := uc.fetchRange(ctx, src, from, now.Add(domain.Horizon))
	if err == nil {
		uc.mu.Lock()
		c := uc.cacheOf(ctx)
		c.windowFrom, c.windowTo = from, now.Add(domain.Horizon)
		uc.mu.Unlock()
	}
	return res, err
}

// fetchRange merges the events of every CalDAV calendar, subscription and
// account calendar; it fails only when every source failed.
func (uc *calendar) fetchRange(ctx context.Context, src sourceSet, from, to time.Time) (AgendaResponse, error) {
	now := uc.clock.Now()
	caldav, subs := src.caldav, src.subs
	all, failed, total := uc.accountEvents(ctx, src.accounts, from, to)
	total += len(subs)
	if caldav != nil {
		total += len(caldav.CalendarURLs)
		for _, u := range caldav.CalendarURLs {
			events, err := uc.source.Events(ctx, *caldav, u, from, to)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (%v)", domain.CalendarName(u), err))
				continue
			}
			all = append(all, events...)
		}
	}
	for _, sub := range subs {
		events, err := uc.feeds.Events(ctx, sub, from, to)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", sub.Name, err))
			continue
		}
		for _, e := range events {
			e.Calendar, e.Color = sub.Name, sub.Color
			all = append(all, e)
		}
	}
	if total > 0 && len(failed) == total {
		return AgendaResponse{}, &fetchError{failures: failed}
	}
	domain.SortEvents(all)
	res := AgendaResponse{UpdatedAt: now, Events: make([]EventView, 0, len(all))}
	if len(failed) > 0 {
		res.Error = "Some calendars could not be loaded: " + strings.Join(failed, "; ")
	}
	for _, e := range all {
		res.Events = append(res.Events, EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Description: e.Description, URL: e.URL, Calendar: e.Calendar,
			Color: e.Color, Start: e.Start, End: e.End, AllDay: e.AllDay})
	}
	return res, nil
}

// fetchError lists the calendars that could not be loaded. Its message is
// shown to the user, so it contains no error-kind suffix.
type fetchError struct{ failures []string }

func (e *fetchError) Error() string { return strings.Join(e.failures, "; ") }

// Unwrap classifies the error as an unavailable dependency.
func (e *fetchError) Unwrap() error { return kernel.ErrUnavailable }

func (uc *calendar) store(ctx context.Context, a AgendaResponse) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	a.Events = slices.Clone(a.Events)
	uc.cacheOf(ctx).cached = &a
}
