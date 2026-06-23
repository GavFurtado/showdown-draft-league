package requests

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
)

type LeagueCreateRequestDTO struct {
	Name                string             `json:"Name" binding:"required"`
	RulesetDescription  string             `json:"RulesetDescription"`
	MaxPokemonPerPlayer int                `json:"MaxPokemonPerPlayer" binding:"gte=1,max=20"`
	MinPokemonPerPlayer int                `json:"MinPokemonPerPlayer" binding:"gte=0,max=20"`
	StartingDraftPoints int                `json:"StartingDraftPoints" binding:"gte=20,max=150"`
	Format              types.LeagueFormat `json:"Format"`
}
