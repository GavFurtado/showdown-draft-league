package models

import (
	"time"

	"github.com/google/uuid"
)

// DraftPick is a single pick made during a league's draft.
// It is a pure event log entry. It records that a player selected
// a specific pool entry at a specific point in the draft.
// DraftPick has no ownership semantics; those are captured by Claim.
type DraftPick struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	DraftID     uuid.UUID `gorm:"type:uuid;not null;column:draft_id" json:"DraftID"`
	PlayerID    uuid.UUID `gorm:"type:uuid;not null;column:player_id" json:"PlayerID"`
	PoolEntryID uuid.UUID `gorm:"type:uuid;not null;column:pool_entry_id" json:"PoolEntryID"`
	RoundNumber int       `gorm:"not null;column:round_number" json:"RoundNumber"`
	PickNumber  int       `gorm:"not null;column:pick_number" json:"PickNumber"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"UpdatedAt"`

	// Relationships
	Draft     *Draft       `gorm:"foreignKey:draft_id;references:id" json:"Draft,omitempty"`
	Player    *LeagueMember `gorm:"foreignKey:player_id;references:id" json:"Player,omitempty"`
	PoolEntry *PoolEntry   `gorm:"foreignKey:pool_entry_id;references:id" json:"PoolEntry,omitempty"`
}
