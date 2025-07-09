// internal/models/drafted_pokemon.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DraftedPokemon represents a specific Pokemon species that has been drafted by a player in a league.
type DraftedPokemon struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeagueID         uuid.UUID `gorm:"type:uuid;not null" json:"league_id"`
	PlayerID         uuid.UUID `gorm:"type:uuid;not null" json:"player_id"`
	PokemonSpeciesID int64     `gorm:"type:uuid;not null" json:"pokemon_species_id"` // Which base species was drafted? (used to skip checking the leaguePokemon)
	LeaguePokemonID  uuid.UUID `gorm:"type:uuid;not null" json:"league_pokemon_id"`

	DraftRoundNumber int `json:"draft_round_number"` // The round this pokemon was drafted in
	DraftPickNumber  int `json:"draft_pick_number"`  // The sequential number of this pick in the draft
	// IsReleased: True if the Pokemon has been released back to the draft pool (free agents)
	IsReleased bool           `gorm:"default:false" json:"is_released"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	League         League         `gorm:"foreignKey:LeagueID;references:ID"`
	Player         Player         `gorm:"foreignKey:PlayerID;references:ID"`
	PokemonSpecies PokemonSpecies `gorm:"foreignKey:PokemonSpeciesID;references:ID"`
	LeaguePokemon  LeaguePokemon  `gorm:"foreignKey:LeaguePokemonID;references:ID"`
}
