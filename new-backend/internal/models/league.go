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
	ID                  uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	Name                string             `gorm:"not null;column:name" json:"Name"`
	StartDate           time.Time          `gorm:"not null;column:start_date" json:"StartDate"`
	EndDate             *time.Time         `gorm:"column:end_date" json:"EndDate"` // this is set when the league is cancelled or actualy ends, nil otherwise
	RulesetDescription  string             `gorm:"type:text;column:ruleset_description" json:"RulesetDescription"`
	Status              enums.LeagueStatus `gorm:"type:varchar(50);not null;default:'pending';column:status" json:"Status"`
	MaxPokemonPerPlayer int                `gorm:"not null;default:0;column:max_pokemon_per_player" json:"MaxPokemonPerPlayer"`
	MinPokemonPerPlayer int                `gorm:"not null;default:0;column:min_pokemon_per_player" json:"MinPokemonPerPlayer"`
	StartingDraftPoints int                `gorm:"not null;default:140;column:starting_draft_points" json:"StartingDraftPoints"`
	Format              *LeagueFormat      `gorm:"type:jsonb;column:format" json:"Format,omitempty"`
	CreatedAt           *time.Time         `gorm:"type:timestamp with time zone;column:created_at" json:"CreatedAt"`
	UpdatedAt           *time.Time         `gorm:"type:timestamp with time zone;column:updated_at" json:"UpdatedAt"`
	DeletedAt           gorm.DeletedAt     `gorm:"index;column:deleted_at" json:"-"`
	DiscordWebhookURL   *string            `gorm:"column:discord_webhook_url" json:"DiscordWebhookURL"`
	CurrentWeekNumber   int                `gorm:"not null;default:0;column:current_week_number" json:"CurrentWeekNumber"` // starts with 1, 0 is invalid and used when not relevant

	NewPlayerGroupNumber int `gorm:"default:1;column:new_player_group_count" json:"NewPlayerGroupCount"` // used to assign a group number for new players

	// Relationships
	Players []Player `gorm:"foreignKey:league_id" json:"Players,omitempty"`
	// League has many LeaguePokemon (its defined draft pool)
	DefinedPokemon []LeaguePokemon `gorm:"foreignKey:league_id" json:"-"`
	// League has many DraftedPokemon (all Pokemon drafted in this league)
	AllDraftedPokemon []DraftedPokemon `gorm:"foreignKey:league_id" json:"-"`
}

// LeagueFormat defines the structure for various optional league settings.
type LeagueFormat struct {
	IsSnakeRoundDraft bool                   `json:"IsSnakeRoundDraft" gorm:"column:is_snake_round_draft"`
	DraftOrderType    enums.DraftOrderType   `gorm:"default:'RANDOM'" json:"DraftOrderType"` // "PENDING", "RANDOM", "MANUAL"
	SeasonType        enums.LeagueSeasonType `json:"SeasonType"`                             // "ROUND_ROBIN_ONLY", "BRACKET_ONLY", "HYBRID"
	GroupCount        int                    `json:"GroupCount"`

	PlayoffType             enums.LeaguePlayoffType        `json:"PlayoffType"`             // "NONE", "SINGLE_ELIM", "DOUBLE_ELIM"
	PlayoffParticipantCount int                            `json:"PlayoffParticipantCount"` // Number of teams that make playoffs
	PlayoffByesCount        int                            `json:"PlayoffByesCount"`        // Number of teams getting a bye in playoffs
	PlayoffSeedingType      enums.LeaguePlayoffSeedingType `json:"PlayoffSeedingType"`      // "STANDARD", "SEEDED", "BYES_ONLY"

	AllowTrading                bool `json:"AllowTrading" gorm:"column:allow_trading"`
	AllowTransferCredits        bool `json:"AllowTransferCredits" gorm:"column:allow_transfer_credits"`
	TransferCreditsPerWindow    int  `json:"TransferCreditsPerWindow" gorm:"column:transfer_credits_per_window"`
	TransferCreditCap           int  `json:"TransferCreditCap" gorm:"column:transfer_credit_cap"`
	TransferWindowFrequencyDays int  `json:"TransferWindowFrequencyDays" gorm:"column:transfer_window_frequency_days"`
	TransferWindowDuration      int  `json:"TransferWindowDuration" gorm:"column:transfer_window_duration"`
	DropCost                    int  `json:"DropCost" gorm:"column:drop_cost"`
	PickupCost                  int  `json:"PickupCost" gorm:"column:pickup_cost"`
	// NextTransferWindowStart stores the next occurence of a trasnfer window
	// if league is in a transfer window, NextTransferWindowStart will store the start time of the window
	// NextTransferWindowStart is updated at the *end* of a transfer window or when the season/bracket starts
	NextTransferWindowStart *time.Time `gorm:"type:timestamp with time zone;column:next_transfer_window_start" json:"NextTransferWindowStart"`
}

func (f *LeagueFormat) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan LeagueFormat: %v", value)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if val, ok := m["is_snake_round_draft"].(bool); ok {
		f.IsSnakeRoundDraft = val
	}
	if val, ok := m["draft_order_type"].(string); ok {
		f.DraftOrderType = enums.DraftOrderType(val)
	}
	if val, ok := m["season_type"].(string); ok {
		f.SeasonType = enums.LeagueSeasonType(val)
	}
	if val, ok := m["group_count"].(float64); ok {
		f.GroupCount = int(val)
	}
	if val, ok := m["games_per_opponent"].(float64); ok {
		f.GamesPerOpponent = int(val)
	}
	if val, ok := m["playoff_type"].(string); ok {
		f.PlayoffType = enums.LeaguePlayoffType(val)
	}
	if val, ok := m["playoff_participant_count"].(float64); ok {
		f.PlayoffParticipantCount = int(val)
	}
	if val, ok := m["playoff_byes_count"].(float64); ok {
		f.PlayoffByesCount = int(val)
	}
	if val, ok := m["playoff_seeding_type"].(string); ok {
		f.PlayoffSeedingType = enums.LeaguePlayoffSeedingType(val)
	}
	if val, ok := m["allow_trading"].(bool); ok {
		f.AllowTrading = val
	}
	if val, ok := m["allow_transfer_credits"].(bool); ok {
		f.AllowTransferCredits = val
	}
	if val, ok := m["transfer_credits_per_window"].(float64); ok {
		f.TransferCreditsPerWindow = int(val)
	}
	if val, ok := m["transfer_credit_cap"].(float64); ok {
		f.TransferCreditCap = int(val)
	}
	if val, ok := m["transfer_window_frequency_days"].(float64); ok {
		f.TransferWindowFrequencyDays = int(val)
	}
	if val, ok := m["transfer_window_duration"].(float64); ok {
		f.TransferWindowDuration = int(val)
	}
	if val, ok := m["drop_cost"].(float64); ok {
		f.DropCost = int(val)
	}
	if val, ok := m["pickup_cost"].(float64); ok {
		f.PickupCost = int(val)
	}
	if val, ok := m["next_transfer_window_start"].(string); ok {
		t, err := time.Parse(time.RFC3339, val)
		if err == nil {
			f.NextTransferWindowStart = &t
		}
	}

	return nil
}

func (f LeagueFormat) Value() (driver.Value, error) {
	m := map[string]any{
		"is_snake_round_draft":           f.IsSnakeRoundDraft,
		"draft_order_type":               f.DraftOrderType,
		"season_type":                    f.SeasonType,
		"group_count":                    f.GroupCount,
		"games_per_opponent":             f.GamesPerOpponent,
		"playoff_type":                   f.PlayoffType,
		"playoff_participant_count":      f.PlayoffParticipantCount,
		"playoff_byes_count":             f.PlayoffByesCount,
		"playoff_seeding_type":           f.PlayoffSeedingType,
		"allow_trading":                  f.AllowTrading,
		"allow_transfer_credits":         f.AllowTransferCredits,
		"transfer_credits_per_window":    f.TransferCreditsPerWindow,
		"transfer_credit_cap":            f.TransferCreditCap,
		"transfer_window_frequency_days": f.TransferWindowFrequencyDays,
		"transfer_window_duration":       f.TransferWindowDuration,
		"drop_cost":                      f.DropCost,
		"pickup_cost":                    f.PickupCost,
		"next_transfer_window_start":     f.NextTransferWindowStart,
	}
	return json.Marshal(m)
}
