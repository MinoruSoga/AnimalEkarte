package model

import "time"

type BillingRefund struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"          tygo:"type: number"`
	ClinicID   uint64    `gorm:"not null"                 json:"clinic_id"`
	BillingID  uint64    `gorm:"not null"                 json:"billing_id"`
	Amount     int64     `gorm:"not null"                 json:"amount"` // 返金額（正の整数、円）
	Reason     string    `gorm:"not null;default:''"      json:"reason"`
	RefundedAt time.Time `gorm:"not null;autoCreateTime"  json:"refunded_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (BillingRefund) TableName() string { return "billing_refunds" }
