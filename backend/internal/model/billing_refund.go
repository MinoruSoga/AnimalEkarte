package model

import "time"

type BillingRefund struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" tygo:"type: number"`
	ClinicID   uint64    `gorm:"not null"`
	BillingID  uint64    `gorm:"not null"`
	Amount     int64     `gorm:"not null"` // 返金額（正の整数、円）
	Reason     string    `gorm:"not null;default:''"`
	RefundedAt time.Time `gorm:"not null;autoCreateTime"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (BillingRefund) TableName() string { return "billing_refunds" }
