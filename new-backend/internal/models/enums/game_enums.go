package enums

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type GameStatus string

const (
	GameStatusPending   GameStatus = "pending"
	GameStatusCompleted GameStatus = "completed"
	GameStatusDisputed  GameStatus = "disputed"
)

var gameStatuses = []GameStatus{
	GameStatusPending,
	GameStatusCompleted,
	GameStatusDisputed,
}

// IsValid checks if the GameStatus is one of the predefined valid statuses.
func (gs GameStatus) IsValid() bool {
	for _, status := range gameStatuses {
		if gs == status {
			return true
		}
	}
	return false
}

// Value implements the driver.Valuer interface for GORM/database saving.
// This tells GORM how to convert the custom type into a database-compatible type (string).
func (gs GameStatus) Value() (driver.Value, error) {
	if !gs.IsValid() {
		return nil, fmt.Errorf("invalid GameStatus value: %s", gs)
	}
	return string(gs), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (gs *GameStatus) Scan(value any) error {
	if value == nil {
		*gs = GameStatusPending // Default or zero value for nil
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
	*gs = newStatus
	return nil
}

func (gs GameStatus) Normalize() GameStatus {
	return GameStatus(strings.ToUpper(string(gs)))
}
