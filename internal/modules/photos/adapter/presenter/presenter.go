// Package presenter maps pictures to JSON (#286).
//
// Ring: Interface Adapters.
package presenter

import (
	"waylight/internal/modules/photos/domain"
	"waylight/internal/modules/photos/usecase"
)

// PhotoDTO is one picture; the image itself is at /api/photos/{id}/image.
type PhotoDTO struct {
	ID     string `json:"id"`
	Album  string `json:"album"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int64  `json:"size"`
}

// AlbumDTO summarises an album.
type AlbumDTO struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Bytes int64  `json:"bytes"`
}

// OverviewDTO lists the albums with the storage used and its limits.
type OverviewDTO struct {
	Albums     []AlbumDTO `json:"albums"`
	Count      int        `json:"count"`
	UsedBytes  int64      `json:"usedBytes"`
	QuotaBytes int64      `json:"quotaBytes"`
	MaxPhotos  int        `json:"maxPhotos"`
	MaxUpload  int64      `json:"maxUploadBytes"`
}

// Photo maps one picture.
func Photo(p domain.Photo) PhotoDTO {
	return PhotoDTO{ID: p.ID, Album: p.Album, Width: p.Width, Height: p.Height, Size: p.Size}
}

// Photos maps a list.
func Photos(list []domain.Photo) []PhotoDTO {
	out := make([]PhotoDTO, 0, len(list))
	for _, p := range list {
		out = append(out, Photo(p))
	}
	return out
}

// Overview maps the album overview.
func Overview(o usecase.Overview) OverviewDTO {
	albums := make([]AlbumDTO, 0, len(o.Albums))
	for _, a := range o.Albums {
		albums = append(albums, AlbumDTO(a))
	}
	return OverviewDTO{Albums: albums, Count: o.Usage.Count, UsedBytes: o.Usage.Bytes,
		QuotaBytes: domain.QuotaBytes, MaxPhotos: domain.MaxPhotos, MaxUpload: domain.MaxUploadBytes}
}
