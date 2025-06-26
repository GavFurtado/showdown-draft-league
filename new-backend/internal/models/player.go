package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Player struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;unique;not null" json:"user_id"`
	LeagueID          uuid.UUID      `gorm:"type:uuid;unique;not null" json:"league_id"`
	InLeagueName      string         `json:"in_league_name"`
	TeamName          string         `gorm:"not null" json:"team_name"`
	Wins              int            `gorm:"default:0;not null" json:"wins"`
	Losses            int            `gorm:"default:0;not null" json:"losses"`
	DraftPoints       int            `gorm:"default:120;not null" json:"points"`
	DraftPosition     int            `json:"draft_position"` // turn order of player pick (possibly/probably randomized)
	IsCommissioner    bool           `gorm:"not null;default:false" json:"is_commissioner"`
	IsSnakeRoundDraft bool           `gorm:"not null;default:true" json:"is_snake_round_draft"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User   User           `gorm:"foreignKey:UserID"`
	League League         `gorm:"foreignKey:LeagueID"`
	Roster []PlayerRoster `gorm:"foreignKey:PlayerID" json:"Roster"`
}
