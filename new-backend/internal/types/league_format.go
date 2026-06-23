package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
)

type LeagueFormat struct {
	IsSnakeRoundDraft           bool                           `json:"IsSnakeRoundDraft"`
	DraftOrderType              enums.DraftOrderType           `json:"DraftOrderType"`
	SeasonType                  enums.LeagueSeasonType         `json:"SeasonType"`
	GroupCount                  int                            `json:"GroupCount"`
	PlayoffType                 enums.LeaguePlayoffType        `json:"PlayoffType"`
	PlayoffParticipantCount     int                            `json:"PlayoffParticipantCount"`
	PlayoffByesCount            int                            `json:"PlayoffByesCount"`
	PlayoffSeedingType          enums.LeaguePlayoffSeedingType `json:"PlayoffSeedingType"`
	AllowTransfers              bool                           `json:"AllowTransfers"`
	TransfersCostCredits        bool                           `json:"TransfersCostCredits"`
	TransferCreditsPerWindow    int                            `json:"TransferCreditsPerWindow"`
	TransferCreditCap           int                            `json:"TransferCreditCap"`
	TransferWindowFrequencyDays int                            `json:"TransferWindowFrequencyDays"`
	TransferWindowDuration      int                            `json:"TransferWindowDuration"`
	DropCost                    int                            `json:"DropCost"`
	PickupCost                  int                            `json:"PickupCost"`
	NextTransferWindowStart     *time.Time                     `json:"NextTransferWindowStart"`
}

// Scan implements the sql.Scanner interface for GORM JSONB deserialization.
func (f *LeagueFormat) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan LeagueFormat: %v", value)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if val, ok := m["is_snake_round_draft"].(bool); ok {
		f.IsSnakeRoundDraft = val
	}
	if val, ok := m["draft_order_type"].(string); ok {
		f.DraftOrderType = enums.DraftOrderType(val)
	}
	if val, ok := m["season_type"].(string); ok {
		f.SeasonType = enums.LeagueSeasonType(val)
	}
	if val, ok := m["group_count"].(float64); ok {
		f.GroupCount = int(val)
	}
	if val, ok := m["playoff_type"].(string); ok {
		f.PlayoffType = enums.LeaguePlayoffType(val)
	}
	if val, ok := m["playoff_participant_count"].(float64); ok {
		f.PlayoffParticipantCount = int(val)
	}
	if val, ok := m["playoff_byes_count"].(float64); ok {
		f.PlayoffByesCount = int(val)
	}
	if val, ok := m["playoff_seeding_type"].(string); ok {
		f.PlayoffSeedingType = enums.LeaguePlayoffSeedingType(val)
	}
	if val, ok := m["allow_transfer"].(bool); ok {
		f.AllowTransfers = val
	}
	if val, ok := m["transfers_cost_credits"].(bool); ok {
		f.TransfersCostCredits = val
	}
	if val, ok := m["transfer_credits_per_window"].(float64); ok {
		f.TransferCreditsPerWindow = int(val)
	}
	if val, ok := m["transfer_credit_cap"].(float64); ok {
		f.TransferCreditCap = int(val)
	}
	if val, ok := m["transfer_window_frequency_days"].(float64); ok {
		f.TransferWindowFrequencyDays = int(val)
	}
	if val, ok := m["transfer_window_duration"].(float64); ok {
		f.TransferWindowDuration = int(val)
	}
	if val, ok := m["drop_cost"].(float64); ok {
		f.DropCost = int(val)
	}
	if val, ok := m["pickup_cost"].(float64); ok {
		f.PickupCost = int(val)
	}
	if val, ok := m["next_transfer_window_start"].(string); ok {
		t, err := time.Parse(time.RFC3339, val)
		if err == nil {
			f.NextTransferWindowStart = &t
		}
	}

	return nil
}

// Value implements the driver.Valuer interface for GORM JSONB serialization.
func (f LeagueFormat) Value() (driver.Value, error) {
	m := map[string]any{
		"is_snake_round_draft":           f.IsSnakeRoundDraft,
		"draft_order_type":               f.DraftOrderType,
		"season_type":                    f.SeasonType,
		"group_count":                    f.GroupCount,
		"playoff_type":                   f.PlayoffType,
		"playoff_participant_count":      f.PlayoffParticipantCount,
		"playoff_byes_count":             f.PlayoffByesCount,
		"playoff_seeding_type":           f.PlayoffSeedingType,
		"allow_trading":                  f.AllowTransfers,
		"allow_transfer_credits":         f.TransfersCostCredits,
		"transfer_credits_per_window":    f.TransferCreditsPerWindow,
		"transfer_credit_cap":            f.TransferCreditCap,
		"transfer_window_frequency_days": f.TransferWindowFrequencyDays,
		"transfer_window_duration":       f.TransferWindowDuration,
		"drop_cost":                      f.DropCost,
		"pickup_cost":                    f.PickupCost,
		"next_transfer_window_start":     f.NextTransferWindowStart,
	}
	return json.Marshal(m)
}
