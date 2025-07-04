package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DiscordID        string         `gorm:"uniqueIndex;not null" json:"discord_id"`
	DiscordUsername  string         `gorm:"not null" json:"discord_username"`
	DiscordAvatarURL string         `json:"discord_avatar_url"`
	ShowdownUsername string         `gorm:"not null; unique" json:"showdown_username"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	IsAdmin          bool           `gorm:"default:false;not null" json:"is_admin"` // cannot be altered through the server. Requires manual intervention on the database
	// Relationships
	LeaguesCreated []League `gorm:"foreignKey:CommissionerUserID;references:ID;inverseOf:CommissionerUser"` // List of Leagues this user has created
	Players        []Player `gorm:"foreignKey:UserID;references:ID"`                                        // Player entities in various leagues
}
