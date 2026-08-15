package model

import (
	"time"

	"gorm.io/gorm"
)

type PetStatus string

const (
	PetStatusAlive    PetStatus = "alive"
	PetStatusDeceased PetStatus = "deceased"
)

type PetGender string

const (
	PetGenderMale    PetGender = "male"
	PetGenderFemale  PetGender = "female"
	PetGenderUnknown PetGender = "unknown"
)

type AcquisitionType string

const (
	AcquisitionTypePurchase AcquisitionType = "purchased"
	AcquisitionTypeTransfer AcquisitionType = "transferred"
	AcquisitionTypeRescued  AcquisitionType = "rescued"
	AcquisitionTypeOther    AcquisitionType = "other"
)

type DangerLevel string

const (
	DangerLevelLow    DangerLevel = "low"
	DangerLevelMedium DangerLevel = "medium"
	DangerLevelHigh   DangerLevel = "high"
)

type Pet struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64           `gorm:"not null"                                       json:"clinic_id"`
	OwnerID         uint64           `gorm:"not null"                                       json:"owner_id"`
	AnimalSpeciesID uint64           `gorm:"not null"                                       json:"animal_species_id"`
	PetNumber       string           `gorm:"default:''"                                     json:"pet_number"`
	Name            string           `gorm:"not null"                                       json:"name"`
	NameKana        string           `gorm:"column:name_kana;default:''"                    json:"name_kana"`
	Gender          PetGender        `gorm:"type:pet_gender;default:'unknown'"               json:"gender"`
	Status          PetStatus        `gorm:"type:pet_status;default:'alive'"                 json:"status"`
	BirthDate       *time.Time       `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Breed           string           `gorm:"default:''"                                     json:"breed"`
	Color           string           `gorm:"default:''"                                     json:"color"`
	BloodType       *string          `gorm:"column:blood_type"                              json:"blood_type,omitempty"`
	MicrochipNumber *string          `gorm:"column:microchip_number"                        json:"microchip_number,omitempty"`
	Weight          *float64         `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	NeuteredDate    *time.Time       `gorm:"type:date"                                      json:"neutered_date,omitempty"`
	AcquisitionType *AcquisitionType `gorm:"type:acquisition_type"                          json:"acquisition_type,omitempty"`
	DangerLevel     DangerLevel      `gorm:"type:danger_level;default:'low'"                 json:"danger_level"`
	DangerReason    *string          `gorm:"column:danger_reason"                           json:"danger_reason,omitempty"`
	Food            string           `gorm:"default:''"                                     json:"food"`
	Environment     string           `gorm:"default:''"                                     json:"environment"`
	Phone           string           `gorm:"default:''"                                     json:"phone"`
	LastVisit       *time.Time       `gorm:"type:date"                                      json:"last_visit,omitempty"`
	InsuranceID     *uint64          `                                                      json:"insurance_id,omitempty"`
	Remarks         string           `gorm:"default:''"                                     json:"remarks"`
	DeceasedAt      *time.Time       `gorm:"column:deceased_at"                             json:"deceased_at,omitempty"`
	DeceasedReason  *string          `gorm:"column:deceased_reason"                         json:"deceased_reason,omitempty"`
	Version         int              `gorm:"default:1"                                      json:"version"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `                                                      json:"-"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Insurance     *Insurance     `gorm:"foreignKey:InsuranceID"      json:"insurance,omitempty"`
	AnimalSpecies *AnimalSpecies `gorm:"foreignKey:AnimalSpeciesID"  json:"animal_species,omitempty"`
}

func (Pet) TableName() string { return "pets" }
