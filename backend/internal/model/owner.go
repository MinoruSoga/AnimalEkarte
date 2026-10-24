package model

import (
	"gorm.io/gorm"
	"time"
)

type MembershipType string

const (
	MembershipTypeNonMember   MembershipType = "non_member"
	MembershipTypeMember      MembershipType = "member"
	MembershipTypeDeceased    MembershipType = "deceased"
	MembershipTypeTransferred MembershipType = "transferred"
)

type Owner struct {
	ID                     uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID               uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Remarks                string         `gorm:"column:remarks;type:text;not null;default:''" json:"remarks"`
	IsDangerous            bool           `gorm:"column:is_dangerous;type:boolean;not null;default:false" json:"is_dangerous"`
	DiscountRate           ExactNumeric   `gorm:"column:discount_rate;type:numeric(5,2);not null;default:0" json:"discount_rate"`
	MembershipType         MembershipType `gorm:"column:membership_type;type:membership_type;not null;default:'non_member'" json:"membership_type"`
	LstepOptOut            bool           `gorm:"column:lstep_opt_out;type:boolean;not null;default:false" json:"lstep_opt_out"`
	LstepOptOutAt          *time.Time     `gorm:"column:lstep_opt_out_at;type:timestamptz" json:"lstep_opt_out_at,omitempty"`
	LstepOptOutReason      *string        `gorm:"column:lstep_opt_out_reason;type:text" json:"lstep_opt_out_reason,omitempty"`
	DeliveryExcluded       bool           `gorm:"column:delivery_excluded;type:boolean;not null;default:false" json:"delivery_excluded"`
	DeliveryExcludedReason *string        `gorm:"column:delivery_excluded_reason;type:varchar(100)" json:"delivery_excluded_reason,omitempty"`
	IsTransferred          bool           `gorm:"column:is_transferred;type:boolean;not null;default:false" json:"is_transferred"`
	TransferAt             *time.Time     `gorm:"column:transfer_at;type:timestamptz" json:"transfer_at,omitempty"`
	DeliveryCaution        bool           `gorm:"column:delivery_caution;type:boolean;not null;default:false" json:"delivery_caution"`
	DeliveryCautionReason  *string        `gorm:"column:delivery_caution_reason;type:varchar(100)" json:"delivery_caution_reason,omitempty"`
	DMPreference           *bool          `gorm:"column:dm_preference;type:boolean" json:"dm_preference,omitempty"`
	CreatedAt              time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	Pets                   []Pet          `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID" json:"pets,omitempty"`

	// Basic / BasicSubject は commonbasic が現行 basic revision から埋める読み取り専用の一時フィールド。
	// DB 列ではなく、nil の間は基本情報が未解決（旧 owners 列へのフォールバックはしない）。
	Basic        *CommonOwnerBasicRevision `gorm:"-" json:"-"`
	BasicSubject *CommonSubjectRef         `gorm:"-" json:"-"`
}

func (Owner) TableName() string { return "owners" }
