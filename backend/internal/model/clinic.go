package model

import (
	"time"

	"github.com/google/uuid"
)

// ClinicInfo は病院情報（シングルトン）テーブルに対応するモデル
type ClinicInfo struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string    `gorm:"column:name;not null" json:"name"`
	BranchName         string    `gorm:"column:branch_name;default:''" json:"branch_name"`
	PostalCode         string    `gorm:"column:postal_code;default:''" json:"postal_code"`
	Address            string    `gorm:"column:address;default:''" json:"address"`
	PhoneNumber        string    `gorm:"column:phone_number;default:''" json:"phone_number"`
	FaxNumber          string    `gorm:"column:fax_number;default:''" json:"fax_number"`
	RegistrationNumber string    `gorm:"column:registration_number;default:''" json:"registration_number"`
	DirectorName       string    `gorm:"column:director_name;default:''" json:"director_name"`
	Email              string    `gorm:"column:email;default:''" json:"email"`
	Website            string    `gorm:"column:website;default:''" json:"website"`
	LogoURL            *string   `gorm:"column:logo_url" json:"logo_url,omitempty"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName は GORM に対してテーブル名を明示する
func (ClinicInfo) TableName() string {
	return "clinic_info"
}

// Clinic はマルチクリニック対応テーブルに対応するモデル（認証用）
type Clinic struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string    `gorm:"column:name;not null" json:"name"`
	BranchName         string    `gorm:"column:branch_name;default:''" json:"branch_name"`
	PostalCode         string    `gorm:"column:postal_code;default:''" json:"postal_code"`
	Address            string    `gorm:"column:address;default:''" json:"address"`
	PhoneNumber        string    `gorm:"column:phone_number;default:''" json:"phone_number"`
	FaxNumber          string    `gorm:"column:fax_number;default:''" json:"fax_number"`
	RegistrationNumber string    `gorm:"column:registration_number;default:''" json:"registration_number"`
	DirectorName       string    `gorm:"column:director_name;default:''" json:"director_name"`
	Email              string    `gorm:"column:email;default:''" json:"email"`
	Website            string    `gorm:"column:website;default:''" json:"website"`
	LogoURL            *string   `gorm:"column:logo_url" json:"logo_url,omitempty"`
	IsActive           bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateClinicInfoRequest は病院情報作成リクエスト
type CreateClinicInfoRequest struct {
	Name               string  `json:"name" binding:"required"`
	BranchName         string  `json:"branch_name"`
	PostalCode         string  `json:"postal_code"`
	Address            string  `json:"address"`
	PhoneNumber        string  `json:"phone_number"`
	FaxNumber          string  `json:"fax_number"`
	RegistrationNumber string  `json:"registration_number"`
	DirectorName       string  `json:"director_name"`
	Email              string  `json:"email"`
	Website            string  `json:"website"`
	LogoURL            *string `json:"logo_url"`
}

// UpdateClinicInfoRequest は病院情報更新リクエスト
type UpdateClinicInfoRequest struct {
	Name               *string `json:"name"`
	BranchName         *string `json:"branch_name"`
	PostalCode         *string `json:"postal_code"`
	Address            *string `json:"address"`
	PhoneNumber        *string `json:"phone_number"`
	FaxNumber          *string `json:"fax_number"`
	RegistrationNumber *string `json:"registration_number"`
	DirectorName       *string `json:"director_name"`
	Email              *string `json:"email"`
	Website            *string `json:"website"`
	LogoURL            *string `json:"logo_url"`
}
