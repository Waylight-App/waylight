package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/appearance/domain"
	"waylight/internal/modules/appearance/usecase"
	"waylight/internal/shared/imaging"
)

// BackgroundStore implements usecase.BackgroundStore on SQLite (#277).
type BackgroundStore struct{ db *sql.DB }

// NewBackgroundStore returns the store.
func NewBackgroundStore(db *sql.DB) *BackgroundStore { return &BackgroundStore{db: db} }

// ListBackgrounds implements usecase.BackgroundStore, newest first.
func (s *BackgroundStore) ListBackgrounds(ctx context.Context) ([]domain.Background, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, width, height, size, created_at FROM appearance_backgrounds ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list backgrounds: %w", err)
	}
	defer rows.Close()
	out := []domain.Background{}
	for rows.Next() {
		var b domain.Background
		var created string
		if err := rows.Scan(&b.ID, &b.Width, &b.Height, &b.Size, &created); err != nil {
			return nil, fmt.Errorf("scan background: %w", err)
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, b)
	}
	return out, rows.Err()
}

// AddBackground implements usecase.BackgroundStore.
func (s *BackgroundStore) AddBackground(ctx context.Context, b domain.Background, image, thumb []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO appearance_backgrounds (id, width, height, size, created_at, image, thumb) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Width, b.Height, b.Size, b.CreatedAt.UTC().Format(time.RFC3339Nano), image, thumb)
	return err
}

// BackgroundImage implements usecase.BackgroundStore.
func (s *BackgroundStore) BackgroundImage(ctx context.Context, id string, thumb bool) ([]byte, error) {
	query := `SELECT image FROM appearance_backgrounds WHERE id = ?`
	if thumb {
		query = `SELECT thumb FROM appearance_backgrounds WHERE id = ?`
	}
	var data []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrBackgroundNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read background: %w", err)
	}
	return data, nil
}

// DeleteBackground implements usecase.BackgroundStore.
func (s *BackgroundStore) DeleteBackground(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM appearance_backgrounds WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete background: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrBackgroundNotFound
	}
	return nil
}

// ImageProcessor implements usecase.ImageProcessor with the standard
// library (internal/shared/imaging): upright, scaled down and re-encoded
// as JPEG, which drops EXIF data and neutralises malformed files.
type ImageProcessor struct{}

// Process implements usecase.ImageProcessor.
func (ImageProcessor) Process(data []byte, maxEdge, thumbEdge int) (usecase.ProcessedImage, error) {
	img, err := imaging.Decode(data)
	if errors.Is(err, imaging.ErrUnsupported) || errors.Is(err, imaging.ErrTooManyPixels) {
		return usecase.ProcessedImage{}, usecase.ErrUnsupportedImage
	}
	if err != nil {
		return usecase.ProcessedImage{}, err
	}
	full := imaging.Fit(img, maxEdge)
	image, err := imaging.EncodeJPEG(full, domain.BackgroundJPEGQuality)
	if err != nil {
		return usecase.ProcessedImage{}, err
	}
	thumb, err := imaging.EncodeJPEG(imaging.Fit(full, thumbEdge), domain.BackgroundJPEGQuality)
	if err != nil {
		return usecase.ProcessedImage{}, err
	}
	b := full.Bounds()
	return usecase.ProcessedImage{Image: image, Thumb: thumb, Width: b.Dx(), Height: b.Dy()}, nil
}
