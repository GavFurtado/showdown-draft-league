package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Draft represents a draft event for a league, managing the real-time state of the draft process.
type Draft struct {
	ID                          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID                    uuid.UUID         `gorm:"type:uuid;not null;index" json:"league_id"`
	Status                      DraftStatus       `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	CurrentTurnPlayerID         *uuid.UUID        `gorm:"type:uuid;index" json:"current_turn_player_id"` // Nullable: Player whose turn it is
	CurrentRound                int               `gorm:"default:1;not null" json:"current_round"`
	CurrentPickInRound          int               `gorm:"default:0;not null" json:"current_pick_in_round"`              // Current pick number within the round (0-based)
	PlayersWithAccumulatedPicks map[uuid.UUID]int `gorm:"type:jsonb" json:"players_with_accumulated_picks"`             // Map of PlayerID to accumulated picks
	CurrentTurnStartTime        *time.Time        `gorm:"type:timestamp with time zone" json:"current_turn_start_time"` // When the current turn started
	TurnTimeLimit               int               `gorm:"default:43200;not null" json:"turn_time_limit"`                // Time limit per turn in seconds (default: 12 hours = 43200 seconds)
	StartTime                   *time.Time        `gorm:"type:timestamp with time zone" json:"start_time"`              // When draft started
	EndTime                     *time.Time        `gorm:"type:timestamp with time zone" json:"end_time"`                // When draft ended
	CreatedAt                   time.Time         `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt                   time.Time         `gorm:"type:timestamp with time zone" json:"updated_at"`
	DeletedAt                   gorm.DeletedAt    `gorm:"index;type:timestamp with time zone" json:"-"`

	// Relationships
	League            League `gorm:"foreignKey:LeagueID"`
	CurrentTurnPlayer Player `gorm:"foreignKey:CurrentTurnPlayerID"`
}

// DraftStatus defines the possible states of a draft.
type DraftStatus string

const (
	DraftStatusPending   DraftStatus = "PENDING"
	DraftStatusStarted   DraftStatus = "STARTED"
	DraftStatusPaused    DraftStatus = "PAUSED"
	DraftStatusCompleted DraftStatus = "COMPLETED"
)

// Validate DraftStatus for database interactions
func (ds DraftStatus) IsValid() bool {
	switch ds {
	case DraftStatusPending, DraftStatusStarted, DraftStatusPaused, DraftStatusCompleted:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface for GORM/database saving.
func (ds DraftStatus) Value() (driver.Value, error) {
	if !ds.IsValid() {
		return nil, fmt.Errorf("invalid DraftStatus value: %s", ds)
	}
	return string(ds), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
func (ds *DraftStatus) Scan(value interface{}) error {
	if value == nil {
		*ds = DraftStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("DraftStatus: expected string, got %T", value)
	}
	newStatus := DraftStatus(str)
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid DraftStatus value retrieved from DB: %s", str)
	}
	*ds = newStatus
	return nil
}
