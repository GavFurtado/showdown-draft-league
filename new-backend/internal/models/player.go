package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Player struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null;column:user_id" json:"UserID"`
	LeagueID        uuid.UUID       `gorm:"type:uuid;not null;column:league_id" json:"LeagueID"`
	InLeagueName    string          `gorm:"column:in_league_name" json:"InLeagueName"`
	TeamName        string          `gorm:"not null;column:team_name" json:"TeamName"`
	Wins            int             `gorm:"default:0;not null;column:wins" json:"Wins"`
	Losses          int             `gorm:"default:0;not null;column:losses" json:"Losses"`
	DraftPoints     int             `gorm:"default:140;not null;column:draft_points" json:"DraftPoints"`
	TransferCredits int             `gorm:"default:0;column:transfer_credits" json:"TransferCredits"`
	DraftPosition   int             `gorm:"default:1;column:draft_position" json:"DraftPosition"` // turn order of player pick (possibly randomized)
	Role            rbac.PlayerRole `gorm:"type:varchar(20);not null;default:'member';column:role" json:"Role"`
	IsParticipating bool            `gorm:"column:is_participating" json:"IsParticipating"`
	CreatedAt       time.Time       `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt       time.Time       `gorm:"column:updated_at" json:"UpdatedAt"`
	DeletedAt       gorm.DeletedAt  `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	User   User   `gorm:"foreignKey:user_id;references:id" json:"User,omitempty"`
	League League `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
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
