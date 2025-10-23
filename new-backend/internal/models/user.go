package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	DiscordID        string    `gorm:"uniqueIndex;not null;column:discord_id" json:"DiscordID"`
	DiscordUsername  string    `gorm:"not null;column:discord_username" json:"DiscordUsername"`
	DiscordAvatarURL string    `gorm:"column:discord_avatar_url" json:"DiscordAvatarURL"`
	ShowdownUsername string    `gorm:"not null; unique;column:showdown_username" json:"ShowdownUsername"`
	Role             string    `gorm:"default:'user';not null;column:role" json:"Role"` // "user", "admin"

	CreatedAt time.Time      `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	// Relationships
	// LeaguesCreated []League `gorm:"foreignKey:CommissionerUserID;references:ID;inverseOf:CommissionerUser"` // List of Leagues this user has created
	Players []Player `gorm:"foreignKey:user_id;references:id" json:"Players,omitempty"` // Player entities in various leagues
}
