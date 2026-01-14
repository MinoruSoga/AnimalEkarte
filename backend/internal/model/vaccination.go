package model

import (
	"time"

	"github.com/google/uuid"
)

// Vaccination ワクチン接種記録モデル
type Vaccination struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PetID           uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null;index:idx_vac_pet_id"`
	OwnerID         uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	DoctorID        *uuid.UUID `json:"doctor_id" gorm:"type:uuid"`
	VaccineMasterID *uuid.UUID `json:"vaccine_master_id" gorm:"type:uuid"`
	VaccineName     string     `json:"vaccine_name" gorm:"type:varchar(100)"`
	VaccinationDate time.Time  `json:"vaccination_date" gorm:"type:date"`
	NextDate        *time.Time `json:"next_date" gorm:"type:date;index:idx_vac_next_date"`
	LotNumber       string     `json:"lot_number" gorm:"type:varchar(50)"`
	Notes           string     `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Pet   *Pet   `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner *Owner `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を指定
func (Vaccination) TableName() string {
	return "vaccinations"
}
