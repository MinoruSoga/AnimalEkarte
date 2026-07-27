package model

import "time"

// PetOwner はペットと副飼主の追加紐付けを表す。
type PetOwner struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"         json:"id"`
	ClinicID     uint64    `gorm:"not null"                         json:"clinic_id"`
	PetID        uint64    `gorm:"not null"                         json:"pet_id"`
	OwnerID      uint64    `gorm:"not null"                         json:"owner_id"`
	Relationship string    `gorm:"type:text;not null;default:''"    json:"relationship"`
	CreatedAt    time.Time `gorm:"autoCreateTime"                   json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                   json:"updated_at"`
}

func (PetOwner) TableName() string { return "pet_owners" }
