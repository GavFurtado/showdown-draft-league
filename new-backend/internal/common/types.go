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

type PlayerRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	LeagueID     uuid.UUID `json:"league_id" binding:"required"`
	InLeagueName *string   `json:"in_league_name" binding:"omitempty" validate:"min=3,max=20"`
	TeamName     *string   `json:"team_name" binding:"omitempty" validate:"min=3,max=20"`
}

type UpdatePlayerInfoRequest struct {
	InLeagueName  *string `json:"in_league_name" binding:"omitempty" validate:"min=3,max=20"`
	TeamName      *string `json:"team_name" binding:"omitempty" validate:"min=3,max=20"`
	Wins          *int    `json:"wins" binding:"omitempty" validate:"min=0"`
	Losses        *int    `json:"losses" binding:"omitempty" validate:"min=0"`
	DraftPoints   *int    `json:"draft_points" binding:"omitempty" validate:"min=0"`
	DraftPosition *int    `json:"draft_position" binding:"omitempty" validate:"min=0"`
}
