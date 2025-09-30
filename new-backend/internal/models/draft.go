package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Draft represents a draft event for a league, managing the real-time state of the draft process.
type Draft struct {
	ID                          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID                    uuid.UUID         `gorm:"type:uuid;not null;index;unique" json:"league_id"` // Added unique, assuming one draft per league
	Status                      enums.DraftStatus `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	CurrentTurnPlayerID         *uuid.UUID        `gorm:"type:uuid;index" json:"current_turn_player_id"` // Nullable: Player whose turn it is
	CurrentRound                int               `gorm:"default:1;not null" json:"current_round"`
	CurrentPickInRound          int               `gorm:"default:0;not null" json:"current_pick_in_round"`
	PlayersWithAccumulatedPicks map[uuid.UUID]int `gorm:"type:jsonb" json:"players_with_accumulated_picks"`
	CurrentTurnStartTime        *time.Time        `gorm:"type:timestamp with time zone" json:"current_turn_start_time"`
	TurnTimeLimit               int               `gorm:"default:43200;not null" json:"turn_time_limit"`
	StartTime                   *time.Time        `gorm:"type:timestamp with time zone" json:"start_time"`
	EndTime                     *time.Time        `gorm:"type:timestamp with time zone" json:"end_time"`
	CreatedAt                   time.Time         `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt                   time.Time         `gorm:"type:timestamp with time zone" json:"updated_at"`
	DeletedAt                   gorm.DeletedAt    `gorm:"index;type:timestamp with time zone" json:"-"`

	// Relationships
	League            League `gorm:"foreignKey:LeagueID;references:ID"`
	CurrentTurnPlayer Player `gorm:"foreignKey:CurrentTurnPlayerID;references:ID"`
}
