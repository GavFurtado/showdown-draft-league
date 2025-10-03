package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type PokemonSpecies struct {
	ID        int64          `gorm:"primaryKey;uniqueIndex" json:"id"`
	DexID     int64          `gorm:"index;not null" json:"dex_id"`
	Name      string         `gorm:"not null" json:"name"`
	Types     StringArray    `gorm:"type:jsonb" json:"types"`
	Abilities AbilitiesArray `gorm:"type:jsonb" json:"abilities"`
	Stats     BaseStats      `gorm:"type:jsonb" json:"stats"`
	Sprites   Sprites        `gorm:"type:jsonb" json:"sprites"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type BaseStats struct {
	Hp             int `gorm:"not null" json:"hp"`
	Attack         int `gorm:"not null" json:"attack"`
	Defense        int `gorm:"not null" json:"defense"`
	SpecialAttack  int `gorm:"not null" json:"special_attack"`
	SpecialDefense int `gorm:"not null" json:"special_defense"`
	Speed          int `gorm:"not null" json:"speed"`
}
type Sprites struct {
	FrontDefault    string `json:"front_default"`
	OfficialArtwork string `json:"official_artwork"`
}

// Value implements the driver.Valuer interface for BaseStats.
func (bs BaseStats) Value() (driver.Value, error) {
	return json.Marshal(bs)
}

// Scan implements the sql.Scanner interface for BaseStats.
func (bs *BaseStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, bs)
}

// Value implements the driver.Valuer interface for Sprites.
func (s Sprites) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for Sprites.
func (s *Sprites) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// StringArray is a custom type for handling JSONB arrays of strings
type StringArray []string

// Value implements the driver.Valuer interface for StringArray.
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements the sql.Scanner interface for StringArray.
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sa)
}

// Ability struct remains the same
type Ability struct {
	Name     string `json:"name"`
	IsHidden bool   `gorm:"default:false" json:"is_hidden"`
}

// AbilitiesArray is a custom type for handling JSONB arrays of Ability structs
type AbilitiesArray []Ability

// Value implements the driver.Valuer interface for AbilitiesArray.
func (aa AbilitiesArray) Value() (driver.Value, error) {
	if aa == nil {
		return nil, nil
	}
	return json.Marshal(aa)
}

// Scan implements the sql.Scanner interface for AbilitiesArray.
func (aa *AbilitiesArray) Scan(value interface{}) error {
	if value == nil {
		*aa = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, aa)
}
