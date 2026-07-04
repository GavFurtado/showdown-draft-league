package requests

import (
	"github.com/google/uuid"
)

type PoolEntryCreateRequestDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" validate:"required"`
	PokemonSpeciesID int64     `json:"PokemonSpeciesID" validate:"required"`
	Cost             *int      `json:"Cost" validate:"max=20"`
}

type PoolEntryUpdateRequestDTO struct {
	PoolEntryID uuid.UUID `json:"PoolEntryID" validate:"required"`
	Cost        *int      `json:"Cost" validate:"max=20"`
	IsAvailable *bool     `json:"IsAvailable"`
}
