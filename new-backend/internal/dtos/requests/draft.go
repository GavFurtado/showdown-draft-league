package requests

import (
	"github.com/google/uuid"
)

type DraftMakePickRequest struct {
	RequestedPickCount int             `json:"RequestedPickCount" binding:"required"`
	RequestedPicks     []RequestedPick `json:"RequestedPicks" binding:"required"`
}

type RequestedPick struct {
	LeaguePokemonID uuid.UUID `json:"LeaguePokemonID" binding:"required"`
	DraftPickNumber int       `json:"DraftPickNumber" binding:"required"`
}
