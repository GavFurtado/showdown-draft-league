package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// League defines the structure of a League
type League struct {
	ID                  uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name                string             `gorm:"not null" json:"name"`
	StartDate           time.Time          `gorm:"not null" json:"start_date"`
	EndDate             *time.Time         `json:"end_date"` // this is set when the league is cancelled or actualy ends, nil otherwise
	RulesetDescription  string             `gorm:"type:text" json:"ruleset_description"`
	Status              enums.LeagueStatus `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	MaxPokemonPerPlayer int                `gorm:"not null;default:0" json:"max_pokemon_per_player"`
	MinPokemonPerPlayer int                `gorm:"not null;default:0" json:"min_pokemon_per_player"`
	StartingDraftPoints int                `gorm:"not null;default:140" json:"starting_draft_points"`
	Format              LeagueFormat       `gorm:"type:jsonb" json:"format"`
	CreatedAt           *time.Time         `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt           *time.Time         `gorm:"type:timestamp with time zone" json:"updated_at"`
	DeletedAt           gorm.DeletedAt     `gorm:"index" json:"-"`
	DiscordWebhookURL   *string            `json:"discord_webhook_url"`

	// Relationships
	Players []Player `gorm:"foreignKey:LeagueID"`
	// League has many LeaguePokemon (its defined draft pool)
	DefinedPokemon []LeaguePokemon `gorm:"foreignKey:LeagueID" json:"-"`
	// League has many DraftedPokemon (all Pokemon drafted in this league)
	AllDraftedPokemon []DraftedPokemon `gorm:"foreignKey:LeagueID" json:"-"`
}

// LeagueFormat defines the structure for various optional league settings.
type LeagueFormat struct {
	IsSnakeRoundDraft bool                   `json:"is_snake_round_draft"`
	DraftOrderType    enums.DraftOrderType   `gorm:"default:'random'" json:"draft_order_type"` // "PENDING", "RANDOM", "MANUAL"
	SeasonType        enums.LeagueSeasonType `json:"season_type"`                              // "ROUND_ROBIN_ONLY", "PLAYOFFS_ONLY", "HYBRID"
	GroupCount        int                    `json:"group_count"`                              // Relevant if SeasonType is "GROUPS"
	GamesPerOpponent  int                    `json:"games_per_opponent"`                       // For round-robin or group stages

	PlayoffType             enums.LeaguePlayoffType        `json:"playoff_type"`              // "NONE", "SINGLE_ELIM", "DOUBLE_ELIM"
	PlayoffParticipantCount int                            `json:"playoff_participant_count"` // Number of teams that make playoffs
	PlayoffByesCount        int                            `json:"playoff_byes_count"`        // Number of teams getting a bye in playoffs
	PlayoffSeedingType      enums.LeaguePlayoffSeedingType `json:"playoff_seeding_type"`      // "STANDARD", "SEEDED", "BYES_ONLY"

	AllowTrading                bool `json:"allow_trading"`
	AllowTransferCredits        bool `json:"allow_transfer_credits"`
	TransferCreditsPerWindow    int  `json:"transfer_credits_per_window"`
	TransferCreditCap           int  `json:"transfer_credit_cap"`
	TransferWindowFrequencyDays int  `json:"transfer_window_frequency_days"`
	TransferWindowDuration      int  `json:"transfer_window_duration"`
	DropCost                    int  `json:"drop_cost"`
	PickupCost                  int  `json:"pickup_cost"`
	// NextTransferWindowStart stores the next occurence of a trasnfer window
	// if league is in a transfer window, NextTransferWindowStart will store the start time of the window
	// NextTransferWindowStart is updated at the *end* of a transfer window or when the season/bracket starts
	NextTransferWindowStart *time.Time `gorm:"type:timestamp with time zone" json:"next_transfer_window_start"`
}

func (f *LeagueFormat) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan LeagueFormat: %v", value)
	}
	return json.Unmarshal(b, f)
}

func (f LeagueFormat) Value() (driver.Value, error) {
	return json.Marshal(f)
}
