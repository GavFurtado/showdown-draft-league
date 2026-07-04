package requests

import (
	"github.com/google/uuid"
)

type ReportGameRequestDTO struct {
	ReporterID  uuid.UUID `json:"ReporterID" validate:"omitempty"`
	WinnerID    uuid.UUID `json:"WinnerID" validate:"required"`
	Player1Wins *int      `json:"Player1Wins" validate:"required,gte=0"`
	Player2Wins *int      `json:"Player2Wins" validate:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" validate:"dive,url"`
}

type FinalizeGameRequestDTO struct {
	FinalizerID uuid.UUID `json:"FinalizerID" validate:"required"`
	WinnerID    uuid.UUID `json:"WinnerID" validate:"required"`
	Player1Wins *int      `json:"Player1Wins" validate:"required,gte=0"`
	Player2Wins *int      `json:"Player2Wins" validate:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" validate:"dive,url"`
}
