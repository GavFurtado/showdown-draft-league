package responses

import (
	"github.com/google/uuid"
)

type TransactionEvent struct {
	PoolEntryID   uuid.UUID `json:"PoolEntryID"`
	PokemonName   string    `json:"PokemonName"`
	PokemonSprite string    `json:"PokemonSprite"`
	EventType     string    `json:"EventType"`
	Cost          int       `json:"Cost"`
	Week          int       `json:"Week"`
	Timestamp     string    `json:"Timestamp"`
}
