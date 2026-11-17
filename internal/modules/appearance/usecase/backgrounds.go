package usecase

import (
	"context"
	"errors"
	"fmt"

	"waylight/internal/modules/appearance/domain"
	"waylight/internal/shared/kernel"
)

// ProcessedImage is an upload turned into what is stored.
type ProcessedImage struct {
	Image, Thumb  []byte
	Width, Height int
}

// ImageProcessor decodes, turns upright and scales an upload (output port).
type ImageProcessor interface {
	Process(data []byte, maxEdge, thumbEdge int) (ProcessedImage, error)
}

// ErrUnsupportedImage is returned by an ImageProcessor for data it cannot read.
var ErrUnsupportedImage = errors.New("unsupported image")

// BackgroundStore keeps uploaded backgrounds (output port).
type BackgroundStore interface {
	ListBackgrounds(ctx context.Context) ([]domain.Background, error)
	AddBackground(ctx context.Context, b domain.Background, image, thumb []byte) error
	BackgroundImage(ctx context.Context, id string, thumb bool) ([]byte, error)
	DeleteBackground(ctx context.Context, id string) error
}

// Backgrounds manages uploaded backgrounds and sets the global one (input port).
type Backgrounds interface {
	List(ctx context.Context) ([]domain.Background, error)
	Upload(ctx context.Context, data []byte) (domain.Background, error)
	Image(ctx context.Context, id string, thumb bool) ([]byte, error)
	Delete(ctx context.Context, id string) error
	// Set changes only the global background, so the dashboard can do it
	// without sending every other preference along.
	Set(ctx context.Context, value string) (AppearanceResponse, error)
}

// BackgroundDeps bundles the ports.
type BackgroundDeps struct {
	Store     BackgroundStore
	Processor ImageProcessor
	Repo      Repository
	Clock     kernel.Clock
	IDs       kernel.IDGenerator
}

type backgrounds struct{ d BackgroundDeps }

// NewBackgrounds returns the Backgrounds use case.
func NewBackgrounds(d BackgroundDeps) Backgrounds { return &backgrounds{d: d} }

func (uc *backgrounds) List(ctx context.Context) ([]domain.Background, error) {
	return uc.d.Store.ListBackgrounds(ctx)
}

func (uc *backgrounds) Upload(ctx context.Context, data []byte) (domain.Background, error) {
	if len(data) > domain.MaxBackgroundUploadBytes {
		return domain.Background{}, domain.ErrBackgroundTooLarge
	}
	existing, err := uc.d.Store.ListBackgrounds(ctx)
	if err != nil {
		return domain.Background{}, fmt.Errorf("list backgrounds: %w", err)
	}
	if len(existing) >= domain.MaxBackgrounds {
		return domain.Background{}, domain.ErrTooManyBackgrounds
	}
	out, err := uc.d.Processor.Process(data, domain.MaxBackgroundEdge, domain.BackgroundThumbEdge)
	if errors.Is(err, ErrUnsupportedImage) {
		return domain.Background{}, domain.ErrBackgroundUnsupported
	}
	if err != nil {
		return domain.Background{}, fmt.Errorf("process background: %w", err)
	}
	b := domain.Background{ID: uc.d.IDs.NewID(), Width: out.Width, Height: out.Height,
		Size: int64(len(out.Image) + len(out.Thumb)), CreatedAt: uc.d.Clock.Now()}
	if err := uc.d.Store.AddBackground(ctx, b, out.Image, out.Thumb); err != nil {
		return domain.Background{}, fmt.Errorf("store background: %w", err)
	}
	return b, nil
}

func (uc *backgrounds) Image(ctx context.Context, id string, thumb bool) ([]byte, error) {
	return uc.d.Store.BackgroundImage(ctx, id, thumb)
}

// Delete removes a picture; if it is the global background, that falls
// back to the default, so the page never points at a missing image.
func (uc *backgrounds) Delete(ctx context.Context, id string) error {
	if err := uc.d.Store.DeleteBackground(ctx, id); err != nil {
		return err
	}
	a, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return fmt.Errorf("load appearance: %w", err)
	}
	if a.BackgroundImage() == (domain.Background{ID: id}).URL() {
		_ = a.SetBackgroundImage("")
		if err := uc.d.Repo.Save(ctx, a); err != nil {
			return fmt.Errorf("save appearance: %w", err)
		}
	}
	return nil
}

func (uc *backgrounds) Set(ctx context.Context, value string) (AppearanceResponse, error) {
	a, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("load appearance: %w", err)
	}
	if err := a.SetBackgroundImage(value); err != nil {
		return AppearanceResponse{}, err
	}
	if err := uc.d.Repo.Save(ctx, a); err != nil {
		return AppearanceResponse{}, fmt.Errorf("save appearance: %w", err)
	}
	return toResponse(a), nil
}
