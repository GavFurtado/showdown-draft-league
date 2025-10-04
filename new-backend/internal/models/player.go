package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Player struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	LeagueID        uuid.UUID       `gorm:"type:uuid;not null" json:"league_id"`
	InLeagueName    string          `json:"in_league_name"`
	TeamName        string          `gorm:"not null" json:"team_name"`
	Wins            int             `gorm:"default:0;not null" json:"wins"`
	Losses          int             `gorm:"default:0;not null" json:"losses"`
	DraftPoints     int             `gorm:"default:140;not null" json:"draft_points"`
	DraftPosition   int             `json:"draft_position"` // turn order of player pick (possibly randomized)
	Role            rbac.PlayerRole `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	IsParticipating bool
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User   User           `gorm:"foreignKey:UserID;references:ID"`
	League League         `gorm:"foreignKey:LeagueID;references:ID"`
	Roster []PlayerRoster `gorm:"foreignKey:PlayerID;references:ID;inverseOf:Player"`
}

// IsLeagueOwner checks if the player has the LeagueOwner role.
func (p *Player) IsLeagueOwner() bool {
	return p.Role.IsOwner()
}

// IsLeagueModerator checks if the player has the LeagueModerator role.
func (p *Player) IsLeagueModerator() bool {
	return p.Role.IsModerator()
}

// Can checks if the player's role has a specific permission.
func (p *Player) Can(permission rbac.Permission) bool {
	return p.Role.HasPermission(permission)
}
