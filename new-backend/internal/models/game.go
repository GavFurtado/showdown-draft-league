package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Game represents a best-of-x series between two players in a league.
type Game struct {
	ID                  uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID            uuid.UUID        `gorm:"type:uuid;not null;column:league_id" json:"LeagueID"`
	Player1ID           uuid.UUID        `gorm:"type:uuid;not null;column:player1_id" json:"Player1ID"`
	Player2ID           uuid.UUID        `gorm:"type:uuid;not null;column:player2_id" json:"Player2ID"`
	WinnerID            *uuid.UUID       `gorm:"type:uuid;column:winner_id" json:"WinnerID"`
	LoserID             *uuid.UUID       `gorm:"type:uuid;column:loser_id" json:"LoserID"`
	Player1Wins         int              `gorm:"default:0;not null;column:player1_wins" json:"Player1Wins"`
	Player2Wins         int              `gorm:"default:0;not null;column:player2_wins" json:"Player2Wins"`
	WeekNumber          int              `gorm:"not null;column:week_number" json:"WeekNumber"`
	Status              enums.GameStatus `gorm:"type:varchar(50);not null;default:'pending';column:status" json:"Status"`
	ReportingPlayerID   uuid.UUID        `gorm:"type:uuid;not null;column:reporting_player_id" json:"ReportingPlayerID"`
	ShowdownReplayLinks []string         `gorm:"type:jsonb;column:showdown_replay_links" binding:"url" json:"ShowdownReplayLinks"`
	CreatedAt           time.Time        `json:"CreatedAt" gorm:"column:created_at"`
	UpdatedAt           time.Time        `json:"UpdatedAt" gorm:"column:updated_at"`
	DeletedAt           gorm.DeletedAt   `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	ReportingPlayer Player  `gorm:"foreignKey:reporting_player_id;references:ID" json:"ReportingPlayer,omitempty"`
	League          League  `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	Player1         *Player `gorm:"foreignKey:player1_id;references:ID" json:"Player1,omitempty"`
	Player2         *Player `gorm:"foreignKey:player2_id;references:ID" json:"Player2,omitempty"`
	Winner          Player  `gorm:"foreignKey:winner_id;references:ID" json:"Winner,omitempty"`
	Loser           Player  `gorm:"foreignKey:loser_id;references:ID" json:"Loser,omitempty"`
}
