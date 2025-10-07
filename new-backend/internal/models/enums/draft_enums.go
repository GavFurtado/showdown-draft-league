package enums

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// DraftStatus defines the possible states of a draft.
type DraftStatus string

const (
	DraftStatusPending   DraftStatus = "PENDING"
	DraftStatusOngoing   DraftStatus = "ONGOING"
	DraftStatusPaused    DraftStatus = "PAUSED"
	DraftStatusCompleted DraftStatus = "COMPLETED"
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
