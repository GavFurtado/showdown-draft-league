package models

import (
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Claim tracks a player's acquisition and tenure of a specific Pokemon species
// within a league. It replaces the ownership semantics that DraftedPokemon
// used to carry.
//
// Each Claim is created when a player acquires a Pokemon (draft or free-agent
// pickup) and remains active until the player drops it. A dropped Claim has
// IsActive=false and ReleasedWeek set; the row itself is never deleted.
//
// Source tells you the acquisition method. SourceID is a polymorphic reference
// that points to DraftPick.ID when Source="draft" and is nil otherwise. There
// is no database-level FK on SourceID. Referential integrity is enforced here
// in the application and not in the database.
//
// Invariant: a player cannot have more than one active Claim for the same
// species (enforced at the application layer, not the database).
type Claim struct {
	ID        uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID  uuid.UUID         `gorm:"type:uuid;not null;column:league_id" json:"LeagueID"`
	PlayerID  uuid.UUID         `gorm:"type:uuid;not null;column:player_id" json:"PlayerID"`
	SpeciesID int64             `gorm:"not null;column:species_id" json:"SpeciesID"`
	Source    enums.ClaimSource `gorm:"type:varchar(20);not null;column:source" json:"Source"`
	// SourceID is polymorphic: points to DraftPick.ID when Source="DRAFT"; nil for FA. No GORM FK constraint
	SourceID     *uuid.UUID `gorm:"type:uuid;column:source_id" json:"SourceID"`
	CostPaid     int        `gorm:"not null;default:0;column:cost_paid" json:"CostPaid"`
	AcquiredWeek int        `gorm:"not null;column:acquired_week" json:"AcquiredWeek"`
	// ReleasedWeek set when IsActive becomes false
	ReleasedWeek *int           `gorm:"column:released_week" json:"ReleasedWeek,omitempty"`
	IsActive     bool           `gorm:"not null;default:true;column:is_active" json:"IsActive"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"UpdatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	League         *League         `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	Player         *Player         `gorm:"foreignKey:player_id;references:id" json:"Player,omitempty"`
	PokemonSpecies *PokemonSpecies `gorm:"foreignKey:species_id;references:id" json:"PokemonSpecies,omitempty"`
}
