package requests

import (
	"github.com/google/uuid"
)

type LeagueMemberCreateRequestDTO struct {
	UserID       uuid.UUID `json:"UserID" binding:"required"`
	LeagueID     uuid.UUID `json:"LeagueID" binding:"required"`
	InLeagueName *string   `json:"InLeagueName" binding:"omitempty" validate:"min=3,max=20"`
	TeamName     *string   `json:"TeamName" binding:"omitempty" validate:"min=3,max=20"`
}

type UpdateLeagueMemberInfoRequestDTO struct {
	InLeagueName  *string `json:"InLeagueName" validate:"min=3,max=20"`
	TeamName      *string `json:"TeamName" validate:"min=3,max=20"`
	Wins          *int    `json:"Wins" validate:"min=0"`
	Losses        *int    `json:"Losses" validate:"min=0"`
	DraftPoints   *int    `json:"DraftPoints" validate:"min=0"`
	DraftPosition *int    `json:"DraftPosition" validate:"min=0"`
}
