package model

import (
	"time"
)

// BillingReviewStatus は会計医師確認ステータス
type BillingReviewStatus string

const (
	BillingReviewStatusPending   BillingReviewStatus = "pending"
	BillingReviewStatusConfirmed BillingReviewStatus = "confirmed"
	BillingReviewStatusReturned  BillingReviewStatus = "returned"
)

// BillingReview は会計医師確認（v7.0追加）
type BillingReview struct {
	ID              uint64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64              `gorm:"not null;uniqueIndex"                           json:"medical_record_id"`
	Status          BillingReviewStatus `gorm:"type:billing_review_status;default:'pending'"   json:"status"`
	ConfirmedBy     *uint64             `                                                      json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time          `gorm:"column:confirmed_at"                            json:"confirmed_at,omitempty"`
	ReturnedBy      *uint64             `                                                      json:"returned_by,omitempty"`
	ReturnedAt      *time.Time          `gorm:"column:returned_at"                             json:"returned_at,omitempty"`
	ReturnReason    string              `gorm:"default:''"                                     json:"return_reason"`
	Memo            string              `gorm:"default:''"                                     json:"memo"`
	CreatedAt       time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord  *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	ConfirmedStaff *Staff         `gorm:"foreignKey:ConfirmedBy"     json:"confirmed_staff,omitempty"`
	ReturnedStaff  *Staff         `gorm:"foreignKey:ReturnedBy"      json:"returned_staff,omitempty"`
}

func (BillingReview) TableName() string { return "billing_reviews" }
