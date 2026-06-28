package requests

import (
	"github.com/google/uuid"
)

type ClaimCreateRequestDTO struct {
	LeagueID  uuid.UUID `json:"LeagueID" binding:"required"`
	PlayerID  uuid.UUID `json:"PlayerID" binding:"required"`
	SpeciesID int64     `json:"SpeciesID" binding:"required"`
	Source    string    `json:"Source" binding:"required"`
	SourceID  *uuid.UUID `json:"SourceID"`
	CostPaid  int       `json:"CostPaid"`
}

type ClaimUpdateRequestDTO struct {
	ClaimID      uuid.UUID `json:"ClaimID" binding:"required"`
	IsActive     *bool     `json:"IsActive"`
	ReleasedWeek *int      `json:"ReleasedWeek"`
}
