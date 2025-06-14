package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Id               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DiscordID        string         `gorm:"uniqueIndex;not null" json:discord_id`
	DiscordUsername  string         `gorm:"not null" json:discord_username`
	DiscordAvatarURL string         `json:"discord_avatar_url"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	// Relationships
	LeaguesCreated []League `gorm:"foreignKey:CommissionerUserID"` // List of Leagues this user has created
	Players        []Player `gorm:"foreignKey:UserID"`             // Player entities in various leagues
}
