package models

import (
	"gorm.io/gorm"
	"time"
)

type PokemonSpecies struct {
	ID        int64     `gorm:"primaryKey;uniqueIndex" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Types     []string  `gorm:"type:jsonb" json:"types"`
	Abilities []Ability `gorm:"type:jsonb" json:"abilities"`
	Stats     BaseStats `gorm:"type:jsonb" json:"stats"`
	Sprites   Sprites   `gorm:"type:jsonb" json:"sprites"`

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
