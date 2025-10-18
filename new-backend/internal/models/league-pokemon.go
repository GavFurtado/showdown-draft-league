package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeaguePokemon are the Pokemon availabile for a particular league
type LeaguePokemon struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_league_pokemon_species;column:league_id" json:"LeagueID"`
	PokemonSpeciesID int64          `gorm:"type:not null;uniqueIndex:idx_league_pokemon_species;column:pokemon_species_id" json:"PokemonSpeciesID"`
	Cost             *int           `gorm:"not null;column:cost" json:"Cost"`                             // League-specific cost for this Pokemon species
	IsAvailable      bool           `gorm:"not null;default:true;column:is_available" json:"IsAvailable"` // Can this species be drafted in this league?
	CreatedAt        time.Time      `json:"CreatedAt" gorm:"column:created_at"`
	UpdatedAt        time.Time      `json:"UpdatedAt" gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	League         *League         `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	PokemonSpecies *PokemonSpecies `gorm:"foreignKey:pokemon_species_id;references:id" json:"PokemonSpecies,omitempty"`
}
