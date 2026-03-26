package model

import (
	"time"
)

// AnimalSpecies はペット種類マスタ（v12.0 pet_species ENUM廃止→マスタテーブル化）
type AnimalSpecies struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Name      string    `gorm:"not null"                                       json:"name"`
	IsActive  bool      `gorm:"default:true"                                   json:"is_active"`
	SortOrder int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (AnimalSpecies) TableName() string { return "animal_species" }
