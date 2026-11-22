// Package presenter maps documents to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"waylight/internal/modules/documents/domain"
	"waylight/internal/modules/documents/usecase"
)

// DocumentDTO is one document in API responses; the file itself is
// downloaded from /api/documents/{id}/file.
type DocumentDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// PurchasedOn and WarrantyEnd are YYYY-MM-DD or empty.
	PurchasedOn string `json:"purchasedOn"`
	WarrantyEnd string `json:"warrantyEnd"`
	Note        string `json:"note"`
	// Status is "none", "running", "ending" or "ended".
	Status string `json:"status"`
	// FileName, ContentType and FileSize describe the attached file; the
	// name is empty when there is none.
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	FileSize    int64  `json:"fileSize"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer created the document and may change it.
	Mine bool `json:"mine"`
}

// Document maps one document.
func Document(v usecase.DocumentView) DocumentDTO {
	return DocumentDTO{ID: v.ID, Title: v.Title, PurchasedOn: domain.Day(v.PurchasedOn), WarrantyEnd: domain.Day(v.WarrantyEnd), Note: v.Note,
		Status: string(v.Status), FileName: v.File.Name, ContentType: v.File.ContentType, FileSize: v.File.Size,
		Visibility: string(v.Access.Visibility), Mine: v.Mine}
}

// Documents maps a list of documents.
func Documents(list []usecase.DocumentView) []DocumentDTO {
	out := make([]DocumentDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Document(v))
	}
	return out
}
