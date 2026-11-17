package controller

import (
	"mime"
	"net/http"
	"strings"

	"waylight/internal/modules/shopping/adapter/presenter"
	"waylight/internal/modules/shopping/usecase"
	"waylight/internal/shared/httpapi"
)

// ListHandlers serve the lists kept on waylight (#471).
type ListHandlers struct{ uc *usecase.Lists }

// NewListHandlers returns the handlers.
func NewListHandlers(uc *usecase.Lists) *ListHandlers { return &ListHandlers{uc: uc} }

// All handles GET /api/shopping/lists.
func (h *ListHandlers) All(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.All(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Lists(v))
}

// Create handles POST /api/shopping/lists.
func (h *ListHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto ListRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Create(r.Context(), dto.Name, dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.List(v))
}

// Update handles PUT /api/shopping/lists/{id}.
func (h *ListHandlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ListRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.Update(r.Context(), r.PathValue("id"), dto.Name, dto.Visibility)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(v))
}

// Delete handles DELETE /api/shopping/lists/{id}.
func (h *ListHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetCurrent handles PUT /api/shopping/current.
func (h *ListHandlers) SetCurrent(w http.ResponseWriter, r *http.Request) {
	var dto CurrentRequestDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.SetCurrent(r.Context(), dto.List); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Items handles GET /api/shopping/lists/{id}/items.
func (h *ListHandlers) Items(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.Items(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Items(items))
}

// Add handles POST /api/shopping/lists/{id}/items.
func (h *ListHandlers) Add(w http.ResponseWriter, r *http.Request) {
	var dto NewItemDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	it, err := h.uc.Add(r.Context(), r.PathValue("id"), dto.Summary)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Item(it))
}

// SetDone handles PUT /api/shopping/lists/{id}/items/{itemId}.
func (h *ListHandlers) SetDone(w http.ResponseWriter, r *http.Request) {
	var dto ItemStateDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.SetDone(r.Context(), r.PathValue("id"), r.PathValue("itemId"), dto.Done); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteItem handles DELETE /api/shopping/lists/{id}/items/{itemId}.
func (h *ListHandlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteItem(r.Context(), r.PathValue("id"), r.PathValue("itemId")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearCompleted handles POST /api/shopping/lists/{id}/clear-completed.
func (h *ListHandlers) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.ClearCompleted(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ClearedDTO{Deleted: n})
}

// Import handles POST /api/shopping/lists/{id}/import.
func (h *ListHandlers) Import(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.Import(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ImportedDTO{Imported: n})
}

// Export handles GET /api/shopping/lists/{id}/export?format=text|ics.
func (h *ListHandlers) Export(w http.ResponseWriter, r *http.Request) {
	e, err := h.uc.Export(r.Context(), r.PathValue("id"), r.URL.Query().Get("format"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", e.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fileName(e.Name)}))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(e.Body))
}

// fileName keeps a list name usable as a file name on every system.
func fileName(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < ' ' {
			return '-'
		}
		return r
	}, s)
	if strings.Trim(s, ". -") == "" {
		return "shopping-list"
	}
	return s
}
