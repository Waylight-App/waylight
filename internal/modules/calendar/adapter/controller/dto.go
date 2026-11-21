package controller

// SettingsInputDTO is the body of PUT /api/calendar/settings; an empty
// password keeps the stored one.
type SettingsInputDTO struct {
	CalendarURLs []string `json:"calendarUrls"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
}

// ConnectInputDTO is the body of POST /api/calendar/connect (#588): the
// server address, user name and app password.
type ConnectInputDTO struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	// SignIn is a finished Nextcloud sign-in of the setup assistant
	// (#654); it replaces server, username and password.
	SignIn string `json:"signIn,omitempty"`
}

// NewEventDTO is the body of POST /api/calendar/events. Timed events use
// RFC 3339 instants; all-day events use dates (YYYY-MM-DD) with an
// inclusive end date.
type NewEventDTO struct {
	CalendarURL string `json:"calendarUrl"`
	Summary     string `json:"summary"`
	Location    string `json:"location"`
	AllDay      bool   `json:"allDay"`
	Start       string `json:"start"`
	End         string `json:"end"`
}

// SubscriptionInputDTO is the body of POST /api/calendar/subscriptions and
// PUT /api/calendar/subscriptions/{id}; an empty URL keeps the stored one
// when updating.
type SubscriptionInputDTO struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Color          string `json:"color"`
	RefreshMinutes int    `json:"refreshMinutes"`
}

// AccountCalendarDTO is a calendar of a connected account (#282).
type AccountCalendarDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Writable bool   `json:"writable"`
	Enabled  bool   `json:"enabled"`
}

// AccountDTO is a connected account; tokens never leave the server.
type AccountDTO struct {
	ID        string               `json:"id"`
	Provider  string               `json:"provider"`
	Name      string               `json:"name"`
	Calendars []AccountCalendarDTO `json:"calendars"`
}

// AccountsDTO lists the accounts and the redirect address to register
// ("" when PUBLIC_URL is not set and signing in is unavailable).
type AccountsDTO struct {
	RedirectURL string       `json:"redirectUrl"`
	Accounts    []AccountDTO `json:"accounts"`
}

// SignInInputDTO starts a sign-in with the user's app registration.
type SignInInputDTO struct {
	Provider     string `json:"provider"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// SignInStartDTO is where the sign-in window goes.
type SignInStartDTO struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

// SignInStatusDTO is "pending", "ready" or "failed".
type SignInStatusDTO struct {
	Status string `json:"status"`
}

// AccountCalendarsInputDTO lists the calendars to show.
type AccountCalendarsInputDTO struct {
	Enabled []string `json:"enabled"`
}

// CalendarChoiceDTO is a calendar new events can be written to.
type CalendarChoiceDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
