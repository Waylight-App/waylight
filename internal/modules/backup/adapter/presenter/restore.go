package presenter

import (
	"errors"

	"waylight/internal/modules/backup/usecase"
	"waylight/internal/shared/kernel"
)

// RestoredDTO answers POST /api/backups/restore (#677).
type RestoredDTO struct {
	Status     string  `json:"status"`
	SafetyCopy FileDTO `json:"safetyCopy"`
}

// Restored maps a restore result.
func Restored(r usecase.RestoreResult) RestoredDTO {
	return RestoredDTO{Status: "restarting", SafetyCopy: File(r.SafetyCopy)}
}

// ProblemDTO is why a backup cannot be restored, as a translatable message.
type ProblemDTO struct {
	Message    string            `json:"message"`
	MessageKey string            `json:"messageKey"`
	Params     map[string]string `json:"params,omitempty"`
}

// PreviewDTO answers GET /api/backups/{name}/preview (#677).
type PreviewDTO struct {
	File       FileDTO     `json:"file"`
	Restorable bool        `json:"restorable"`
	Problem    *ProblemDTO `json:"problem"`
	Modules    int         `json:"modules"`
	Tabs       int         `json:"tabs"`
	Widgets    int         `json:"widgets"`
	Updates    int         `json:"updates"`
}

// Preview maps a preview view.
func Preview(v usecase.PreviewView) PreviewDTO {
	dto := PreviewDTO{File: File(v.File), Restorable: v.Problem == nil,
		Modules: v.Modules, Tabs: v.Tabs, Widgets: v.Widgets, Updates: v.Updates}
	if v.Problem != nil {
		dto.Problem = &ProblemDTO{Message: v.Problem.Error()}
		var perr *kernel.PublicError
		if errors.As(v.Problem, &perr) {
			dto.Problem = &ProblemDTO{Message: perr.Message, MessageKey: kernel.MessageKey(perr.Template), Params: perr.Params}
		}
	}
	return dto
}
