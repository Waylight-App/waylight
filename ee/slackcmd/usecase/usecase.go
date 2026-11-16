// Package usecase answers Slack slash commands for the member behind them
// (#333, ADR 0023).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"strings"
	"sync"
	"time"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/ee/slackcmd/domain"
)

// Command is a slash command as Slack delivers it.
type Command struct {
	TeamID, UserID string
	Command, Text  string
}

// Connections finds the members who connected Slack and their tokens
// (output port, #326).
type Connections interface {
	AccountsWith(ctx context.Context, service string) ([]string, error)
	AccessToken(ctx context.Context, accountID, service string) (string, error)
}

// Slack tells whose a user token is (output port, auth.test).
type Slack interface {
	Identity(ctx context.Context, userToken string) (teamID, userID string, err error)
}

// Finder searches the start page as the given account (output port; the
// server composes services, bookmarks and wikis).
type Finder interface {
	Find(ctx context.Context, accountID, query string) ([]domain.Hit, error)
}

// Licensing tells whether chat apps are licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Deps are the collaborators.
type Deps struct {
	Connections Connections
	Slack       Slack
	Finder      Finder
	Licensing   Licensing
	Now         func() time.Time
	// SettingsURL is where someone connects their Slack account; "" without
	// PUBLIC_URL.
	SettingsURL string
}

// rescanAfter limits how often an unknown Slack user makes the service ask
// Slack again who every connected token belongs to.
const rescanAfter = 30 * time.Second

// Service implements the use case.
type Service struct {
	d Deps

	mu       sync.Mutex
	members  map[string]string // team|user → account
	scanned  time.Time
	scanning sync.Mutex
}

// New returns the use case.
func New(d Deps) *Service { return &Service{d: d, members: map[string]string{}} }

// Handle answers a command. The reply is only seen by the person who asked.
func (s *Service) Handle(ctx context.Context, c Command) string {
	if s.d.Licensing.Level(ctx, licence.FeatureChatApps) != licence.LevelFull {
		return domain.Unlicensed
	}
	query := strings.TrimSpace(c.Text)
	if query == "" {
		return domain.Help(c.Command)
	}
	account, ok := s.member(ctx, c.TeamID, c.UserID)
	if !ok {
		return domain.NotConnected(s.d.SettingsURL)
	}
	hits, err := s.d.Finder.Find(ctx, account, query)
	if err != nil && len(hits) == 0 {
		return "The search did not work this time; try again or use the start page."
	}
	return domain.Results(query, hits, false)
}

// member finds the account whose connected Slack token belongs to the user.
// Known pairs are kept in memory; an unknown user triggers a new scan of
// the connected tokens at most every rescanAfter.
func (s *Service) member(ctx context.Context, team, user string) (string, bool) {
	key := team + "|" + user
	s.mu.Lock()
	acc, ok := s.members[key]
	stale := s.d.Now().Sub(s.scanned) >= rescanAfter
	s.mu.Unlock()
	if ok {
		return acc, true
	}
	if !stale {
		return "", false
	}
	s.scan(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	acc, ok = s.members[key]
	return acc, ok
}

// Warm fills the map of Slack users ahead of the first command, so the
// first answer does not wait for Slack.
func (s *Service) Warm(ctx context.Context) { s.scan(ctx) }

// scanWorkers bounds the parallel auth.test calls of a scan.
const scanWorkers = 8

// scan asks Slack who each connected token belongs to and replaces the map,
// so a member who disconnected Slack is forgotten. The calls run in
// parallel: a command has to be answered within Slack's three seconds.
func (s *Service) scan(ctx context.Context) {
	s.scanning.Lock()
	defer s.scanning.Unlock()
	accounts, err := s.d.Connections.AccountsWith(ctx, "slack")
	if err != nil {
		return
	}
	var mu sync.Mutex
	found := map[string]string{}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for range min(scanWorkers, len(accounts)) {
		wg.Go(func() {
			for acc := range jobs {
				token, err := s.d.Connections.AccessToken(ctx, acc, "slack")
				if err != nil {
					continue
				}
				team, user, err := s.d.Slack.Identity(ctx, token)
				if err != nil {
					continue
				}
				mu.Lock()
				found[team+"|"+user] = acc
				mu.Unlock()
			}
		})
	}
	for _, acc := range accounts {
		jobs <- acc
	}
	close(jobs)
	wg.Wait()
	s.mu.Lock()
	s.members, s.scanned = found, s.d.Now()
	s.mu.Unlock()
}
