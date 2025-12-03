package common

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
)

// TODO: fix the fact that some of these fields are uncessarily pointers (i misunderstood omitempty)
type DiscordUser struct {
	ID            string `json:"ID" gorm:"column:id"`
	Username      string `json:"Username" gorm:"column:username"`
	Discriminator string `json:"Discriminator" gorm:"column:discriminator"`
	Avatar        string `json:"Avatar" gorm:"column:avatar"`
}

// Request Structs
type LeagueCreateRequestDTO struct {
	Name                string              `json:"Name" binding:"required" gorm:"column:name"`
	RulesetDescription  string              `json:"RulesetDescription" gorm:"column:ruleset_description"`
	MaxPokemonPerPlayer int                 `json:"MaxPokemonPerPlayer" binding:"gte=1,max=20" gorm:"column:max_pokemon_per_player"`
	MinPokemonPerPlayer int                 `json:"MinPokemonPerPlayer" binding:"gte=0,max=20" gorm:"column:min_pokemon_per_player"`
	StartingDraftPoints int                 `json:"StartingDraftPoints" binding:"gte=20,max=150" gorm:"column:starting_draft_points"`
	StartDate           time.Time           `json:"StartDate" binding:"required" gorm:"column:start_date"`
	Format              models.LeagueFormat `json:"Format" gorm:"column:format"`
}

// -- Player Related --
type PlayerCreateRequest struct {
	UserID       uuid.UUID `json:"UserID" binding:"required" gorm:"column:user_id"`
	LeagueID     uuid.UUID `json:"LeagueID" binding:"required" gorm:"column:league_id"`
	InLeagueName *string   `json:"InLeagueName" binding:"omitempty" validate:"min=3,max=20" gorm:"column:in_league_name"`
	TeamName     *string   `json:"TeamName" binding:"omitempty" validate:"min=3,max=20" gorm:"column:team_name"`
}

type UserUpdateProfileRequest struct {
	ShowdownName *string `json:"ShowdownName" binding:"omitempty" gorm:"column:showdown_name"`
}

type UpdatePlayerInfoRequest struct {
	InLeagueName  *string `json:"InLeagueName" validate:"min=3,max=20" gorm:"column:in_league_name"`
	TeamName      *string `json:"TeamName" validate:"min=3,max=20" gorm:"column:team_name"`
	Wins          *int    `json:"Wins" validate:"min=0" gorm:"column:wins"`
	Losses        *int    `json:"Losses" validate:"min=0" gorm:"column:losses"`
	DraftPoints   *int    `json:"DraftPoints" validate:"min=0" gorm:"column:draft_points"`
	DraftPosition *int    `json:"DraftPosition" validate:"min=0" gorm:"column:draft_position"`
}

// -- LeaguePokemon Related --
type LeaguePokemonCreateRequestDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" binding:"required" gorm:"column:league_id"`
	PokemonSpeciesID int64     `json:"PokemonSpeciesID" binding:"required" gorm:"column:pokemon_species_id"`
	Cost             *int      `json:"Cost" validate:"max=20" gorm:"column:cost"`
}

type LeaguePokemonUpdateRequest struct {
	LeaguePokemonID uuid.UUID `json:"LeaguePokemonID" binding:"required" gorm:"column:league_pokemon_id"`
	Cost            *int      `json:"Cost" validate:"max=20" gorm:"column:cost"`
	IsAvailable     *bool     `json:"IsAvailable" gorm:"column:is_available"`
}

// -- Draft Related --
type DraftMakePickDTO struct {
	RequestedPickCount int             `json:"RequestedPickCount" binding:"required" gorm:"column:requested_pick_count"`
	RequestedPicks     []RequestedPick `json:"RequestedPicks" binding:"required" gorm:"column:requested_picks"`
}

type RequestedPick struct {
	LeaguePokemonID uuid.UUID `json:"LeaguePokemonID" binding:"required" gorm:"column:league_pokemon_id"`
	DraftPickNumber int       `json:"DraftPickNumber" binding:"required" gorm:"column:draft_pick_number"`
}

// -- DraftedPokemon Related --
type DraftedPokemonCreateDTO struct {
	LeagueID         uuid.UUID `json:"LeagueID" binding:"required" gorm:"column:league_id"`
	PlayerID         uuid.UUID `json:"PlayerID" binding:"required" gorm:"column:player_id"`
	PokemonSpeciesID uuid.UUID `json:"PokemonSpeciesID" binding:"required" gorm:"column:pokemon_species_id"`
	DraftRoundNumber int       `json:"DraftRoundNumber" gorm:"column:draft_round_number"`
	DraftPickNumber  int       `json:"DraftPickNumber" gorm:"column:draft_pick_number"`
	IsReleased       *bool     `json:"IsReleased" gorm:"column:is_released"`
}

type DraftedPokemonUpdateRequest struct {
	LeagueID         *uuid.UUID `json:"LeagueID" gorm:"column:league_id"`
	PlayerID         *uuid.UUID `json:"PlayerID" gorm:"column:player_id"`
	PokemonSpeciesID *uuid.UUID `json:"PokemonSpeciesID" gorm:"column:pokemon_species_id"`
	DraftRoundNumber *int       `json:"DraftRoundNumber" gorm:"column:draft_round_number"`
	DraftPickNumber  *int       `json:"DraftPickNumber" gorm:"column:draft_pick_number"`
	IsReleased       *bool      `json:"IsReleased" gorm:"column:is_released"`
}

// -- Webhook related --
// represents the structure for sending messages to Discord webhooks.
type DiscordWebhookPayload struct {
	Content   string                `json:"Content" gorm:"column:content"`
	Username  string                `json:"Username" gorm:"column:username"`
	AvatarURL string                `json:"AvatarURL" gorm:"column:avatar_url"`
	Embeds    []DiscordWebhookEmbed `json:"Embeds" gorm:"column:embeds"`
}

// represents an embed object within a Discord webhook payload.
type DiscordWebhookEmbed struct {
	Title       string `json:"Title" gorm:"column:title"`
	Description string `json:"Description" gorm:"column:description"`
	Color       int    `json:"Color"` // RGB color integer
}

// PokemonSpeciesListDTO represents a simplified view of a PokemonSpecies for list displays.
type PokemonSpeciesListDTO struct {
	ID           int64    `json:"ID" gorm:"column:id"`
	Name         string   `json:"Name" gorm:"column:name"`
	Types        []string `json:"Types" gorm:"column:types"`
	FrontDefault string   `json:"FrontDefault" gorm:"column:front_default"`
}

// -- Game Related --

type ReportGameDTO struct {
	ReporterID  uuid.UUID `json:"ReporterID" binding:"required"`
	WinnerID    uuid.UUID `json:"WinnerID" binding:"required"`
	Player1Wins int       `json:"Player1Wins" binding:"required,gte=0"`
	Player2Wins int       `json:"Player2Wins" binding:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" binding:"dive,url"`
}

type FinalizeGameDTO struct {
	FinalizerID uuid.UUID `json:"FinalizerID" binding:"required"`
	WinnerID    uuid.UUID `json:"WinnerID" binding:"required"`
	Player1Wins int       `json:"Player1Wins" binding:"required,gte=0"`
	Player2Wins int       `json:"Player2Wins" binding:"required,gte=0"`
	ReplayLinks []string  `json:"ReplayLinks" binding:"dive,url"`
}