package enums

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"
)

// enum types
type LeagueStatus string
type LeagueSeasonType string
type LeaguePlayoffType string
type LeaguePlayoffSeedingType string

const (
	LeagueStatusPending           LeagueStatus = "PENDING"
	LeagueStatusSetup             LeagueStatus = "SETUP"
	LeagueStatusDrafting          LeagueStatus = "DRAFTING"
	LeagueStatusPostDraft         LeagueStatus = "POST_DRAFT"
	LeagueStatusTransferWindow    LeagueStatus = "TRANSFER_WINDOW"
	LeagueStatusRegularSeason     LeagueStatus = "REGULAR_SEASON"
	LeagueStatusPostRegularSeason LeagueStatus = "POST_REGULAR_SEASON"
	LeagueStatusPlayoffs          LeagueStatus = "PLAYOFFS"
	LeagueStatusCompleted         LeagueStatus = "COMPLETED"
	LeagueStatusCancelled         LeagueStatus = "CANCELLED"
)

const (
	LeagueSeasonTypeRoundRobinOnly LeagueSeasonType = "ROUND_ROBIN_ONLY"
	LeagueSeasonTypeBracketOnly    LeagueSeasonType = "BRACKET_ONLY"
	LeagueSeasonTypeHybrid         LeagueSeasonType = "HYBRID"
)

const (
	LeaguePlayoffTypeNone       LeaguePlayoffType = "NONE"
	LeaguePlayoffTypeSingleElim LeaguePlayoffType = "SINGLE_ELIM"
	LeaguePlayoffTypeDoubleElim LeaguePlayoffType = "DOUBLE_ELIM"
)

const (
	LeaguePlayoffSeedingTypeStandard LeaguePlayoffSeedingType = "STANDARD"
	LeaguePlayoffSeedingTypeSeeded   LeaguePlayoffSeedingType = "SEEDED"
	LeaguePlayoffSeedingTypeByesOnly LeaguePlayoffSeedingType = "BYES_ONLY"
)

// ------------------------
//  Enum Related Functions
// ------------------------

// LeagueStatus stuff
var LeagueStatuses = []LeagueStatus{
	LeagueStatusPending,
	LeagueStatusSetup,
	LeagueStatusDrafting,
	LeagueStatusPostDraft,
	LeagueStatusTransferWindow,
	LeagueStatusRegularSeason,
	LeagueStatusPostRegularSeason,
	LeagueStatusPlayoffs,
	LeagueStatusCompleted,
	LeagueStatusCancelled,
}

func (ls LeagueStatus) IsValid() bool {
	return slices.Contains(LeagueStatuses, ls)
}

// Stringer() interface implementation in case it's needed
func (ls LeagueStatus) String() string {
	return string(ls)
}

// Value() implements the driver.Valuer interface for GORM/database saving.
// Tells GORM how to convert the custom type into a database-compatible type (string).
func (ls LeagueStatus) Value() (driver.Value, error) {
	if !ls.IsValid() {
		return nil, fmt.Errorf("invalid LeagueStatus value: %s", ls)
	}
	return string(ls), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// This tells GORM how to convert the database string back into the custom type.
func (ls *LeagueStatus) Scan(value any) error {
	if value == nil {
		*ls = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("LeagueStatus: expected string, got %T", value)
	}

	// Capitalize to keep everything normalized
	newStatus := LeagueStatus(strings.ToUpper(str))
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid LeagueStatus value retrieved from DB: %s", str)
	}
	*ls = newStatus
	return nil
}

// LeagueSeasonTypes stuff
var LeagueSeasonTypes = []LeagueSeasonType{
	LeagueSeasonTypeRoundRobinOnly,
	LeagueSeasonTypeBracketOnly,
	LeagueSeasonTypeHybrid,
}

func (st LeagueSeasonType) IsValid() bool {
	return slices.Contains(LeagueSeasonTypes, st)
}

// Stringer() interface implementation in case it's needed
func (st LeagueSeasonType) String() string {
	return string(st)
}

// Value() implements the driver.Valuer interface for GORM/database saving.
// Tells GORM how to convert the custom type into a database-compatible type (string).
func (st LeagueSeasonType) Value() (driver.Value, error) {
	if !st.IsValid() {
		return nil, fmt.Errorf("invalid LeagueSeasonType value: %s", st)
	}
	return string(st), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// tells GORM how to convert the database string back into the custom type.
func (st *LeagueSeasonType) Scan(value any) error {
	if value == nil {
		*st = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("LeagueSeasonType: expected string, got %T", value)
	}

	// Capitalize to keep everything normalized
	newSeasonType := LeagueSeasonType(strings.ToUpper(str))
	if !newSeasonType.IsValid() {
		return fmt.Errorf("invalid LeagueSeasonType value retrieved from DB: %s", str)
	}
	*st = newSeasonType
	return nil
}

//
// LeaguePlayoffType stuff
//

var LeaguePlayoffTypes = []LeaguePlayoffType{
	LeaguePlayoffTypeNone,
	LeaguePlayoffTypeSingleElim,
	LeaguePlayoffTypeDoubleElim,
}

func (pt LeaguePlayoffType) IsValid() bool {
	return slices.Contains(LeaguePlayoffTypes, pt)
}

// Stringer() interface implementation in case it's needed
func (pt LeaguePlayoffType) String() string {
	return string(pt)
}

// Value() implements the driver.Valuer interface for GORM/database saving.
// Tells GORM how to convert the custom type into a database-compatible type (string).
func (pt LeaguePlayoffType) Value() (driver.Value, error) {
	if !pt.IsValid() {
		return nil, fmt.Errorf("invalid LeaguePlayoffType value: %s", pt)
	}
	return string(pt), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// Tells GORM how to convert the database string back into the custom type.
func (pt *LeaguePlayoffType) Scan(value any) error {
	if value == nil {
		*pt = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("LeaguePlayoffType: expected string, got %T", value)
	}

	// Capitalize to keep everything normalized
	newPlayoffType := LeaguePlayoffType(strings.ToUpper(str))
	if !newPlayoffType.IsValid() {
		return fmt.Errorf("invalid LeaguePlayoffType value retrieved from DB: %s", str)
	}
	*pt = newPlayoffType
	return nil
}

//
// LeaguePlayoffSeedingType stuff
//

var LeaguePlayoffSeedingTypes = []LeaguePlayoffSeedingType{
	LeaguePlayoffSeedingTypeStandard,
	LeaguePlayoffSeedingTypeSeeded,
	LeaguePlayoffSeedingTypeByesOnly,
}

func (p LeaguePlayoffSeedingType) IsValid() bool {
	return slices.Contains(LeaguePlayoffSeedingTypes, p)
}

// Stringer() interface implementation in case it's needed
func (p LeaguePlayoffSeedingType) String() string {
	return string(p)
}

// Value() implements the driver.Valuer interface for GORM/database saving.
// Tells GORM how to convert the custom type into a database-compatible type (string).
func (p LeaguePlayoffSeedingType) Value() (driver.Value, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("invalid LeaguePlayoffSeedingType value: %s", p)
	}
	return string(p), nil
}

// Scan implements the sql.Scanner interface for GORM/database loading.
// Tells GORM how to convert the database string back into the custom type.
func (p *LeaguePlayoffSeedingType) Scan(value any) error {
	if value == nil {
		*p = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("LeaguePlayoffSeedingType: expected string, got %T", value)
	}

	// Capitalize to keep everything normalized
	newPlayoffSeedingType := LeaguePlayoffSeedingType(strings.ToUpper(str))
	if !newPlayoffSeedingType.IsValid() {
		return fmt.Errorf("invalid LeaguePlayoffType value retrieved from DB: %s", str)
	}
	*p = newPlayoffSeedingType
	return nil
}
