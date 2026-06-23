package requests

import (
	"github.com/google/uuid"
)

type ReportGameRequestDTO struct {
	ReporterID  uuid.UUID `json:"ReporterID" binding:"omitempty"`
	WinnerID    uuid.UUID `json:"WinnerID" binding:"required"`
	Player1Wins *int      `json:"Player1Wins" binding:"required,gte=0"`
	Player2Wins *int      `json:"Player2Wins" binding:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" binding:"dive,url"`
}

type FinalizeGameRequestDTO struct {
	FinalizerID uuid.UUID `json:"FinalizerID" binding:"required"`
	WinnerID    uuid.UUID `json:"WinnerID" binding:"required"`
	Player1Wins *int      `json:"Player1Wins" binding:"required,gte=0"`
	Player2Wins *int      `json:"Player2Wins" binding:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" binding:"dive,url"`
}
