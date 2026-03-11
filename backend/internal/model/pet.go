package model

import (
	"time"

	"github.com/google/uuid"
)

type PetSpecies string

const (
	PetSpeciesDog   PetSpecies = "犬"
	PetSpeciesCat   PetSpecies = "猫"
	PetSpeciesBird  PetSpecies = "鳥"
	PetSpeciesOther PetSpecies = "その他"
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
	ID               uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID          uuid.UUID        `gorm:"type:uuid;not null"                             json:"owner_id"`
	PetNumber        string           `gorm:"default:''"                                     json:"pet_number"`
	Name             string           `gorm:"not null"                                       json:"name"`
	PetNameKana      string           `gorm:"default:''"                                     json:"pet_name_kana"`
	Species          PetSpecies       `gorm:"type:pet_species;not null"                      json:"species"`
	Gender           PetGender        `gorm:"type:pet_gender;default:'不明'"                   json:"gender"`
	Status           PetStatus        `gorm:"type:pet_status;default:'生存'"                   json:"status"`
	BirthDate        *time.Time       `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Breed            string           `gorm:"default:''"                                     json:"breed"`
	Color            string           `gorm:"default:''"                                     json:"color"`
	Weight           string           `gorm:"default:''"                                     json:"weight"`
	NeuteredDate     *time.Time       `gorm:"type:date"                                      json:"neutered_date,omitempty"`
	AcquisitionType  *AcquisitionType `gorm:"type:acquisition_type"                          json:"acquisition_type,omitempty"`
	DangerLevel      DangerLevel      `gorm:"type:danger_level;default:'低'"                  json:"danger_level"`
	Food             string           `gorm:"default:''"                                     json:"food"`
	Environment      string           `gorm:"default:''"                                     json:"environment"`
	Phone            string           `gorm:"default:''"                                     json:"phone"`
	LastVisit        *time.Time       `gorm:"type:date"                                      json:"last_visit,omitempty"`
	InsuranceID      *uuid.UUID       `gorm:"type:uuid"                                      json:"insurance_id,omitempty"`
	InsuranceName    string           `gorm:"default:''"                                     json:"insurance_name"`
	InsuranceDetails string           `gorm:"default:''"                                     json:"insurance_details"`
	Remarks          string           `gorm:"default:''"                                     json:"remarks"`
	CreatedAt        time.Time        `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time        `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner     Owner      `gorm:"foreignKey:OwnerID"     json:"owner,omitempty"`
	Insurance *Insurance `gorm:"foreignKey:InsuranceID" json:"insurance,omitempty"`
}

func (Pet) TableName() string { return "pets" }
