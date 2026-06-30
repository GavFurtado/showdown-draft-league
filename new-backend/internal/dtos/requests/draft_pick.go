package requests

import (
	"github.com/google/uuid"
)

type DraftPickCreateRequestDTO struct {
	DraftID     uuid.UUID `json:"DraftID" binding:"required"`
	PlayerID    uuid.UUID `json:"PlayerID" binding:"required"`
	PoolEntryID uuid.UUID `json:"PoolEntryID" binding:"required"`
	RoundNumber int       `json:"RoundNumber"`
	PickNumber  int       `json:"PickNumber"`
}
