// internal/models/drafted_pokemon.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DraftedPokemon represents a specific Pokemon species that has been drafted by a player in a league.
type DraftedPokemon struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"ID"`
	LeagueID         uuid.UUID `gorm:"type:uuid;not null;column:league_id" json:"LeagueID"`
	PlayerID         uuid.UUID `gorm:"type:uuid;not null;column:player_id" json:"PlayerID"`
	PokemonSpeciesID int64     `gorm:"type:int64;not null;column:pokemon_species_id" json:"PokemonSpeciesID"` // Which base species was drafted? (used to skip checking the leaguePokemon)
	LeaguePokemonID  uuid.UUID `gorm:"type:uuid;not null;column:league_pokemon_id" json:"LeaguePokemonID"`

	DraftRoundNumber int  `gorm:"column:draft_round_number" json:"DraftRoundNumber,omitempty"`
	DraftPickNumber  int  `gorm:"column:draft_pick_number" json:"DraftPickNumber,omitempty"`
	IsReleased       bool `gorm:"column:is_released" json:"IsReleased"`
	AcquiredWeek     int  `gorm:"column:acquired_week" json:"AcquiredWeek"`
	ReleasedWeek     *int `gorm:"column:released_week" json:"ReleasedWeek,omitempty"`

	CreatedAt time.Time      `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// Relationships
	League         *League         `gorm:"foreignKey:league_id;references:id" json:"League,omitempty"`
	Player         *Player         `gorm:"foreignKey:player_id;references:id" json:"Player,omitempty"`
	PokemonSpecies *PokemonSpecies `gorm:"foreignKey:pokemon_species_id;references:id" json:"PokemonSpecies,omitempty"`
	LeaguePokemon  *LeaguePokemon  `gorm:"foreignKey:league_pokemon_id;references:id" json:"LeaguePokemon,omitempty"}`
}
