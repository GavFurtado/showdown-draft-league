package requests

import (
	"github.com/google/uuid"
)

type DraftMakePickRequestDTO struct {
	RequestedPickCount int                `json:"RequestedPickCount" validate:"required"`
	RequestedPicks     []RequestedPickDTO `json:"RequestedPicks" validate:"required"`
}

type RequestedPickDTO struct {
	PoolEntryID     uuid.UUID `json:"PoolEntryID" validate:"required"`
	DraftPickNumber int       `json:"DraftPickNumber" validate:"required"`
}
