package requests

import (
	"github.com/google/uuid"
)

type DraftedPokemonCreateRequestDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" binding:"required"`
	PlayerID         uuid.UUID `json:"PlayerID" binding:"required"`
	PokemonSpeciesID uuid.UUID `json:"PokemonSpeciesID" binding:"required"`
	DraftRoundNumber int       `json:"DraftRoundNumber"`
	DraftPickNumber  int       `json:"DraftPickNumber"`
	IsReleased       *bool     `json:"IsReleased"`
}

type DraftedPokemonUpdateRequestDTO struct {
	LeagueID         *uuid.UUID `json:"LeagueID"`
	PlayerID         *uuid.UUID `json:"PlayerID"`
	PokemonSpeciesID *uuid.UUID `json:"PokemonSpeciesID"`
	DraftRoundNumber *int       `json:"DraftRoundNumber"`
	DraftPickNumber  *int       `json:"DraftPickNumber"`
	IsReleased       *bool      `json:"IsReleased"`
}
