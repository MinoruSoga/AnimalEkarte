package model

import (
	"time"

	"github.com/google/uuid"
)

// AnimalSpecies はペット種類マスタ（v12.0 pet_species ENUM廃止→マスタテーブル化）
type AnimalSpecies struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string       `gorm:"default:''"                                     json:"code"`
	Name      string       `gorm:"not null"                                       json:"name"`
	Status    MasterStatus `gorm:"type:master_status;default:'active'"            json:"status"`
	SortOrder int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (AnimalSpecies) TableName() string { return "animal_species" }
