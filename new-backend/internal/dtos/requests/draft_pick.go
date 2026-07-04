package requests

import (
	"github.com/google/uuid"
)

type DraftPickCreateRequestDTO struct {
	DraftID     uuid.UUID `json:"DraftID" validate:"required"`
	PlayerID    uuid.UUID `json:"PlayerID" validate:"required"`
	PoolEntryID uuid.UUID `json:"PoolEntryID" validate:"required"`
	RoundNumber int       `json:"RoundNumber"`
	PickNumber  int       `json:"PickNumber"`
}
