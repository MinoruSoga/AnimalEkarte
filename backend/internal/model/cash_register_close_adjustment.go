package model

import "time"

// CashRegisterCloseAdjustment はレジ締め後の会計訂正台帳（W-013 append-only）。
// close レコード自体の reverse/取消は productize しない。締め後の会計編集は本テーブルへ追記し、
// 監査ログ（AuditActionBillingPostCloseEdit）と同一 transaction で fail-closed に記録する。
type CashRegisterCloseAdjustment struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID           uint64    `gorm:"not null"                 json:"clinic_id"`
	CloseID            uint64    `gorm:"not null"                 json:"close_id"`
	BillingID          uint64    `gorm:"not null"                 json:"billing_id"`
	AccountingDelta    int64     `gorm:"not null;default:0"       json:"accounting_delta"`
	CashMovementAmount int64     `gorm:"not null;default:0"       json:"cash_movement_amount"`
	Reason             string    `gorm:"type:text;not null"       json:"reason"`
	ActorID            *uint64   `                                json:"actor_id,omitempty"`
	ExecutedAt         time.Time `gorm:"not null;default:now()"   json:"executed_at"`
	CreatedAt          time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (CashRegisterCloseAdjustment) TableName() string { return "cash_register_close_adjustments" }
