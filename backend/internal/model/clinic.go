package model

import "time"

type Clinic struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	CompanyID          uint64    `gorm:"not null"                                       json:"company_id"`
	Name               string    `gorm:"not null"                                       json:"name"`
	PostalCode         string    `gorm:"default:''"                                     json:"postal_code"`
	Address            string    `gorm:"default:''"                                     json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                     json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                     json:"fax_number"`
	RegistrationNumber string    `gorm:"default:''"                                     json:"registration_number"`
	DirectorName       string    `gorm:"default:''"                                     json:"director_name"`
	Email              string    `gorm:"default:''"                                     json:"email"`
	Website            string    `gorm:"default:''"                                     json:"website"`
	LogoURL            string    `gorm:"default:''"                                     json:"logo_url"`
	IsActive           bool      `gorm:"default:true"                                   json:"is_active"`
	StandardTaxRate    float64   `gorm:"type:numeric;not null;default:0.10"             json:"standard_tax_rate"`
	ReducedTaxRate     float64   `gorm:"type:numeric;not null;default:0.08"             json:"reduced_tax_rate"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Clinic) TableName() string { return "clinics" }
