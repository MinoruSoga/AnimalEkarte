package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PetStatus string

const (
	PetStatusAlive    PetStatus = "生存"
	PetStatusDeceased PetStatus = "死亡"
)

type PetGender string

const (
	PetGenderMale    PetGender = "雄"
	PetGenderFemale  PetGender = "雌"
	PetGenderUnknown PetGender = "不明"
)

type AcquisitionType string

const (
	AcquisitionTypePurchase  AcquisitionType = "購入"
	AcquisitionTypeTransfer  AcquisitionType = "譲渡"
	AcquisitionTypeProtected AcquisitionType = "保護"
	AcquisitionTypeOther     AcquisitionType = "その他"
)

type DangerLevel string

const (
	DangerLevelLow    DangerLevel = "低"
	DangerLevelMedium DangerLevel = "中"
	DangerLevelHigh   DangerLevel = "高"
)

type Pet struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID        uuid.UUID        `gorm:"type:uuid;not null"                             json:"clinic_id"`
	OwnerID         uuid.UUID        `gorm:"type:uuid;not null"                             json:"owner_id"`
	AnimalSpeciesID uuid.UUID        `gorm:"type:uuid;not null"                             json:"animal_species_id"`
	PetNumber       string           `gorm:"default:''"                                     json:"pet_number"`
	Name            string           `gorm:"not null"                                       json:"name"`
	PetNameKana     string           `gorm:"default:''"                                     json:"pet_name_kana"`
	Gender          PetGender        `gorm:"type:pet_gender;default:'不明'"                   json:"gender"`
	Status          PetStatus        `gorm:"type:pet_status;default:'生存'"                   json:"status"`
	BirthDate       *time.Time       `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Breed           string           `gorm:"default:''"                                     json:"breed"`
	Color           string           `gorm:"default:''"                                     json:"color"`
	Weight          *float64         `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	NeuteredDate    *time.Time       `gorm:"type:date"                                      json:"neutered_date,omitempty"`
	AcquisitionType *AcquisitionType `gorm:"type:acquisition_type"                          json:"acquisition_type,omitempty"`
	DangerLevel     DangerLevel      `gorm:"type:danger_level;default:'低'"                  json:"danger_level"`
	Food            string           `gorm:"default:''"                                     json:"food"`
	Environment     string           `gorm:"default:''"                                     json:"environment"`
	Phone           string           `gorm:"default:''"                                     json:"phone"`
	LastVisit       *time.Time       `gorm:"type:date"                                      json:"last_visit,omitempty"`
	InsuranceID     *uuid.UUID       `gorm:"type:uuid"                                      json:"insurance_id,omitempty"`
	Remarks         string           `gorm:"default:''"                                     json:"remarks"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `                                                      json:"deleted_at"`

	// Relations
	Owner         Owner          `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Insurance     *Insurance     `gorm:"foreignKey:InsuranceID"      json:"insurance,omitempty"`
	AnimalSpecies *AnimalSpecies `gorm:"foreignKey:AnimalSpeciesID"  json:"animal_species,omitempty"`
}

func (Pet) TableName() string { return "pets" }
