package models

import (
	"gorm.io/gorm"
	"time"
)

type PokemonSpecies struct {
	ID    uint     `gorm:"primaryKey;uniqueIndex" json:"id"`
	Name  string   `gorm:"not null" json:"name"`
	Types []string `gorm:"text[]" json:"types"`

	// Abilities: Stored as JSONB, GORM will marshal/unmarshal slice of structs
	Abilities []Ability `gorm:"type:jsonb" json:"abilities"`

	// Stats: Stored as JSONB, GORM will marshal/unmarshal struct
	Stats BaseStats `gorm:"type:jsonb" json:"stats"`

	// Sprites: Stored as JSONB, GORM will marshal/unmarshal struct
	Sprites Sprites `gorm:"type:jsonb" json:"sprites"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Ability struct {
	Name     string `json:"name"`
	IsHidden bool   `gorm:"default:false" json:"is_hidden"`
}

type BaseStats struct {
	Hp    int `gorm:"not null" json:"hp"`
	Att   int `gorm:"not null" json:"att"`
	Def   int `gorm:"not null" json:"def"`
	SpAtt int `gorm:"not null" json:"sp_att"`
	SpDef int `gorm:"not null" json:"sp_def"`
	Speed int `gorm:"not null" json:"speed"`
}

type Sprites struct {
	FrontDefault    string `json:"front_default"`
	OfficialArtwork string `json:"official_artwork"` // not used
}
