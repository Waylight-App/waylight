package modules

// DataKind says who owns the data a module stores (ADR 0018). It is part of
// the Module interface, so a new module does not build until it decides.
type DataKind string

const (
	// InstanceData belongs to the instance: services, Docker, backups and
	// the like. Admins manage it; other accounts read it as their role
	// allows.
	InstanceData DataKind = "instance"
	// PersonalSettings are one row per account: theme, search shortcuts,
	// connected calendar or mail accounts.
	PersonalSettings DataKind = "personal-settings"
	// PersonalContent is created by an account and carries an owner and a
	// visibility per record: notes, to-dos, bookmarks. Its repositories
	// filter every read with sqlutil.VisibleTo once their tables have the
	// owner_id column; internal/archtest enforces that.
	PersonalContent DataKind = "personal-content"
)
