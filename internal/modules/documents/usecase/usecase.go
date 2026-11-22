package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/documents/domain"
	"waylight/internal/shared/kernel"
)

// Documents is the input port of the module; every method is one use case.
type Documents interface {
	// List returns the visible documents: warranties by end date, the
	// soonest first, then documents without a warranty, ended ones last.
	List(ctx context.Context) ([]DocumentView, error)
	// Create adds a document; an empty visibility uses the default for new
	// records, private for members and admins.
	Create(ctx context.Context, in domain.Input, visibility string) (DocumentView, error)
	// Update changes a document of the principal; an empty visibility
	// keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (DocumentView, error)
	// Delete removes a document of the principal together with its file.
	Delete(ctx context.Context, id string) error
	// AttachFile stores the file of a document of the principal and
	// replaces an earlier one.
	AttachFile(ctx context.Context, id, name string, data []byte) (DocumentView, error)
	// File returns the file of a visible document.
	File(ctx context.Context, id string) (domain.Document, []byte, error)
}

// DocumentView is a document with its warranty state today.
type DocumentView struct {
	domain.Document
	Status domain.Status
	// Mine says whether the principal created the document.
	Mine bool
}

// Repository persists documents (output port, implemented in
// adapter/gateway). Reads return only what the principal in ctx may see;
// writes change only the principal's own documents (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Document, error)
	Get(ctx context.Context, id string) (domain.Document, error)
	// Save stores the fields a person enters; the file stays as it is.
	Save(ctx context.Context, d domain.Document) error
	// Delete removes the document and its file; an unknown one is no error,
	// one of someone else is kernel.ErrNotOwner.
	Delete(ctx context.Context, id string) error
	// SaveFile stores the file and its description in one step.
	SaveFile(ctx context.Context, id string, f domain.File, data []byte) error
	// File returns the bytes of the file, domain.ErrNoFile when none.
	File(ctx context.Context, id string) ([]byte, error)
}

// Deps bundles the output ports. Location decides what "today" is
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewDocuments returns the document use cases.
func NewDocuments(d Deps) Documents {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) view(ctx context.Context, d domain.Document) DocumentView {
	return DocumentView{Document: d, Status: d.StatusOn(u.d.Clock.Now().In(u.d.Location)), Mine: d.Access.Mine(kernel.PrincipalFrom(ctx))}
}

// rank orders the groups: covered, no warranty, ended.
func rank(s domain.Status) int {
	switch s {
	case domain.StatusNone:
		return 1
	case domain.StatusEnded:
		return 2
	}
	return 0
}

// List implements Documents.
func (u *uc) List(ctx context.Context) ([]DocumentView, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	out := make([]DocumentView, 0, len(list))
	for _, d := range list {
		out = append(out, u.view(ctx, d))
	}
	slices.SortStableFunc(out, func(a, b DocumentView) int {
		if ra, rb := rank(a.Status), rank(b.Status); ra != rb {
			return ra - rb
		}
		if a.Status == domain.StatusEnded {
			return b.WarrantyEnd.Compare(a.WarrantyEnd)
		}
		return a.WarrantyEnd.Compare(b.WarrantyEnd)
	})
	return out, nil
}

// Create implements Documents.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (DocumentView, error) {
	p := kernel.PrincipalFrom(ctx)
	access := kernel.NewAccess(p)
	if visibility != "" {
		v, err := kernel.CheckShare(p, visibility)
		if err != nil {
			return DocumentView{}, err
		}
		access.Visibility = v
	}
	d, err := domain.New(u.d.IDs.NewID(), in, access, u.d.Clock.Now())
	if err != nil {
		return DocumentView{}, err
	}
	if err := u.d.Repo.Save(ctx, d); err != nil {
		return DocumentView{}, fmt.Errorf("save document: %w", err)
	}
	return u.view(ctx, d), nil
}

// own loads a document the principal created.
func (u *uc) own(ctx context.Context, id string) (domain.Document, error) {
	d, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Document{}, err
	}
	if !d.Access.Mine(kernel.PrincipalFrom(ctx)) {
		return domain.Document{}, kernel.ErrNotOwner
	}
	return d, nil
}

// Update implements Documents.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (DocumentView, error) {
	d, err := u.own(ctx, id)
	if err != nil {
		return DocumentView{}, err
	}
	if visibility != "" {
		if d.Access.Visibility, err = kernel.CheckShare(kernel.PrincipalFrom(ctx), visibility); err != nil {
			return DocumentView{}, err
		}
	}
	if d, err = d.Update(in); err != nil {
		return DocumentView{}, err
	}
	if err := u.d.Repo.Save(ctx, d); err != nil {
		return DocumentView{}, err
	}
	return u.view(ctx, d), nil
}

// Delete implements Documents.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// AttachFile implements Documents.
func (u *uc) AttachFile(ctx context.Context, id, name string, data []byte) (DocumentView, error) {
	d, err := u.own(ctx, id)
	if err != nil {
		return DocumentView{}, err
	}
	f, err := domain.NewFile(name, data)
	if err != nil {
		return DocumentView{}, err
	}
	if err := u.d.Repo.SaveFile(ctx, id, f, data); err != nil {
		return DocumentView{}, fmt.Errorf("save file: %w", err)
	}
	d.File = f
	return u.view(ctx, d), nil
}

// File implements Documents.
func (u *uc) File(ctx context.Context, id string) (domain.Document, []byte, error) {
	d, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Document{}, nil, err
	}
	if d.File.Name == "" {
		return domain.Document{}, nil, domain.ErrNoFile
	}
	data, err := u.d.Repo.File(ctx, id)
	if err != nil {
		return domain.Document{}, nil, err
	}
	return d, data, nil
}
