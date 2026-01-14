package model

import (
	"time"

	"github.com/google/uuid"
)

// Pet ペットモデル
type Pet struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OwnerID          uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null;index:idx_pets_owner_id"`
	PetNumber        string     `json:"pet_number" gorm:"type:varchar(20);uniqueIndex:idx_pets_pet_number"`
	Name             string     `json:"name" gorm:"type:varchar(100);not null"`
	Species          string     `json:"species" gorm:"type:varchar(50);not null"`
	Breed            string     `json:"breed" gorm:"type:varchar(100)"`
	Gender           string     `json:"gender" gorm:"type:varchar(10)"`
	BirthDate        *time.Time `json:"birth_date" gorm:"type:date"`
	Weight           *float64   `json:"weight" gorm:"type:decimal(5,2)"`
	MicrochipID      string     `json:"microchip_id" gorm:"type:varchar(50)"`
	Environment      string     `json:"environment" gorm:"type:varchar(50)"`
	Status           string     `json:"status" gorm:"type:varchar(10);default:'生存'"`
	InsuranceName    string     `json:"insurance_name" gorm:"type:varchar(100)"`
	InsuranceDetails string     `json:"insurance_details" gorm:"type:text"`
	LastVisit        *time.Time `json:"last_visit" gorm:"type:date"`
	Notes            string     `json:"notes" gorm:"type:text"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relations
	Owner          *Owner          `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	MedicalRecords []MedicalRecord `json:"medical_records,omitempty" gorm:"foreignKey:PetID"`
}

// TableName テーブル名を指定
func (Pet) TableName() string {
	return "pets"
}

// CreatePetRequest ペット作成リクエスト
type CreatePetRequest struct {
	OwnerID          string  `json:"owner_id" binding:"required"`
	PetNumber        string  `json:"pet_number"`
	Name             string  `json:"name" binding:"required"`
	Species          string  `json:"species" binding:"required"`
	Breed            string  `json:"breed"`
	Gender           string  `json:"gender"`
	BirthDate        string  `json:"birth_date"`
	Weight           float64 `json:"weight"`
	MicrochipID      string  `json:"microchip_id"`
	Environment      string  `json:"environment"`
	Status           string  `json:"status"`
	InsuranceName    string  `json:"insurance_name"`
	InsuranceDetails string  `json:"insurance_details"`
	Notes            string  `json:"notes"`
}

// UpdatePetRequest ペット更新リクエスト
type UpdatePetRequest struct {
	OwnerID          string  `json:"owner_id"`
	PetNumber        string  `json:"pet_number"`
	Name             string  `json:"name"`
	Species          string  `json:"species"`
	Breed            string  `json:"breed"`
	Gender           string  `json:"gender"`
	BirthDate        string  `json:"birth_date"`
	Weight           float64 `json:"weight"`
	MicrochipID      string  `json:"microchip_id"`
	Environment      string  `json:"environment"`
	Status           string  `json:"status"`
	InsuranceName    string  `json:"insurance_name"`
	InsuranceDetails string  `json:"insurance_details"`
	Notes            string  `json:"notes"`
}
