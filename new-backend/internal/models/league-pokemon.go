package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeaguePokemon are the Pokemon availabile for a particular league
// This (the whole table) essentially represents the "League's Draft Pool" for a given league.
type LeaguePokemon struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_league_pokemon_species" json:"league_id"`
	PokemonSpeciesID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_league_pokemon_species" json:"pokemon_species_id"`
	Cost             int       `gorm:"not null" json:"cost"`                      // League-specific cost for this Pokemon species
	IsAvailable      bool      `gorm:"not null;default:true" json:"is_available"` // Can this species be drafted in this league?

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	League         League         `gorm:"foreignKey:LeagueID;references:ID"`
	PokemonSpecies PokemonSpecies `gorm:"foreignKey:PokemonSpeciesID;references:ID"`
}
