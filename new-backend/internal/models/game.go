package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Game represents a best-of-x series between two players in a league.
type Game struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID            uuid.UUID      `gorm:"type:uuid;not null" json:"league_id"`
	Player1ID           uuid.UUID      `gorm:"type:uuid;not null" json:"player_1_id"`
	Player2ID           uuid.UUID      `gorm:"type:uuid;not null" json:"player_2_id"`
	WinnerID            *uuid.UUID     `gorm:"type:uuid" json:"winner_id"`
	LoserID             *uuid.UUID     `gorm:"type:uuid" json:"loser_id"`
	Player1Wins         int            `gorm:"default:0;not null" json:"player_1_wins"`
	Player2Wins         int            `gorm:"default:0;not null" json:"player_2_wins"`
	RoundNumber         int            `gorm:"not null" json:"round_number"`
	Status              GameStatus     `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	ReportedByUserID    *uuid.UUID     `gorm:"type:uuid" json:"reported_by_user_id"`
	ShowdownReplayLinks []string       `gorm:"type:jsonb" binding:"url" json:"replay_links"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	League   League `gorm:"foreignKey:LeagueID;references:ID"`
	Player1  Player `gorm:"foreignKey:Player1ID;references:ID"`
	Player2  Player `gorm:"foreignKey:Player2ID;references:ID"`
	Winner   Player `gorm:"foreignKey:WinnerID;references:ID"`
	Loser    Player `gorm:"foreignKey:LoserID;references:ID"`
	Reporter User   `gorm:"foreignKey:ReportedByUserID;references:ID"`
}

// GameStatus defines the possible states of a game.
type GameStatus string

const (
	GameStatusPending   GameStatus = "pending"
	GameStatusCompleted GameStatus = "completed"
	GameStatusDisputed  GameStatus = "disputed"
)

// List of all valid GameStatuses for validation
var gameStatuses = []GameStatus{
	GameStatusPending,
	GameStatusCompleted,
	GameStatusDisputed,
}

// IsValid checks if the GameStatus is one of the predefined valid statuses.
func (gs GameStatus) IsValid() bool {
	for _, status := range gameStatuses {
		if gs == status {
			return true
		}
	}
	return false
}

// Value implements the driver.Valuer interface for GORM/database saving.
// This tells GORM how to convert the custom type into a database-compatible type (string).
func (gs GameStatus) Value() (driver.Value, error) {
	if !gs.IsValid() {
		return nil, fmt.Errorf("invalid GameStatus value: %s", gs)
	}
	return string(gs), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (gs *GameStatus) Scan(value interface{}) error {
	if value == nil {
		*gs = GameStatusPending // Default or zero value for nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("GameStatus: expected string, got %T", value)
	}
	// Important: Validate the string from the database to ensure it's a known status
	newStatus := GameStatus(str).Normalize()
	if !newStatus.IsValid() {
		// This indicates potential data integrity issues in your DB if an invalid status is retrieved
		return fmt.Errorf("invalid GameStatus value retrieved from DB: %s", str)
	}
	*gs = newStatus
	return nil
}

func (gs GameStatus) Normalize() GameStatus {
	return GameStatus(strings.ToUpper(string(gs)))
}
