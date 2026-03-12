package model

import (
	"time"

	"github.com/google/uuid"
)

// Company は法人情報（ノア動物病院、シングルトン）
type Company struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string    `gorm:"not null"                                       json:"name"`
	BranchName         string    `gorm:"default:''"                                     json:"branch_name"`
	PostalCode         string    `gorm:"default:''"                                     json:"postal_code"`
	Address            string    `gorm:"default:''"                                     json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                     json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                     json:"fax_number"`
	Email              string    `gorm:"default:''"                                     json:"email"`
	Website            string    `gorm:"default:''"                                     json:"website"`
	DirectorName       string    `gorm:"default:''"                                     json:"director_name"`
	RegistrationNumber string    `gorm:"default:''"                                     json:"registration_number"`
	LogoURL            string    `gorm:"default:''"                                     json:"logo_url"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Company) TableName() string { return "company" }
