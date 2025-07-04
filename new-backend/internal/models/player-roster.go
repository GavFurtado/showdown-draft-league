package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Represents **a** specific Pokemon that is currently on a player's active roster.
// Track the current active roster for a player in a league.
type PlayerRoster struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PlayerID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_player_drafted_pokemon" json:"player_id"`
	DraftedPokemonID uuid.UUID      `gorm:"type:uuid;unique;not null;uniqueIndex:idx_player_drafted_pokemon" json:"drafted_pokemon_id"` // specific drafted instance
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Player         Player         `gorm:"foreignKey:PlayerID;references:ID;inverseOf:Roster"`
	DraftedPokemon DraftedPokemon `gorm:"foreignKey:DraftedPokemonID;references:ID"`
}
