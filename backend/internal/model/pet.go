package model

import (
	"gorm.io/gorm"
	"time"
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

// PetNeuteredStatus は避妊・去勢の施術状況を表す。
// 値は性別中立（オスの去勢・メスの避妊を done で共通に表し、表示層がラベルを分ける）。
type PetNeuteredStatus string

const (
	PetNeuteredStatusUnknown PetNeuteredStatus = "unknown"
	PetNeuteredStatusNotDone PetNeuteredStatus = "not_done"
	PetNeuteredStatusDone    PetNeuteredStatus = "done"
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
	ID            uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID      uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID       uint64         `gorm:"column:owner_id;type:bigint;not null" json:"owner_id"`
	PetNumber     string         `gorm:"column:pet_number;type:text;not null;default:''" json:"pet_number"`
	InsuranceID   *uint64        `gorm:"column:insurance_id;type:bigint" json:"insurance_id,omitempty"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Version       int            `gorm:"column:version;type:integer;not null;default:1" json:"version"`
	Owner         *Owner         `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID"          json:"owner,omitempty"`
	Insurance     *Insurance     `gorm:"foreignKey:InsuranceID;references:ID"      json:"insurance,omitempty"`
	AnimalSpecies *AnimalSpecies `gorm:"-"  json:"animal_species,omitempty"`

	// Basic / BasicSubject は commonbasic が現行 basic revision から埋める読み取り専用の一時フィールド。
	// DB 列ではなく、nil の間は基本情報が未解決（旧 pets 列へのフォールバックはしない）。
	Basic        *CommonPetBasicRevision `gorm:"-" json:"-"`
	BasicSubject *CommonSubjectRef       `gorm:"-" json:"-"`
}

func (Pet) TableName() string { return "pets" }
