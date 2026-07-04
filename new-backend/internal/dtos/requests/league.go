package requests

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
)

type LeagueCreateRequestDTO struct {
	Name                string                 `json:"Name" validate:"required"`
	RulesetDescription  string                 `json:"RulesetDescription"`
	MaxPokemonPerPlayer int                    `json:"MaxPokemonPerPlayer" validate:"gte=1,max=20"`
	MinPokemonPerPlayer int                    `json:"MinPokemonPerPlayer" validate:"gte=0,max=20"`
	StartingDraftPoints int                    `json:"StartingDraftPoints" validate:"gte=20,max=1000"`
	MaxPlayers          int                    `json:"MaxPlayers" validate:"gte=0,max=20"`
	Visibility          enums.LeagueVisibility `json:"Visibility" validate:"isValid"`
	Format              types.LeagueFormat     `json:"Format"`
}
