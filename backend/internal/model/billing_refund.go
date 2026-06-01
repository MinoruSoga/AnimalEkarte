package model

import "time"

type BillingRefund struct {
	ID         uint64  `gorm:"primaryKey;autoIncrement" json:"id"          tygo:"type: number"`
	ClinicID   uint64  `gorm:"not null"                 json:"clinic_id"`
	BillingID  uint64  `gorm:"not null"                 json:"billing_id"`
	Amount     int64   `gorm:"not null"                 json:"amount"` // 返金額（正の整数、円）
	Reason     string  `gorm:"not null;default:''"      json:"reason"`
	RefundedBy *uint64 `gorm:""                         json:"refunded_by"` // 返金処理スタッフID（nullable）
	// PaymentMethod は返金先の支払手段（nullable）。混在会計でどの手段へ返金したか記録する。
	// PaymentMethodID と dual maintain する（payment_splits と同パターン）。記録のみ・Phase 1 では未指定可。
	PaymentMethod   *PaymentMethod `gorm:"type:payment_method" json:"payment_method,omitempty"`
	PaymentMethodID *uint64        `gorm:""                    json:"payment_method_id,omitempty"`
	RefundedAt      time.Time      `gorm:"not null;autoCreateTime"  json:"refunded_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"           json:"created_at"`

	// Relations
	RefundedByStaff *Staff `gorm:"foreignKey:RefundedBy" json:"refunded_by_staff,omitempty" tygo:"-"`
}

func (BillingRefund) TableName() string { return "billing_refunds" }
