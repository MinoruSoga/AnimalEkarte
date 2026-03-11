package model

import (
	"time"

	"github.com/google/uuid"
)

// Pet はペット（患者）テーブルに対応するモデル
type Pet struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID          uuid.UUID  `gorm:"column:owner_id;type:uuid;not null" json:"owner_id"`
	OwnerName        string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetNumber        *string    `gorm:"column:pet_number" json:"pet_number,omitempty"`
	Name             string     `gorm:"column:name;not null" json:"name"`
	PetNameKana      *string    `gorm:"column:pet_name_kana" json:"pet_name_kana,omitempty"`
	Species          string     `gorm:"column:species;type:pet_species;not null" json:"species"`
	Gender           *string    `gorm:"column:gender;type:pet_gender" json:"gender,omitempty"`
	Status           string     `gorm:"column:status;type:pet_status;default:'生存'" json:"status"`
	BirthDate        *time.Time `gorm:"column:birth_date" json:"birth_date,omitempty"`
	Breed            *string    `gorm:"column:breed" json:"breed,omitempty"`
	Color            *string    `gorm:"column:color" json:"color,omitempty"`
	Weight           *string    `gorm:"column:weight" json:"weight,omitempty"`
	NeuteredDate     *time.Time `gorm:"column:neutered_date" json:"neutered_date,omitempty"`
	AcquisitionType  *string    `gorm:"column:acquisition_type;type:acquisition_type" json:"acquisition_type,omitempty"`
	DangerLevel      *string    `gorm:"column:danger_level;type:danger_level" json:"danger_level,omitempty"`
	Food             *string    `gorm:"column:food" json:"food,omitempty"`
	Environment      *string    `gorm:"column:environment" json:"environment,omitempty"`
	Phone            *string    `gorm:"column:phone" json:"phone,omitempty"`
	LastVisit        *time.Time `gorm:"column:last_visit" json:"last_visit,omitempty"`
	InsuranceName    *string    `gorm:"column:insurance_name" json:"insurance_name,omitempty"`
	InsuranceDetails *string    `gorm:"column:insurance_details" json:"insurance_details,omitempty"`
	Remarks          *string    `gorm:"column:remarks" json:"remarks,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreatePetRequest はペット作成リクエスト
type CreatePetRequest struct {
	OwnerID          string  `json:"owner_id" binding:"required"`
	OwnerName        string  `json:"owner_name" binding:"required"`
	PetNumber        *string `json:"pet_number"`
	Name             string  `json:"name" binding:"required"`
	PetNameKana      *string `json:"pet_name_kana"`
	Species          string  `json:"species" binding:"required"`
	Gender           *string `json:"gender"`
	Status           string  `json:"status"`
	BirthDate        *string `json:"birth_date"`
	Breed            *string `json:"breed"`
	Color            *string `json:"color"`
	Weight           *string `json:"weight"`
	NeuteredDate     *string `json:"neutered_date"`
	AcquisitionType  *string `json:"acquisition_type"`
	DangerLevel      *string `json:"danger_level"`
	Food             *string `json:"food"`
	Environment      *string `json:"environment"`
	Phone            *string `json:"phone"`
	InsuranceName    *string `json:"insurance_name"`
	InsuranceDetails *string `json:"insurance_details"`
	Remarks          *string `json:"remarks"`
}

// UpdatePetRequest はペット更新リクエスト
type UpdatePetRequest struct {
	OwnerID          *string `json:"owner_id"`
	OwnerName        *string `json:"owner_name"`
	PetNumber        *string `json:"pet_number"`
	Name             *string `json:"name"`
	PetNameKana      *string `json:"pet_name_kana"`
	Species          *string `json:"species"`
	Gender           *string `json:"gender"`
	Status           *string `json:"status"`
	BirthDate        *string `json:"birth_date"`
	Breed            *string `json:"breed"`
	Color            *string `json:"color"`
	Weight           *string `json:"weight"`
	NeuteredDate     *string `json:"neutered_date"`
	AcquisitionType  *string `json:"acquisition_type"`
	DangerLevel      *string `json:"danger_level"`
	Food             *string `json:"food"`
	Environment      *string `json:"environment"`
	Phone            *string `json:"phone"`
	InsuranceName    *string `json:"insurance_name"`
	InsuranceDetails *string `json:"insurance_details"`
	Remarks          *string `json:"remarks"`
}
