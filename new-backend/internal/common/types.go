package common

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
)

// TODO: fix the fact that some of these fields are uncessarily pointers (i misunderstood omitempty)

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

// Request Structs
type LeagueRequest struct {
	Name                string              `json:"name" binding:"required"`
	RulesetDescription  string              `json:"ruleset_description"`
	MaxPokemonPerPlayer int                 `json:"max_pokemon_per_player" binding:"gte=1,max=12"`
	StartingDraftPoints int                 `json:"starting_draft_points" binding:"gte=20,max=150"`
	StartDate           time.Time           `json:"start_date" binding:"required"`
	EndDate             *time.Time          `json:"end_date" binding:"omitempty"`
	Format              models.LeagueFormat `json:"format"`
}

type UpdateProfileRequest struct {
	ShowdownName *string `json:"showdown_name" binding:"omitempty"`
}

// -- Player Related --
type PlayerCreateRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	LeagueID     uuid.UUID `json:"league_id" binding:"required"`
	InLeagueName *string   `json:"in_league_name" binding:"omitempty" validate:"min=3,max=20"`
	TeamName     *string   `json:"team_name" binding:"omitempty" validate:"min=3,max=20"`
}

type UpdatePlayerInfoRequest struct {
	InLeagueName  *string `json:"in_league_name,omitempty" validate:"min=3,max=20"`
	TeamName      *string `json:"team_name,omitempty" validate:"min=3,max=20"`
	Wins          *int    `json:"wins,omitempty" validate:"min=0"`
	Losses        *int    `json:"losses,omitempty" validate:"min=0"`
	DraftPoints   *int    `json:"draft_points,omitempty" validate:"min=0"`
	DraftPosition *int    `json:"draft_position,omitempty" validate:"min=0"`
}

// -- LeaguePokemon Related --
type LeaguePokemonCreateRequest struct {
	LeagueID         uuid.UUID `json:"league_id" binding:"required"`
	PokemonSpeciesID int64     `json:"pokemon_species_id" binding:"required"`
	Cost             *int      `json:"cost" binding:"required" validate:"max=20"`
}

type LeaguePokemonUpdateRequest struct {
	LeaguePokemonID uuid.UUID `json:"league_pokemon_id" binding:"required"`
	Cost            *int      `json:"cost,omitempty" validate:"max=20"`
	IsAvailable     bool      `json:"is_available,omitempty"`
}

// -- DraftedPokemon Related --
type DraftedPokemonCreateDTO struct {
	LeagueID         uuid.UUID `json:"league_id" binding:"required"`
	PlayerID         uuid.UUID `json:"player_id" binding:"required"`
	PokemonSpeciesID uuid.UUID `json:"pokemon_species_id" binding:"required"`
	DraftRoundNumber int       `json:"draft_round_number"`
	DraftPickNumber  int       `json:"draft_pick_number"`
	IsReleased       *bool     `json:"is_released,omitempty"`
}

type DraftedPokemonUpdateRequest struct {
	LeagueID         *uuid.UUID `json:"league_id,omitempty"`
	PlayerID         *uuid.UUID `json:"player_id,omitempty"`
	PokemonSpeciesID *uuid.UUID `json:"pokemon_species_id,omitempty"`
	DraftRoundNumber *int       `json:"draft_round_number,omitempty"`
	DraftPickNumber  *int       `json:"draft_pick_number,omitempty"`
	IsReleased       *bool      `json:"is_released,omitempty"`
}

// -- Webhook related --
// represents the structure for sending messages to Discord webhooks.
type DiscordWebhookPayload struct {
	Content   string                `json:"content,omitempty"`
	Username  string                `json:"username,omitempty"`
	AvatarURL string                `json:"avatar_url,omitempty"`
	Embeds    []DiscordWebhookEmbed `json:"embeds,omitempty"`
}

// represents an embed object within a Discord webhook payload.
type DiscordWebhookEmbed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color,omitempty"` // RGB color integer
}
