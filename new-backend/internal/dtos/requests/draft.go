package requests

import (
	"github.com/google/uuid"
)

type DraftMakePickRequestDTO struct {
	RequestedPickCount int                `json:"RequestedPickCount" binding:"required"`
	RequestedPicks     []RequestedPickDTO `json:"RequestedPicks" binding:"required"`
}

type RequestedPickDTO struct {
	PoolEntryID     uuid.UUID `json:"PoolEntryID" binding:"required"`
	DraftPickNumber int       `json:"DraftPickNumber" binding:"required"`
}
