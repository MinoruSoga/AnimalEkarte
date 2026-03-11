package model

import (
	"time"

	"github.com/google/uuid"
)

// Owner は飼主（顧客）テーブルに対応するモデル
type Owner struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerName      string     `gorm:"column:owner_name;not null" json:"owner_name"`
	OwnerNameKana  *string    `gorm:"column:owner_name_kana" json:"owner_name_kana,omitempty"`
	Company        string     `gorm:"column:company;default:''" json:"company"`
	PostalCode     string     `gorm:"column:postal_code;default:''" json:"postal_code"`
	Address1       string     `gorm:"column:address1;default:''" json:"address1"`
	Address2       string     `gorm:"column:address2;default:''" json:"address2"`
	HomePostalCode string     `gorm:"column:home_postal_code;default:''" json:"home_postal_code"`
	HomeAddress1   string     `gorm:"column:home_address1;default:''" json:"home_address1"`
	HomeAddress2   string     `gorm:"column:home_address2;default:''" json:"home_address2"`
	BirthDate      *time.Time `gorm:"column:birth_date" json:"birth_date,omitempty"`
	Phone          string     `gorm:"column:phone;default:''" json:"phone"`
	CompanyPhone   string     `gorm:"column:company_phone;default:''" json:"company_phone"`
	Email          string     `gorm:"column:email;default:''" json:"email"`
	Remarks        string     `gorm:"column:remarks;default:''" json:"remarks"`
	IsDangerous    bool       `gorm:"column:is_dangerous;default:false" json:"is_dangerous"`
	DiscountRate   float64    `gorm:"column:discount_rate;type:numeric(5,2);default:0" json:"discount_rate"`
	MembershipType string     `gorm:"column:membership_type;type:membership_type;default:'非会員'" json:"membership_type"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Pets []Pet `gorm:"foreignKey:OwnerID" json:"pets,omitempty"`
}

// CreateOwnerRequest は飼主作成リクエスト
type CreateOwnerRequest struct {
	OwnerName      string   `json:"owner_name" binding:"required"`
	OwnerNameKana  *string  `json:"owner_name_kana"`
	Company        string   `json:"company"`
	PostalCode     string   `json:"postal_code"`
	Address1       string   `json:"address1"`
	Address2       string   `json:"address2"`
	HomePostalCode string   `json:"home_postal_code"`
	HomeAddress1   string   `json:"home_address1"`
	HomeAddress2   string   `json:"home_address2"`
	Phone          string   `json:"phone"`
	CompanyPhone   string   `json:"company_phone"`
	Email          string   `json:"email"`
	Remarks        string   `json:"remarks"`
	IsDangerous    bool     `json:"is_dangerous"`
	DiscountRate   float64  `json:"discount_rate"`
	MembershipType string   `json:"membership_type"`
}

// UpdateOwnerRequest は飼主更新リクエスト（全フィールドoptional）
type UpdateOwnerRequest struct {
	OwnerName      *string  `json:"owner_name"`
	OwnerNameKana  *string  `json:"owner_name_kana"`
	Company        *string  `json:"company"`
	PostalCode     *string  `json:"postal_code"`
	Address1       *string  `json:"address1"`
	Address2       *string  `json:"address2"`
	HomePostalCode *string  `json:"home_postal_code"`
	HomeAddress1   *string  `json:"home_address1"`
	HomeAddress2   *string  `json:"home_address2"`
	Phone          *string  `json:"phone"`
	CompanyPhone   *string  `json:"company_phone"`
	Email          *string  `json:"email"`
	Remarks        *string  `json:"remarks"`
	IsDangerous    *bool    `json:"is_dangerous"`
	DiscountRate   *float64 `json:"discount_rate"`
	MembershipType *string  `json:"membership_type"`
}
