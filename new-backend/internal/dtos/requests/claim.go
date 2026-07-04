package requests

import (
	"github.com/google/uuid"
)

type ClaimCreateRequestDTO struct {
	LeagueID  uuid.UUID  `json:"LeagueID" validate:"required"`
	PlayerID  uuid.UUID  `json:"PlayerID" validate:"required"`
	SpeciesID int64      `json:"SpeciesID" validate:"required"`
	Source    string     `json:"Source" validate:"required"`
	SourceID  *uuid.UUID `json:"SourceID"`
	CostPaid  int        `json:"CostPaid"`
}

type ClaimUpdateRequestDTO struct {
	ClaimID      uuid.UUID `json:"ClaimID" validate:"required"`
	IsActive     *bool     `json:"IsActive"`
	ReleasedWeek *int      `json:"ReleasedWeek"`
}
