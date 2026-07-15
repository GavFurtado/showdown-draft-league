package requests

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
)

type LeagueFormatRequestDTO struct {
	IsSnakeRoundDraft           bool                           `json:"IsSnakeRoundDraft"`
	DraftOrderType              enums.DraftOrderType           `json:"DraftOrderType" validate:"isValid"`
	SeasonType                  enums.LeagueSeasonType         `json:"SeasonType" validate:"isValid"`
	GroupCount                  int                            `json:"GroupCount"`
	PlayoffType                 enums.LeaguePlayoffType        `json:"PlayoffType" validate:"isValid"`
	PlayoffParticipantCount     int                            `json:"PlayoffParticipantCount"`
	PlayoffByesCount            int                            `json:"PlayoffByesCount"`
	PlayoffSeedingType          enums.LeaguePlayoffSeedingType `json:"PlayoffSeedingType" validate:"isValid"`
	AllowTransfers              bool                           `json:"AllowTransfers"`
	TransfersCostCredits        bool                           `json:"TransfersCostCredits"`
	TransferCreditsPerWindow    int                            `json:"TransferCreditsPerWindow"`
	TransferCreditCap           int                            `json:"TransferCreditCap"`
	TransferWindowFrequencyDays int                            `json:"TransferWindowFrequencyDays"`
	TransferWindowDuration      int                            `json:"TransferWindowDuration"`
	DropCost                    int                            `json:"DropCost"`
	PickupCost                  int                            `json:"PickupCost"`
}

func (f LeagueFormatRequestDTO) ToLeagueFormat() types.LeagueFormat {
	return types.LeagueFormat{
		IsSnakeRoundDraft:           f.IsSnakeRoundDraft,
		DraftOrderType:              f.DraftOrderType,
		SeasonType:                  f.SeasonType,
		GroupCount:                  f.GroupCount,
		PlayoffType:                 f.PlayoffType,
		PlayoffParticipantCount:     f.PlayoffParticipantCount,
		PlayoffByesCount:            f.PlayoffByesCount,
		PlayoffSeedingType:          f.PlayoffSeedingType,
		AllowTransfers:              f.AllowTransfers,
		TransfersCostCredits:        f.TransfersCostCredits,
		TransferCreditsPerWindow:    f.TransferCreditsPerWindow,
		TransferCreditCap:           f.TransferCreditCap,
		TransferWindowFrequencyDays: f.TransferWindowFrequencyDays,
		TransferWindowDuration:      f.TransferWindowDuration,
		DropCost:                    f.DropCost,
		PickupCost:                  f.PickupCost,
	}
}

func (f LeagueFormatRequestDTO) ToLeagueFormatPtr() *types.LeagueFormat {
	v := f.ToLeagueFormat()
	return &v
}

type LeagueCreateRequestDTO struct {
	Name                string                 `json:"Name" validate:"required"`
	RulesetDescription  string                 `json:"RulesetDescription"`
	MaxPokemonPerPlayer int                    `json:"MaxPokemonPerPlayer" validate:"gte=1,max=20"`
	MinPokemonPerPlayer int                    `json:"MinPokemonPerPlayer" validate:"gte=0,max=20"`
	StartingDraftPoints int                    `json:"StartingDraftPoints" validate:"gte=20,max=1000"`
	MaxPlayers          int                    `json:"MaxPlayers" validate:"gte=0,max=20"`
	Visibility          enums.LeagueVisibility `json:"Visibility" validate:"isValid"`
	Format              LeagueFormatRequestDTO `json:"Format"`
}
