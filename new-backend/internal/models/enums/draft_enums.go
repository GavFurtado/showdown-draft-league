package enums

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// DraftStatus defines the possible states of a draft.
type DraftStatus string

// DraftOrderType defines the possible methods for determining draft order.
type DraftOrderType string

const (
	DraftStatusPending   DraftStatus = "PENDING"
	DraftStatusOngoing   DraftStatus = "ONGOING"
	DraftStatusPaused    DraftStatus = "PAUSED"
	DraftStatusCompleted DraftStatus = "COMPLETED"
)

const (
	DraftOrderTypeRandom DraftOrderType = "RANDOM"
	DraftOrderTypeManual DraftOrderType = "MANUAL"
)

// Validate DraftStatus for database interactions
func (ds DraftStatus) IsValid() bool {
	switch ds {
	case DraftStatusPending, DraftStatusOngoing, DraftStatusPaused, DraftStatusCompleted:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface for GORM/database saving.
func (ds DraftStatus) Value() (driver.Value, error) {
	if !ds.IsValid() {
		return nil, fmt.Errorf("invalid DraftStatus value: %s", ds)
	}
	return string(ds), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
func (ds *DraftStatus) Scan(value any) error {
	if value == nil {
		*ds = DraftStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("DraftStatus: expected string, got %T", value)
	}
	newStatus := DraftStatus(str).Normalize()
	if !newStatus.IsValid() {
		// Log or handle this error appropriately, as it indicates bad data in DB
		return fmt.Errorf("invalid DraftStatus value retrieved from DB: %s", str)
	}
	*ds = newStatus
	return nil
}

func (ds DraftStatus) Normalize() DraftStatus {
	return DraftStatus(strings.ToUpper(string(ds)))
}

// Validate DraftOrderType for database interactions
func (dot DraftOrderType) IsValid() bool {
	switch dot {
	case DraftOrderTypeRandom, DraftOrderTypeManual:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface for GORM/database saving.
func (dot DraftOrderType) Value() (driver.Value, error) {
	if !dot.IsValid() {
		return nil, fmt.Errorf("invalid DraftOrderType value: %s", dot)
	}
	return string(dot), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
func (dot *DraftOrderType) Scan(value any) error {
	if value == nil {
		*dot = DraftOrderTypeRandom
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("DraftOrderType: expected string, got %T", value)
	}
	newType := DraftOrderType(str).Normalize()
	if !newType.IsValid() {
		return fmt.Errorf("invalid DraftOrderType value retrieved from DB: %s", str)
	}
	*dot = newType
	return nil
}

func (dot DraftOrderType) Normalize() DraftOrderType {
	return DraftOrderType(strings.ToUpper(string(dot)))
}
