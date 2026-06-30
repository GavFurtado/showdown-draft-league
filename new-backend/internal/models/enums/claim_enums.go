package enums

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// ClaimSource defines how a Pokemon was acquired by a player.
// A claim's Source determines which ID (if any) the SourceID field references.
// Source is set at creation time and never changes.
type ClaimSource string

const (
	ClaimSourceDraft     ClaimSource = "draft"
	ClaimSourceFreeAgent ClaimSource = "free_agent"
)

func (cs ClaimSource) IsValid() bool {
	switch cs {
	case ClaimSourceDraft, ClaimSourceFreeAgent:
		return true
	}
	return false
}

func (cs ClaimSource) String() string {
	return string(cs)
}

// Value implements the driver.Valuer interface.
func (cs ClaimSource) Value() (driver.Value, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf("invalid ClaimSource value: %s", cs)
	}
	return string(cs), nil
}

// Scan implements the sql.Scanner interface.
func (cs *ClaimSource) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("ClaimSource: expected string, got nil")
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("ClaimSource: expected string, got %T", value)
	}

	parsed := ClaimSource(strings.ToLower(str))
	if !parsed.IsValid() {
		return fmt.Errorf("invalid ClaimSource value: %s", str)
	}
	*cs = parsed
	return nil
}
