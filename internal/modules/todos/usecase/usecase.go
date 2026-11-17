package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/shared/kernel"
)

// TodoInput carries the editable fields of a to-do. Nil Repeat or
// Reminder keep the stored values (older clients only send title and due).
type TodoInput struct {
	Title    string
	Due      time.Time
	Repeat   *RepeatInput
	Reminder *ReminderInput
	// ID is a client-generated ID for create (#80): replaying the same
	// create returns the existing to-do instead of a duplicate. Empty = the
	// server picks one.
	ID string
	// Version is the version the client edited (update); 0 = no check.
	Version int
	// Source is the list ("" or "local" = the built-in one, else the ID of
	// a connected external list, #169).
	Source string
}

// RepeatInput describes a recurrence (#68); an empty Kind means none.
type RepeatInput struct {
	Kind  string
	Every int
	Days  []int
}

// ReminderInput describes a reminder (#68); an empty At means none.
type ReminderInput struct {
	At         string
	DaysBefore int
}

// RepeatView is a recurrence at the use case boundary.
type RepeatView struct {
	Kind  string
	Every int
	Days  []int
}

// ReminderView is a reminder at the use case boundary.
type ReminderView struct {
	At         string
	DaysBefore int
}

// TodoView is a to-do at the use case boundary.
type TodoView struct {
	ID          string
	Title       string
	Due         time.Time
	Done        bool
	Overdue     bool
	Position    int
	CreatedAt   time.Time
	CompletedAt time.Time
	Repeat      RepeatView
	Reminder    ReminderView
	Version     int
	// Visibility is "private", "everyone" or "group:<id>" for local
	// to-dos and empty for external lists (ADR 0018).
	Visibility string
	// Mine is true when the viewer owns the to-do and may share it.
	Mine bool
}

// ListTodos returns open to-dos first (by position), then done ones (input port).
type ListTodos interface {
	Execute(ctx context.Context, source string, includeDone bool) ([]TodoView, error)
}

// CreateTodo appends a to-do (input port).
type CreateTodo interface {
	Execute(ctx context.Context, in TodoInput) (TodoView, error)
}

// UpdateTodo edits text, due date and completion (input port).
type UpdateTodo interface {
	Execute(ctx context.Context, id string, in TodoInput, done bool) (TodoView, error)
}

// DeleteTodo removes a to-do (input port).
type DeleteTodo interface {
	Execute(ctx context.Context, source, id string) error
}

// ClearCompleted removes all done to-dos (input port).
type ClearCompleted interface {
	Execute(ctx context.Context, source string) (int, error)
}

// SendDueReminders pushes reminders of open to-dos that are due (input
// port, run every minute); it returns how many were sent.
type SendDueReminders interface {
	Execute(ctx context.Context) (int, error)
}

// Repository persists to-dos (output port). The local list and every
// external provider (#169) implement it; external ones fail writes with
// domain.ErrVersionConflict when the to-do changed in the other app.
type Repository interface {
	List(ctx context.Context) ([]domain.Todo, error)
	Get(ctx context.Context, id string) (domain.Todo, error)
	Save(ctx context.Context, t domain.Todo) error
	Delete(ctx context.Context, ids ...string) error
}

// VisibilityStore changes who else may see a local to-do (output port).
type VisibilityStore interface {
	SetVisibility(ctx context.Context, id string, v kernel.Visibility) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Clock kernel.Clock
	IDs   kernel.IDGenerator
	// Location is the user's time zone for completion days and reminder
	// times (nil = time.Local).
	Location *time.Location
	// Alerts delivers reminders (nil = reminders are not sent).
	Alerts kernel.AlertSink
	// Sources stores the connected external lists and Providers opens them
	// (#169); nil = only the local list.
	Sources   SourceRepository
	Providers Providers
	// OAuth signs in to Microsoft, Google and TickTick (#171); nil = off.
	OAuth OAuth
	// Visibility shares local to-dos (ADR 0018, #301); nil = not offered.
	Visibility VisibilityStore
}

// list is the resolved to-do list of a request.
type list struct {
	repo     Repository
	caps     domain.Capabilities
	external bool
}

// open resolves a source ID to its repository.
func (d Deps) open(ctx context.Context, source string) (list, error) {
	if source == "" || source == domain.LocalSource {
		return list{repo: d.Repo, caps: domain.CapabilitiesOf("")}, nil
	}
	if d.Sources == nil || d.Providers == nil || !domain.ValidSourceID(source) {
		return list{}, domain.ErrSourceNotFound
	}
	s, err := d.Sources.Get(ctx, source)
	if err != nil {
		return list{}, err
	}
	return list{repo: external{d.Providers.Open(s), s.Kind}, caps: domain.CapabilitiesOf(s.Kind), external: true}, nil
}

// external wraps an external provider: errors that are not domain errors
// (network, refused login) become "unavailable" with the target named.
type external struct {
	r    Repository
	kind domain.SourceKind
}

func unavailable(err error) error {
	var pe *kernel.PublicError
	if err == nil || errors.As(err, &pe) {
		return err
	}
	return kernel.TargetFailure(kernel.ErrUnavailable, err)
}

func (e external) List(ctx context.Context) ([]domain.Todo, error) {
	t, err := e.r.List(ctx)
	return t, unavailable(err)
}

func (e external) Get(ctx context.Context, id string) (domain.Todo, error) {
	valid := domain.ValidExternalID(id)
	if e.kind != domain.KindCalDAV {
		valid = domain.ValidProviderID(id)
	}
	if !valid {
		return domain.Todo{}, domain.ErrTodoNotFound
	}
	t, err := e.r.Get(ctx, id)
	return t, unavailable(err)
}

func (e external) Save(ctx context.Context, t domain.Todo) error {
	return unavailable(e.r.Save(ctx, t))
}

func (e external) Delete(ctx context.Context, ids ...string) error {
	return unavailable(e.r.Delete(ctx, ids...))
}

func (d Deps) loc() *time.Location {
	if d.Location == nil {
		return time.Local
	}
	return d.Location
}

// applySchedule sets recurrence and reminder from the input; nil fields
// keep the current values. Lists without them refuse a set rule.
func applySchedule(t *domain.Todo, in TodoInput, caps domain.Capabilities) error {
	if !caps.Repeat && in.Repeat != nil && in.Repeat.Kind != "" || !caps.Reminders && in.Reminder != nil && in.Reminder.At != "" {
		return domain.ErrUnsupported
	}
	repeat, reminder := t.Repeat(), t.Reminder()
	var err error
	if in.Repeat != nil {
		if repeat, err = domain.NewRecurrence(domain.RepeatKind(in.Repeat.Kind), in.Repeat.Every, in.Repeat.Days); err != nil {
			return err
		}
	}
	if in.Reminder != nil {
		if reminder, err = domain.NewReminder(in.Reminder.At, in.Reminder.DaysBefore); err != nil {
			return err
		}
	}
	return t.Schedule(repeat, reminder)
}

func nextPosition(todos []domain.Todo) int {
	next := 0
	for _, t := range todos {
		next = max(next, t.Position()+1)
	}
	return next
}

func (d Deps) view(ctx context.Context, t domain.Todo) TodoView {
	return TodoView{
		ID: t.ID(), Title: t.Title(), Due: t.Due(), Done: t.Done(), Overdue: t.IsOverdue(d.Clock.Now()),
		Position: t.Position(), CreatedAt: t.CreatedAt(), CompletedAt: t.CompletedAt(),
		Repeat:     RepeatView{Kind: string(t.Repeat().Kind()), Every: t.Repeat().Every(), Days: t.Repeat().Days()},
		Reminder:   ReminderView{At: t.Reminder().At(), DaysBefore: t.Reminder().DaysBefore()},
		Version:    t.Version(),
		Visibility: string(t.Access().Visibility), Mine: t.Access().Mine(kernel.PrincipalFrom(ctx)),
	}
}

type listTodos struct{ d Deps }

// NewListTodos returns the ListTodos use case.
func NewListTodos(d Deps) ListTodos { return &listTodos{d: d} }

// Execute implements ListTodos.
func (uc *listTodos) Execute(ctx context.Context, source string, includeDone bool) ([]TodoView, error) {
	l, err := uc.d.open(ctx, source)
	if err != nil {
		return nil, err
	}
	todos, err := l.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	slices.SortStableFunc(todos, func(a, b domain.Todo) int {
		switch {
		case a.Done() != b.Done():
			if a.Done() {
				return 1
			}
			return -1
		case a.Done():
			return b.CompletedAt().Compare(a.CompletedAt())
		default:
			return a.Position() - b.Position()
		}
	})
	out := make([]TodoView, 0, len(todos))
	for _, t := range todos {
		if includeDone || !t.Done() {
			out = append(out, uc.d.view(ctx, t))
		}
	}
	return out, nil
}

type createTodo struct{ d Deps }

// NewCreateTodo returns the CreateTodo use case.
func NewCreateTodo(d Deps) CreateTodo { return &createTodo{d: d} }

// Execute implements CreateTodo.
func (uc *createTodo) Execute(ctx context.Context, in TodoInput) (TodoView, error) {
	l, err := uc.d.open(ctx, in.Source)
	if err != nil {
		return TodoView{}, err
	}
	id := uc.d.IDs.NewID()
	if in.ID != "" {
		if !kernel.ValidClientID(in.ID) {
			return TodoView{}, kernel.ErrInvalidClientID
		}
		// A replayed create (offline outbox, #80) returns what it created.
		if existing, err := l.repo.Get(ctx, in.ID); err == nil {
			return uc.d.view(ctx, existing), nil
		} else if !errors.Is(err, domain.ErrTodoNotFound) {
			return TodoView{}, err
		}
		id = in.ID
	}
	position := 0
	if l.caps.Ordering {
		todos, err := l.repo.List(ctx)
		if err != nil {
			return TodoView{}, fmt.Errorf("list todos: %w", err)
		}
		position = nextPosition(todos)
	}
	t, err := domain.NewTodo(id, in.Title, in.Due, position, uc.d.Clock.Now())
	if err != nil {
		return TodoView{}, err
	}
	if in.Source == "" {
		t = t.WithAccess(kernel.NewAccess(kernel.PrincipalFrom(ctx)))
	}
	if err := applySchedule(&t, in, l.caps); err != nil {
		return TodoView{}, err
	}
	if err := l.repo.Save(ctx, t); err != nil {
		return TodoView{}, fmt.Errorf("save todo: %w", err)
	}
	return uc.d.saved(ctx, l, t)
}

// saved returns the view of a stored to-do; external lists are read back,
// so the answer has the provider's ID and version (ETag).
func (d Deps) saved(ctx context.Context, l list, t domain.Todo) (TodoView, error) {
	if l.external {
		stored, err := l.repo.Get(ctx, t.ID())
		if err != nil {
			return TodoView{}, err
		}
		t = stored
	}
	return d.view(ctx, t), nil
}

type updateTodo struct{ d Deps }

// NewUpdateTodo returns the UpdateTodo use case.
func NewUpdateTodo(d Deps) UpdateTodo { return &updateTodo{d: d} }

// Execute implements UpdateTodo.
func (uc *updateTodo) Execute(ctx context.Context, id string, in TodoInput, done bool) (TodoView, error) {
	l, err := uc.d.open(ctx, in.Source)
	if err != nil {
		return TodoView{}, err
	}
	t, err := l.repo.Get(ctx, id)
	if err != nil {
		return TodoView{}, err
	}
	if err := t.CheckVersion(in.Version); err != nil {
		return TodoView{}, err
	}
	if err := t.Edit(in.Title, in.Due); err != nil {
		return TodoView{}, err
	}
	if err := applySchedule(&t, in, l.caps); err != nil {
		return TodoView{}, err
	}
	now := uc.d.Clock.Now()
	completing := done && !t.Done()
	t.SetDone(done, now)
	t.Bump()
	if completing {
		// Completing a recurring to-do creates its next occurrence (#68).
		if err := uc.createNext(ctx, l.repo, t, now); err != nil {
			return TodoView{}, err
		}
	}
	if err := l.repo.Save(ctx, t); err != nil {
		return TodoView{}, fmt.Errorf("save todo: %w", err)
	}
	return uc.d.saved(ctx, l, t)
}

func (uc *updateTodo) createNext(ctx context.Context, repo Repository, t domain.Todo, now time.Time) error {
	if !t.Repeat().Repeats() {
		return nil
	}
	todos, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list todos: %w", err)
	}
	next, _ := t.NextOccurrence(uc.d.IDs.NewID(), nextPosition(todos), now.In(uc.d.loc()), now)
	if err := repo.Save(ctx, next); err != nil {
		return fmt.Errorf("save next occurrence: %w", err)
	}
	return nil
}

type sendDueReminders struct{ d Deps }

// NewSendDueReminders returns the SendDueReminders use case.
func NewSendDueReminders(d Deps) SendDueReminders { return &sendDueReminders{d: d} }

// Execute implements SendDueReminders.
func (uc *sendDueReminders) Execute(ctx context.Context) (int, error) {
	if uc.d.Alerts == nil {
		return 0, nil
	}
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list todos: %w", err)
	}
	now, loc := uc.d.Clock.Now(), uc.d.loc()
	sent := 0
	for _, t := range todos {
		if !t.ReminderDue(now, loc) {
			continue
		}
		// Mark first: a failed save must not lead to a push every minute.
		t.MarkReminded()
		t.Bump()
		if err := uc.d.Repo.Save(ctx, t); err != nil {
			return sent, fmt.Errorf("save todo: %w", err)
		}
		uc.d.Alerts.Alert(ctx, kernel.Alert{Kind: "reminder", Title: t.Title(), Body: reminderBody(t, now.In(loc))})
		sent++
	}
	return sent, nil
}

// reminderBody describes the due date relative to today (English; push
// texts follow the server language like other alerts).
func reminderBody(t domain.Todo, today time.Time) string {
	y, m, d := today.Date()
	days := int(t.Due().Sub(time.Date(y, m, d, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	switch {
	case days == 0:
		return "To-do due today."
	case days == 1:
		return "To-do due tomorrow."
	case days < 0:
		return fmt.Sprintf("To-do overdue since %s.", t.Due().Format("2 Jan 2006"))
	default:
		return fmt.Sprintf("To-do due on %s (in %d days).", t.Due().Format("Mon, 2 Jan"), days)
	}
}

type deleteTodo struct{ d Deps }

// NewDeleteTodo returns the DeleteTodo use case.
func NewDeleteTodo(d Deps) DeleteTodo { return &deleteTodo{d: d} }

// Execute implements DeleteTodo.
func (uc *deleteTodo) Execute(ctx context.Context, source, id string) error {
	l, err := uc.d.open(ctx, source)
	if err != nil {
		return err
	}
	t, err := l.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := l.repo.Delete(ctx, t.ID()); err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	return nil
}

type clearCompleted struct{ d Deps }

// NewClearCompleted returns the ClearCompleted use case.
func NewClearCompleted(d Deps) ClearCompleted { return &clearCompleted{d: d} }

// Execute implements ClearCompleted.
func (uc *clearCompleted) Execute(ctx context.Context, source string) (int, error) {
	l, err := uc.d.open(ctx, source)
	if err != nil {
		return 0, err
	}
	todos, err := l.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list todos: %w", err)
	}
	var ids []string
	for _, t := range todos {
		if t.Done() {
			ids = append(ids, t.ID())
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := l.repo.Delete(ctx, ids...); err != nil {
		return 0, fmt.Errorf("delete todos: %w", err)
	}
	return len(ids), nil
}

// ShareTodo changes who else may see a local to-do (input port, ADR 0018).
type ShareTodo interface {
	Execute(ctx context.Context, id, visibility string) (TodoView, error)
}

type shareTodo struct{ d Deps }

// NewShareTodo returns the ShareTodo use case.
func NewShareTodo(d Deps) ShareTodo { return &shareTodo{d: d} }

var errSharingUnavailable = kernel.NewPublicError(kernel.ErrNotFound, "Sharing is not available here.")

// Execute implements ShareTodo: the owner shares with only themselves,
// everyone, or one of their groups.
func (uc *shareTodo) Execute(ctx context.Context, id, visibility string) (TodoView, error) {
	if uc.d.Visibility == nil {
		return TodoView{}, errSharingUnavailable
	}
	p := kernel.PrincipalFrom(ctx)
	v, err := kernel.CheckShare(p, visibility)
	if err != nil {
		return TodoView{}, err
	}
	t, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return TodoView{}, err
	}
	if !t.Access().Mine(p) {
		return TodoView{}, kernel.ErrNotOwner
	}
	if err := uc.d.Visibility.SetVisibility(ctx, id, v); err != nil {
		return TodoView{}, err
	}
	a := t.Access()
	a.Visibility = v
	return uc.d.view(ctx, t.WithAccess(a)), nil
}
