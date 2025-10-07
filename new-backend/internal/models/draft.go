package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Draft represents a draft event for a league, managing the real-time state of the draft process.
type Draft struct {
	ID                          uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID                    uuid.UUID              `gorm:"type:uuid;not null;index;unique" json:"league_id"`
	Status                      enums.DraftStatus      `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	CurrentTurnPlayerID         *uuid.UUID             `gorm:"type:uuid;index" json:"current_turn_player_id"` // Nullable: Player whose turn it is
	CurrentRound                int                    `gorm:"default:0;not null" json:"current_round"`
	CurrentPickInRound          int                    `gorm:"default:1;not null" json:"current_pick_in_round"`
	CurrentPickOnClock          int                    `gorm:"default:1;not null" json:"current_pick_on_clock"` // aka CurrentOverallPickNumber
	PlayersWithAccumulatedPicks PlayerAccumulatedPicks `gorm:"type:jsonb" json:"players_with_accumulated_picks"`
	CurrentTurnStartTime        *time.Time             `gorm:"type:timestamp with time zone" json:"current_turn_start_time"`
	TurnTimeLimit               int                    `gorm:"default:43200;not null" json:"turn_time_limit"`
	StartTime                   *time.Time             `gorm:"type:timestamp with time zone" json:"start_time"`
	EndTime                     *time.Time             `gorm:"type:timestamp with time zone" json:"end_time"`
	CreatedAt                   time.Time              `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt                   time.Time              `gorm:"type:timestamp with time zone" json:"updated_at"`
	DeletedAt                   gorm.DeletedAt         `gorm:"index;type:timestamp with time zone" json:"-"`

	// Relationships
	League            League `gorm:"foreignKey:LeagueID;references:ID"`
	CurrentTurnPlayer Player `gorm:"foreignKey:CurrentTurnPlayerID;references:ID"`
}

// PlayerAccumulatedPicks is a custom type for storing a map of player IDs to their accumulated pick numbers.
type PlayerAccumulatedPicks map[uuid.UUID][]int

// Value implements the driver.Valuer interface for PlayerAccumulatedPicks.
func (p PlayerAccumulatedPicks) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	j, err := json.Marshal(p)
	return j, err
}

// Scan implements the sql.Scanner interface for PlayerAccumulatedPicks.
func (p *PlayerAccumulatedPicks) Scan(value interface{}) error {
	if value == nil {
		*p = make(PlayerAccumulatedPicks)
		return nil
	}
	var byteValue []byte
	switch v := value.(type) {
	case []byte:
		byteValue = v
	case string:
		byteValue = []byte(v)
	default:
		return errors.New("unsupported type for PlayerAccumulatedPicks")
	}

	if len(byteValue) == 0 {
		*p = make(PlayerAccumulatedPicks)
		return nil
	}

	// Ensure the map is initialized before unmarshaling
	if *p == nil {
		*p = make(PlayerAccumulatedPicks)
	}
	return json.Unmarshal(byteValue, p)
}
