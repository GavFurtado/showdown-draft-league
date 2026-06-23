package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// League defines the structure of a League
type League struct {
	ID                     uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	Name                   string              `gorm:"not null;column:name;uniqueIndex:idx_owner_league_name" json:"Name"`
	OwnerUserID            uuid.UUID           `gorm:"type:uuid;not null;column:owner_user_id;uniqueIndex:idx_owner_league_name" json:"OwnerUserID"`
	StartDate              time.Time           `gorm:"not null;column:start_date" json:"StartDate"`
	EndDate                *time.Time          `gorm:"column:end_date" json:"EndDate"` // this is set when the league is cancelled or actualy ends, nil otherwise
	RulesetDescription     string              `gorm:"type:text;column:ruleset_description" json:"RulesetDescription"`
	Status                 enums.LeagueStatus  `gorm:"type:varchar(50);not null;default:'pending';column:status" json:"Status"`
	PlayerCount            int                 `gorm:"column:player_count" json:"PlayerCount"`
	MaxPokemonPerPlayer    int                 `gorm:"not null;default:0;column:max_pokemon_per_player" json:"MaxPokemonPerPlayer"`
	MinPokemonPerPlayer    int                 `gorm:"not null;default:0;column:min_pokemon_per_player" json:"MinPokemonPerPlayer"`
	MaxPlayers             int                 `gorm:"default:0;column:max_players" json:"MaxPlayers"` // 0 = unlimited
	StartingDraftPoints    int                 `gorm:"not null;default:140;column:starting_draft_points" json:"StartingDraftPoints"`
	Format                 *types.LeagueFormat `gorm:"type:jsonb;column:format" json:"Format,omitempty"`
	CreatedAt              *time.Time          `gorm:"type:timestamp with time zone;column:created_at" json:"CreatedAt"`
	UpdatedAt              *time.Time          `gorm:"type:timestamp with time zone;column:updated_at" json:"UpdatedAt"`
	DeletedAt              gorm.DeletedAt      `gorm:"index;column:deleted_at" json:"-"`
	DiscordWebhookURL      *string             `gorm:"column:discord_webhook_url" json:"DiscordWebhookURL"`
	CurrentWeekNumber      int                 `gorm:"not null;default:0;column:current_week_number" json:"CurrentWeekNumber"` // starts with 1, 0 is invalid and used when not relevant
	NextWeeklyTick         *time.Time          `gorm:"type:timestamp with time zone;column:next_weekly_tick" json:"NextWeeklyTick"`
	RegularSeasonStartDate *time.Time          `gorm:"type:timestamp with time zone;column:regular_season_start_date" json:"RegularSeasonStartDate"`

	NewPlayerGroupNumber int `gorm:"default:1;column:new_player_group_count" json:"NewPlayerGroupNumber"` // used to assign a group number for new players

	// Relationships
	OwnerUser *User   `gorm:"foreignKey:owner_user_id;references:id" json:"OwnerUser,omitempty"`
	Players   []Player `gorm:"foreignKey:league_id" json:"Players,omitempty"`
	// League has many LeaguePokemon (its defined draft pool)
	DefinedPokemon []LeaguePokemon `gorm:"foreignKey:league_id" json:"-"`
	// League has many DraftedPokemon (all Pokemon drafted in this league)
	AllDraftedPokemon []DraftedPokemon `gorm:"foreignKey:league_id" json:"-"`
}
