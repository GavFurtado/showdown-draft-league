package common

import (
	"time"

	"github.com/google/uuid"
)

// represents the basic information retrieved from Discord's /users/@me endpoint.
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

// Response Structs
type LeagueRequest struct {
	Name                  string     `json:"name" binding:"required"`
	RulesetID             *uuid.UUID `json:"ruleset_id"`
	MaxPokemonPerPlayer   uint       `json:"max_pokemon_per_player" binding:"gte=1, max=12"`
	StartingDraftPoints   uint       `json:"starting_draft_points" binding:"gte=20, max=150"`
	AllowWeeklyFreeAgents bool       `json:"allow_free_agents"`
	StartDate             time.Time  `json:"start_date" binding:"required,datetime=02/01/2006"`
	EndDate               *time.Time `json:"end_date" binding:"omitempty,datetime=02/01/2006"`
}

type UpdateProfileRequest struct {
	ShowdownName string `json:"showdown_name" binding:"omitempty"`
}
