package model

import (
	"time"
)

// PetOwner はペットと副飼主の追加紐付けを表す。
type PetOwner struct {
	ID           uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID     uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PetID        uint64    `gorm:"column:pet_id;type:bigint;not null;uniqueIndex:pet_owners_pet_id_owner_id_key,priority:1" json:"pet_id"`
	OwnerID      uint64    `gorm:"column:owner_id;type:bigint;not null;uniqueIndex:pet_owners_pet_id_owner_id_key,priority:2" json:"owner_id"`
	Relationship string    `gorm:"column:relationship;type:text;not null;default:''" json:"relationship"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (PetOwner) TableName() string { return "pet_owners" }
