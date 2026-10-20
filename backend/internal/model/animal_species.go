package model

import (
	"time"
)

// AnimalSpecies はペット種類マスタ（v12.0 pet_species ENUM廃止→マスタテーブル化）
type AnimalSpecies struct {
	ID        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Name      string    `gorm:"column:name;type:text;not null" json:"name"`
	IsActive  bool      `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	SortOrder *int      `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (AnimalSpecies) TableName() string { return "animal_species" }
