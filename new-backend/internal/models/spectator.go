package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Spectator struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null;column:user_id;uniqueIndex:idx_spectator_user_league" json:"UserID"`
	LeagueID     uuid.UUID       `gorm:"type:uuid;not null;column:league_id;uniqueIndex:idx_spectator_user_league" json:"LeagueID"`
	InLeagueName *string         `gorm:"column:in_league_name" json:"InLeagueName"`
	Role         rbac.MemberRole `gorm:"type:varchar(20);not null;default:'MEMBER';column:role" json:"Role"`
	IsViewer     bool            `gorm:"not null;default:false;column:is_viewer" json:"IsViewer"`
	ExpiresAt    *time.Time      `gorm:"type:timestamp with time zone;column:expires_at" json:"ExpiresAt"`
	CreatedAt    *time.Time      `gorm:"type:timestamp with time zone;column:created_at" json:"CreatedAt"`
	UpdatedAt    *time.Time      `gorm:"type:timestamp with time zone;column:updated_at" json:"UpdatedAt"`
	DeletedAt    gorm.DeletedAt  `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	User   *User   `gorm:"foreignKey:user_id;references:id" json:"User,omitempty"`
	League *League `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
}

// IsOwner checks if the spectator has the LeagueOwner role.
func (m *Spectator) IsOwner() bool {
	return m.Role.IsOwner()
}

// IsModerator checks if the spectator has the LeagueModerator role.
func (m *Spectator) IsModerator() bool {
	return m.Role.IsModerator()
}

// Can checks if the spectator's role has a specific permission.
func (m *Spectator) Can(permission rbac.Permission) bool {
	return m.Role.HasPermission(permission)
}
