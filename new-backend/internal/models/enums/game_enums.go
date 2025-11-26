package enums

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"
)

type GameStatus string
type GameType string

const (
	GameStatusScheduled       GameStatus = "SCHEDULED"
	GameStatusApprovalPending GameStatus = "APPROVAL_PENDING"
	GameStatusCompleted       GameStatus = "COMPLETED"
	GameStatusDisputed        GameStatus = "DISPUTED"
)
const (
	// REGULAR_SEASON or HYBRID leagues
	GameTypeRegularSeason     GameType = "REGULAR_SEASON"
	GameTypePlayoffUpper      GameType = "PLAYOFF_UPPER"
	GameTypePlayoffLower      GameType = "PLAYOFF_LOWER"
	GameTypePlayoffSingleElim GameType = "PLAYOFF_SINGLEELIM"
	// BRACKET_ONLY leagues
	GameTypeTournamentSingleElim GameType = "TOURNAMENT_SINGLEELIM"
	GameTypeTournamentUpper      GameType = "TOURNAMENT_UPPER"
	GameTypeTournamentLower      GameType = "TOURNAMENT_LOWER"
)

var gameStatuses = []GameStatus{
	GameStatusScheduled,
	GameStatusApprovalPending,
	GameStatusCompleted,
	GameStatusDisputed,
}
var gameTypes = []GameType{
	GameTypePlayoffUpper,
	GameTypePlayoffLower,
	GameTypeRegularSeason,
	GameTypeTournamentSingleElim,
	GameTypeTournamentUpper,
	GameTypeTournamentLower,
}

// IsValid checks if the GameStatus is one of the predefined valid statuses.
func (gt GameStatus) IsValid() bool {
	return slices.Contains(gameStatuses, gt)
}

// Value implements the driver.Valuer interface for GORM/database saving.
// This tells GORM how to convert the custom type into a database-compatible type (string).
func (gt GameStatus) Value() (driver.Value, error) {
	if !gt.IsValid() {
		return nil, fmt.Errorf("invalid GameStatus value: %s", gt)
	}
	return string(gt), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (gt *GameStatus) Scan(value any) error {
	if value == nil {
		*gt = GameStatusScheduled // Default or zero value for nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("GameStatus: expected string, got %T", value)
	}
	// Validate the string from the database to ensure it's a known status
	newStatus := GameStatus(str).Normalize()
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid GameStatus value retrieved from DB: %s", str)
	}
	*gt = newStatus
	return nil
}

func (gt GameStatus) Normalize() GameStatus {
	return GameStatus(strings.ToUpper(string(gt)))
}

// IsValid checks if the GameStatus is one of the predefined valid statuses.
func (gt GameType) IsValid() bool {
	return slices.Contains(gameTypes, gt)
}

// Value implements the driver.Valuer interface for GORM/database saving.
// This tells GORM how to convert the custom type into a database-compatible type (string).
func (gt GameType) Value() (driver.Value, error) {
	if !gt.IsValid() {
		return nil, fmt.Errorf("invalid GameType value: %s", gt)
	}
	return string(gt), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (gt *GameType) Scan(value any) error {
	if value == nil {
		*gt = GameTypeRegularSeason // Default or zero value for nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("GameType: expected string, got %T", value)
	}
	// Validate the string from the database to ensure it's a known status
	newGameType := GameType(str).Normalize()
	if !newGameType.IsValid() {
		return fmt.Errorf("invalid GameStatus value retrieved from DB: %s", str)
	}
	*gt = newGameType
	return nil
}

func (gt GameType) Normalize() GameType {
	return GameType(strings.ToUpper(string(gt)))
}
