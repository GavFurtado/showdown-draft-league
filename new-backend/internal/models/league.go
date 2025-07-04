package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type League struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name                     string         `gorm:"not null" json:"name"`
	StartDate                time.Time      `gorm:"not null" json:"start_date"`
	EndDate                  *time.Time     `json:"end_date"`
	RulesetID                *uuid.UUID     `gorm:"type:uuid;" json:"ruleset_id"`
	Status                   LeagueStatus   `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	MaxPokemonPerPlayer      uint           `gorm:"not null;default:0" json:"max_pokemon_per_player"`
	AllowWeeklyFreeAgents    bool           `gorm:"not null;default:false" json:"allow_free_agents"` // in case this gets more complex
	FreeAgencyPointsPerRound int            `gorm:"not null;default:1" json:"free_agency_points_per_round"`
	AllowTrading             bool           `gorm:"not null;default:false" json:"allow_trading"`
	StartingDraftPoints      int            `gorm:"not null;default:140" json:"starting_draft_points"`
	IsSnakeRoundDraft        bool           `gorm:"not null;default:true" json:"is_snake_round_draft"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
	DiscordWebhookURL        *string        `json:"discord_webhoook_url"`

	// Relationships
	CommissionerUserID uuid.UUID `gorm:"type:uuid;not null" json:"commissioner_id"`
	CommissionerUser   User      `gorm:"foreignKey:CommissionerUserID;references:ID"`
	Players            []Player  `gorm:"foreignKey:LeagueID"`
	// League has many LeaguePokemon (its defined draft pool)
	DefinedPokemon []LeaguePokemon `gorm:"foreignKey:LeagueID" json:"-"`
	// League has many DraftedPokemon (all Pokemon drafted in this league)
	AllDraftedPokemon []DraftedPokemon `gorm:"foreignKey:LeagueID" json:"-"` // Useful for checking global draft status
}

// this might be breaking convention by having functions in the models but idgaf
type LeagueStatus string

const (
	LeagueStatusSetup         LeagueStatus = "SETUP"
	LeagueStatusDrafting      LeagueStatus = "DRAFTING"
	LeagueStatusRegularSeason LeagueStatus = "REGULARSEASON"
	LeagueStatusPlayoffs      LeagueStatus = "PLAYOFFS"
	LeagueStatusCompleted     LeagueStatus = "COMPLETED"
	LeagueStatusCancelled     LeagueStatus = "CANCELLED"
)

var LeagueStatuses = []LeagueStatus{
	LeagueStatusSetup,
	LeagueStatusDrafting,
	LeagueStatusRegularSeason,
	LeagueStatusPlayoffs,
	LeagueStatusCompleted,
	LeagueStatusCancelled,
}

// checks if LeagueStatus IsValid
func (ls LeagueStatus) IsValid() bool {
	for _, status := range LeagueStatuses {
		if ls == status {
			return true
		}
	}
	return false
}

// Stringer() interface implementation in case it's needed
func (ls LeagueStatus) String() string {
	return string(ls)
}

// Extra work to make it work with the DB (more specifically GORM)

// Value() implements the driver.Valuer interface for GORM/database saving.
// Tells GORM how to convert the custom type into a database-compatible type (string).
func (ls LeagueStatus) Value() (driver.Value, error) {
	if !ls.IsValid() {
		return nil, fmt.Errorf("invalid LeagueStatus value: %s", ls)
	}
	return string(ls), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (ls *LeagueStatus) Scan(value interface{}) error {
	if value == nil {
		*ls = "" // Or some default "empty" state if appropriate
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("LeagueStatus: expected string, got %T", value)
	}
	// Important: Validate the string from the database to ensure it's a known status
	newStatus := LeagueStatus(str)
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid LeagueStatus value retrieved from DB: %s", str)
	}
	*ls = newStatus
	return nil
}

// Capitalizes the whole string to ensure it's all Normalized
func (ls LeagueStatus) Normalize() LeagueStatus {
	return LeagueStatus(strings.ToUpper(string(ls)))
}
