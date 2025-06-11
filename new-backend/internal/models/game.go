package models

import (
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
	WinnerID            *uuid.UUID     `gorm:"type:uuid" json:"winner_id"`              // Nullable: until a winner is determined
	LoserID             *uuid.UUID     `gorm:"type:uuid" json:"loser_id"`               // Nullable
	Player1Wins         int            `gorm:"default:0;not null" json:"player_1_wins"` // Score for Player 1
	Player2Wins         int            `gorm:"default:0;not null" json:"player_2_wins"` // Score for Player 2
	RoundNumber         int            `gorm:"not null" json:"round_number"`            // The round/week number in the tournament
	Status              GameStatus     `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	ReportedByUserID    *uuid.UUID     `gorm:"type:uuid" json:"reported_by_user_id"` // Who reported the result
	ShowdownReplayLinks []string       `gorm:"type:text[]" json:"replay_links"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	League   League `gorm:"foreignKey:LeagueID"`
	Player1  Player `gorm:"foreignKey:Player1ID"`
	Player2  Player `gorm:"foreignKey:Player2ID"`
	Winner   Player `gorm:"foreignKey:WinnerID"`
	Loser    Player `gorm:"foreignKey:LoserID"`
	Reporter User   `gorm:"foreignKey:ReportedByUserID"`
}

// GameStatus defines the possible states of a game.
type GameStatus string

const (
	GameStatusPending   GameStatus = "pending"
	GameStatusCompleted GameStatus = "completed"
	GameStatusDisputed  GameStatus = "disputed"
)
