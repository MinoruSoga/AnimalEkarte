package model

import (
	"time"

	"github.com/google/uuid"
)

type TrimmingStatus string

const (
	TrimmingStatusCompleted  TrimmingStatus = "完了"
	TrimmingStatusReserved   TrimmingStatus = "予約"
	TrimmingStatusInProgress TrimmingStatus = "進行中"
)

type BodyWeightUnit string

const (
	BodyWeightUnitKg BodyWeightUnit = "Kg"
	BodyWeightUnitG  BodyWeightUnit = "g"
)

type TrimmingRecord struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Date           time.Time      `gorm:"type:date;not null"                             json:"date"`
	PetID          uuid.UUID      `gorm:"type:uuid;not null"                             json:"pet_id"`
	PetNumber      string         `gorm:"not null;default:''"                            json:"pet_number"`
	PetName        string         `gorm:"not null;default:''"                            json:"pet_name"`
	OwnerName      string         `gorm:"not null;default:''"                            json:"owner_name"`
	Species        PetSpecies     `gorm:"type:pet_species;not null"                      json:"species"`
	Weight         string         `gorm:"default:''"                                     json:"weight"`
	StyleRequest   string         `gorm:"default:''"                                     json:"style_request"`
	StaffID        uuid.UUID      `gorm:"type:uuid;not null"                             json:"staff_id"`
	Status         TrimmingStatus `gorm:"type:trimming_status;default:'予約'"              json:"status"`
	CourseID       *uuid.UUID     `gorm:"type:uuid"                                      json:"course_id,omitempty"`
	BW             string         `gorm:"default:''"                                     json:"bw"`
	BWUnit         BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'"             json:"bw_unit"`
	BT             string         `gorm:"default:''"                                     json:"bt"`
	UsedShampoo    string         `gorm:"default:''"                                     json:"used_shampoo"`
	UsedRibbon     string         `gorm:"default:''"                                     json:"used_ribbon"`
	Remarks        string         `gorm:"default:''"                                     json:"remarks"`
	StyleImage     string         `gorm:"default:''"                                     json:"style_image"`
	CompletedImage string         `gorm:"default:''"                                     json:"completed_image"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Pet     Pet              `gorm:"foreignKey:PetID"    json:"pet,omitempty"`
	Staff   Staff            `gorm:"foreignKey:StaffID"  json:"staff,omitempty"`
	Course  *TrimmingCourse  `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Options []TrimmingOption `gorm:"many2many:trimming_record_options;joinForeignKey:TrimmingRecordID;joinReferences:OptionID" json:"options,omitempty"`
}

func (TrimmingRecord) TableName() string { return "trimming_records" }

type TrimmingRecordOption struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TrimmingRecordID uuid.UUID `gorm:"type:uuid;not null"                             json:"trimming_record_id"`
	OptionID         uuid.UUID `gorm:"type:uuid;not null"                             json:"option_id"`
	SortOrder        int       `gorm:"default:0"                                      json:"sort_order"`
}

func (TrimmingRecordOption) TableName() string { return "trimming_record_options" }
