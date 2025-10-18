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
	ID                          uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID                    uuid.UUID              `gorm:"type:uuid;not null;index;unique;column:league_id" json:"LeagueID"`
	Status                      enums.DraftStatus      `gorm:"type:varchar(50);not null;default:'PENDING';column:status" json:"Status"`
	CurrentTurnPlayerID         *uuid.UUID             `gorm:"type:uuid;index;column:current_turn_player_id" json:"CurrentTurnPlayerID"` // Nullable: Player whose turn it is
	CurrentRound                int                    `gorm:"default:0;not null;column:current_round" json:"CurrentRound"`
	CurrentPickInRound          int                    `gorm:"default:1;not null;column:current_pick_in_round" json:"CurrentPickInRound"`
	CurrentPickOnClock          int                    `gorm:"default:1;not null" json:"CurrentPickOnClock"` // aka CurrentOverallPickNumber
	PlayersWithAccumulatedPicks PlayerAccumulatedPicks `gorm:"type:jsonb;column:players_with_accumulated_picks" json:"PlayersWithAccumulatedPicks"`
	CurrentTurnStartTime        *time.Time             `gorm:"type:timestamp with time zone;column:current_turn_start_time" json:"CurrentTurnStartTime"`
	TurnTimeLimit               int                    `gorm:"default:1440;not null;column:turn_time_limit" json:"TurnTimeLimit"`
	StartTime                   time.Time              `gorm:"type:timestamp with time zone;column:start_time" json:"StartTime"`
	EndTime                     time.Time              `gorm:"type:timestamp with time zone;column:end_time" json:"EndTime"`
	CreatedAt                   time.Time              `gorm:"type:timestamp with time zone;column:created_at" json:"CreatedAt"`
	UpdatedAt                   time.Time              `gorm:"type:timestamp with time zone;column:updated_at" json:"UpdatedAt"`
	DeletedAt                   gorm.DeletedAt         `gorm:"index;type:timestamp with time zone;column:deleted_at" json:"-"`

	// Relationships
	League            *League `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	CurrentTurnPlayer *Player `gorm:"foreignKey:current_turn_player_id;references:id" json:"CurrentTurnPlayer,omitempty"`
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
