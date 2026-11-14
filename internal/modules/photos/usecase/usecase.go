// Package usecase holds the photo frame use cases (#286).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"

	"trisse-home/internal/modules/photos/domain"
	"trisse-home/internal/shared/kernel"
)

// Repository stores pictures (output port).
type Repository interface {
	List(ctx context.Context, album string) ([]domain.Photo, error)
	Albums(ctx context.Context) ([]domain.Album, error)
	Usage(ctx context.Context) (domain.Usage, error)
	// Add stores a picture at the end of its album.
	Add(ctx context.Context, p domain.Photo, image, thumb []byte) (domain.Photo, error)
	Delete(ctx context.Context, id string) error
	// Reorder sets the positions of an album's pictures to the order of ids.
	Reorder(ctx context.Context, album string, ids []string) error
	// Image returns the stored JPEG or its thumbnail.
	Image(ctx context.Context, id string, thumb bool) ([]byte, error)
}

// Processed is an upload turned into what is stored.
type Processed struct {
	Image, Thumb  []byte
	Width, Height int
}

// Processor decodes, turns upright and scales an upload (output port).
type Processor interface {
	Process(data []byte) (Processed, error)
}

// ErrUnsupportedImage is returned by a Processor for data it cannot read.
var ErrUnsupportedImage = errors.New("unsupported image")

// Overview is the album list with the storage used.
type Overview struct {
	Albums []domain.Album
	Usage  domain.Usage
}

// Photos is the input port.
type Photos interface {
	List(ctx context.Context, album string) ([]domain.Photo, error)
	Overview(ctx context.Context) (Overview, error)
	Upload(ctx context.Context, album string, data []byte) (domain.Photo, error)
	Reorder(ctx context.Context, album string, ids []string) error
	Delete(ctx context.Context, id string) error
	Image(ctx context.Context, id string, thumb bool) ([]byte, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo      Repository
	Processor Processor
	Clock     kernel.Clock
	IDs       kernel.IDGenerator
}

type photos struct{ d Deps }

// NewPhotos returns the Photos use case.
func NewPhotos(d Deps) Photos { return &photos{d: d} }

func (uc *photos) List(ctx context.Context, album string) ([]domain.Photo, error) {
	album, err := domain.NormaliseAlbum(album)
	if err != nil {
		return nil, err
	}
	return uc.d.Repo.List(ctx, album)
}

func (uc *photos) Overview(ctx context.Context) (Overview, error) {
	albums, err := uc.d.Repo.Albums(ctx)
	if err != nil {
		return Overview{}, fmt.Errorf("list albums: %w", err)
	}
	u, err := uc.d.Repo.Usage(ctx)
	if err != nil {
		return Overview{}, fmt.Errorf("storage usage: %w", err)
	}
	return Overview{Albums: albums, Usage: u}, nil
}

func (uc *photos) Upload(ctx context.Context, album string, data []byte) (domain.Photo, error) {
	album, err := domain.NormaliseAlbum(album)
	if err != nil {
		return domain.Photo{}, err
	}
	if len(data) > domain.MaxUploadBytes {
		return domain.Photo{}, domain.ErrTooLarge
	}
	u, err := uc.d.Repo.Usage(ctx)
	if err != nil {
		return domain.Photo{}, fmt.Errorf("storage usage: %w", err)
	}
	// Checked before the expensive decode and again with the real size.
	if !u.Fits(0) {
		return domain.Photo{}, domain.ErrQuotaFull
	}
	out, err := uc.d.Processor.Process(data)
	if errors.Is(err, ErrUnsupportedImage) {
		return domain.Photo{}, domain.ErrUnsupported
	}
	if err != nil {
		return domain.Photo{}, fmt.Errorf("process picture: %w", err)
	}
	size := int64(len(out.Image) + len(out.Thumb))
	if !u.Fits(size) {
		return domain.Photo{}, domain.ErrQuotaFull
	}
	p := domain.Photo{ID: uc.d.IDs.NewID(), Album: album, Width: out.Width, Height: out.Height, Size: size, CreatedAt: uc.d.Clock.Now()}
	return uc.d.Repo.Add(ctx, p, out.Image, out.Thumb)
}

func (uc *photos) Reorder(ctx context.Context, album string, ids []string) error {
	album, err := domain.NormaliseAlbum(album)
	if err != nil {
		return err
	}
	list, err := uc.d.Repo.List(ctx, album)
	if err != nil {
		return err
	}
	if err := domain.CheckOrder(list, ids); err != nil {
		return err
	}
	return uc.d.Repo.Reorder(ctx, album, ids)
}

func (uc *photos) Delete(ctx context.Context, id string) error {
	return uc.d.Repo.Delete(ctx, id)
}

func (uc *photos) Image(ctx context.Context, id string, thumb bool) ([]byte, error) {
	return uc.d.Repo.Image(ctx, id, thumb)
}
