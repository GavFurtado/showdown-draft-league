package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type PokemonSpecies struct {
	ID        int64          `gorm:"primaryKey;uniqueIndex;column:id" json:"ID"`
	DexID     int64          `gorm:"index;not null;column:dex_id" json:"DexID"`
	Name      string         `gorm:"not null;column:name" json:"Name"`
	Types     StringArray    `gorm:"type:jsonb;column:types" json:"Types"`
	Abilities AbilitiesArray `gorm:"type:jsonb;column:abilities" json:"Abilities"`
	Stats     BaseStats      `gorm:"type:jsonb;column:stats" json:"Stats"`
	Sprites   Sprites        `gorm:"type:jsonb;column:sprites" json:"Sprites"`

	CreatedAt time.Time      `json:"CreatedAt" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"UpdatedAt" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
}

type BaseStats struct {
	Hp             int `gorm:"column:hp" json:"Hp"`
	Attack         int `gorm:"column:attack" json:"Attack"`
	Defense        int `gorm:"column:defense" json:"Defense"`
	SpecialAttack  int `gorm:"column:special_attack" json:"SpecialAttack"`
	SpecialDefense int `gorm:"column:special_defense" json:"SpecialDefense"`
	Speed          int `gorm:"column:speed" json:"Speed"`
}
type Sprites struct {
	FrontDefault    string `gorm:"column:front_default" json:"FrontDefault"`
	OfficialArtwork string `gorm:"column:official_artwork" json:"OfficialArtwork"`
}

// Ability struct remains the same
type Ability struct {
	Name     string `gorm:"column:name" json:"Name"`
	IsHidden bool   `gorm:"default:false;column:is_hidden" json:"IsHidden"`
}

// Value implements the driver.Valuer interface for BaseStats.
func (bs BaseStats) Value() (driver.Value, error) {
	m := map[string]int{
		"hp":               bs.Hp,
		"attack":           bs.Attack,
		"defense":          bs.Defense,
		"special_attack":   bs.SpecialAttack,
		"special_defense":  bs.SpecialDefense,
		"speed":            bs.Speed,
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for BaseStats.
func (bs *BaseStats) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	var m map[string]int
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	bs.Hp = m["hp"]
	bs.Attack = m["attack"]
	bs.Defense = m["defense"]
	bs.SpecialAttack = m["special_attack"]
	bs.SpecialDefense = m["special_defense"]
	bs.Speed = m["speed"]
	return nil
}

// Value implements the driver.Valuer interface for Sprites.
func (s Sprites) Value() (driver.Value, error) {
	m := map[string]string{
		"front_default":    s.FrontDefault,
		"official_artwork": s.OfficialArtwork,
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for Sprites.
func (s *Sprites) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	var m map[string]string
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	s.FrontDefault = m["front_default"]
	s.OfficialArtwork = m["official_artwork"]
	return nil
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

// AbilitiesArray is a custom type for handling JSONB arrays of Ability structs
type AbilitiesArray []Ability

// Value implements the driver.Valuer interface for AbilitiesArray.
func (aa AbilitiesArray) Value() (driver.Value, error) {
	if aa == nil {
		return nil, nil
	}
	var dbAbilities []map[string]interface{}
	for _, a := range aa {
		dbAbilities = append(dbAbilities, map[string]interface{}{
			"name":      a.Name,
			"is_hidden": a.IsHidden,
		})
	}
	return json.Marshal(dbAbilities)
}

// Scan implements the sql.Scanner interface for AbilitiesArray.
func (aa *AbilitiesArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	var dbAbilities []map[string]interface{}
	if err := json.Unmarshal(bytes, &dbAbilities); err != nil {
		return err
	}
	*aa = make(AbilitiesArray, len(dbAbilities))
	for i, dbA := range dbAbilities {
		name, _ := dbA["name"].(string)
		isHidden, _ := dbA["is_hidden"].(bool)
		(*aa)[i] = Ability{
			Name:     name,
			IsHidden: isHidden,
		}
	}
	return nil
}
