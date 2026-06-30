package requests

import (
	"github.com/google/uuid"
)

type PoolEntryCreateRequestDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" binding:"required"`
	PokemonSpeciesID int64     `json:"PokemonSpeciesID" binding:"required"`
	Cost             *int      `json:"Cost" validate:"max=20"`
}

type PoolEntryUpdateRequestDTO struct {
	PoolEntryID uuid.UUID `json:"PoolEntryID" binding:"required"`
	Cost        *int      `json:"Cost" validate:"max=20"`
	IsAvailable *bool     `json:"IsAvailable"`
}
