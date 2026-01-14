package model

import (
	"time"

	"github.com/google/uuid"
)

// Clinic クリニック情報モデル
type Clinic struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name               string    `json:"name" gorm:"type:varchar(100)"`
	BranchName         string    `json:"branch_name" gorm:"type:varchar(100)"`
	PostalCode         string    `json:"postal_code" gorm:"type:varchar(10)"`
	Address            string    `json:"address" gorm:"type:text"`
	PhoneNumber        string    `json:"phone_number" gorm:"type:varchar(20)"`
	FaxNumber          string    `json:"fax_number" gorm:"type:varchar(20)"`
	RegistrationNumber string    `json:"registration_number" gorm:"type:varchar(50)"`
	DirectorName       string    `json:"director_name" gorm:"type:varchar(100)"`
	Email              string    `json:"email" gorm:"type:varchar(255)"`
	Website            string    `json:"website" gorm:"type:varchar(255)"`
	LogoURL            string    `json:"logo_url" gorm:"type:text"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relations
	Staffs []Staff `json:"staffs,omitempty" gorm:"foreignKey:ClinicID"`
}

// TableName テーブル名を指定
func (Clinic) TableName() string {
	return "clinics"
}

// Staff スタッフモデル
type Staff struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ClinicID  *uuid.UUID `json:"clinic_id" gorm:"type:uuid"`
	Name      string     `json:"name" gorm:"type:varchar(100)"`
	Role      string     `json:"role" gorm:"type:varchar(50)"` // veterinarian, nurse, groomer, admin
	Email     string     `json:"email" gorm:"type:varchar(255)"`
	Phone     string     `json:"phone" gorm:"type:varchar(20)"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relations
	Clinic *Clinic `json:"clinic,omitempty" gorm:"foreignKey:ClinicID"`
}

// TableName テーブル名を指定
func (Staff) TableName() string {
	return "staffs"
}
