package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PoolEntry represents a Pokemon species in a league's draft/transfer pool.
// Each league has its own pool of PoolEntries with league-specific costs.
// When a player acquires a Pokemon (via draft or free-agent pickup), the
// corresponding PoolEntry is marked as unavailable.
// (Previously named LeaguePokemon)
type PoolEntry struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_pool_entry_species;column:league_id" json:"LeagueID"`
	PokemonSpeciesID int64          `gorm:"not null;uniqueIndex:idx_pool_entry_species;column:pokemon_species_id" json:"PokemonSpeciesID"`
	Cost             *int           `gorm:"not null;column:cost" json:"Cost"`
	IsAvailable      bool           `gorm:"not null;default:true;column:is_available" json:"IsAvailable"`
	CreatedAt        time.Time      `json:"CreatedAt" gorm:"column:created_at"`
	UpdatedAt        time.Time      `json:"UpdatedAt" gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	League         *League         `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	PokemonSpecies *PokemonSpecies `gorm:"foreignKey:pokemon_species_id;references:id" json:"PokemonSpecies,omitempty"`
}
