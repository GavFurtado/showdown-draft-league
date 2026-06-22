package requests

import (
	"github.com/google/uuid"
)

type LeaguePokemonCreateRequestDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" binding:"required"`
	PokemonSpeciesID int64     `json:"PokemonSpeciesID" binding:"required"`
	Cost             *int      `json:"Cost" validate:"max=20"`
}

type LeaguePokemonUpdateRequestDTO struct {
	LeaguePokemonID uuid.UUID `json:"LeaguePokemonID" binding:"required"`
	Cost            *int      `json:"Cost" validate:"max=20"`
	IsAvailable     *bool     `json:"IsAvailable"`
}
