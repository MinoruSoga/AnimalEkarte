package model

import (
	"time"

	"gorm.io/gorm"
)

type MembershipType string

const (
	MembershipTypeNonMember   MembershipType = "non_member"
	MembershipTypeMember      MembershipType = "member"
	MembershipTypeDeceased    MembershipType = "deceased"
	MembershipTypeTransferred MembershipType = "transferred"
)

type Owner struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name              string         `gorm:"column:name;not null"                           json:"name"`
	NameKana          string         `gorm:"column:name_kana;default:''"                    json:"name_kana"`
	BirthDate         *time.Time     `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Company           string         `gorm:"default:''"                                     json:"company"`
	PostalCode        string         `gorm:"default:''"                                     json:"postal_code"`
	Address1          string         `gorm:"default:''"                                     json:"address1"`
	Address2          string         `gorm:"default:''"                                     json:"address2"`
	HomePostalCode    string         `gorm:"default:''"                                     json:"home_postal_code"`
	HomeAddress1      string         `gorm:"default:''"                                     json:"home_address1"`
	HomeAddress2      string         `gorm:"default:''"                                     json:"home_address2"`
	Phone             string         `gorm:"default:''"                                     json:"phone"`
	CompanyPhone      string         `gorm:"default:''"                                     json:"company_phone"`
	Email             string         `gorm:"default:''"                                     json:"email"`
	Remarks           string         `gorm:"default:''"                                     json:"remarks"`
	IsDangerous       bool           `gorm:"default:false"                                  json:"is_dangerous"`
	DiscountRate      float64        `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	MembershipType    MembershipType `gorm:"type:membership_type;default:'non_member'"      json:"membership_type"`
	LineUserID        *string        `gorm:"column:line_user_id"                            json:"line_user_id,omitempty"`
	LstepOptOut       bool           `gorm:"column:lstep_opt_out;default:false"             json:"lstep_opt_out"`
	LstepOptOutAt     *time.Time     `gorm:"column:lstep_opt_out_at"                        json:"lstep_opt_out_at,omitempty"`
	LstepOptOutReason *string        `gorm:"column:lstep_opt_out_reason"                    json:"lstep_opt_out_reason,omitempty"`
	LineFollowedAt    *time.Time     `gorm:"column:line_followed_at"                        json:"line_followed_at,omitempty"`
	LineBlockedAt     *time.Time     `gorm:"column:line_blocked_at"                         json:"line_blocked_at,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt         gorm.DeletedAt `                                                      json:"-"`

	// Relations
	Pets []Pet `gorm:"foreignKey:OwnerID" json:"pets,omitempty"`
}

func (Owner) TableName() string { return "owners" }
