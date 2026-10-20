package model

import (
	"time"
)

// Company は法人情報（ノア動物病院、シングルトン）
type Company struct {
	ID                        uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Name                      string    `gorm:"column:name;type:text;not null" json:"name"`
	PostalCode                string    `gorm:"column:postal_code;type:text;not null;default:''" json:"postal_code"`
	Address                   string    `gorm:"column:address;type:text;not null;default:''" json:"address"`
	PhoneNumber               string    `gorm:"column:phone_number;type:text;not null;default:''" json:"phone_number"`
	FaxNumber                 string    `gorm:"column:fax_number;type:text;not null;default:''" json:"fax_number"`
	RegistrationNumber        string    `gorm:"column:registration_number;type:text;not null;default:''" json:"registration_number"`
	InvoiceRegistrationNumber string    `gorm:"column:invoice_registration_number;type:text;not null;default:''" json:"invoice_registration_number"`
	DirectorName              string    `gorm:"column:director_name;type:text;not null;default:''" json:"director_name"`
	Email                     string    `gorm:"column:email;type:text;not null;default:''" json:"email"`
	Website                   string    `gorm:"column:website;type:text;not null;default:''" json:"website"`
	LogoURL                   string    `gorm:"column:logo_url;type:text;not null;default:''" json:"logo_url"`
	CreatedAt                 time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (Company) TableName() string { return "companies" }
