package model

import (
	"encoding/json"
	"time"
)

// CashRegisterClose はレジ締めレコード
type CashRegisterClose struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement"          json:"id"`
	ClinicID          uint64          `gorm:"not null"                          json:"clinic_id"`
	CloseDate         time.Time       `gorm:"type:date;not null"                json:"close_date"`
	Period            string          `gorm:"type:varchar(2);not null"          json:"period"` // "am" or "pm"
	TheoreticalCash   int64           `gorm:"not null;default:0"                json:"theoretical_cash"`
	ActualCash        int64           `gorm:"not null;default:0"                json:"actual_cash"`
	CashDifference    int64           `gorm:"not null;default:0"                json:"cash_difference"`
	CategoryBreakdown json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"  json:"category_breakdown"`
	Memo              string          `gorm:"not null;default:''"               json:"memo"`
	ClosedBy          *uint64         `                                         json:"closed_by,omitempty"`
	ClosedAt          time.Time       `gorm:"not null;default:now()"            json:"closed_at"`
	CreatedAt         time.Time       `gorm:"autoCreateTime"                    json:"created_at"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime"                    json:"updated_at"`

	// Relations
	ClosedByStaff *Staff `gorm:"foreignKey:ClosedBy" json:"closed_by_staff,omitempty" tygo:"-"`
}

func (CashRegisterClose) TableName() string { return "cash_register_closes" }

// CategoryBreakdownSchema は category_breakdown JSONB の型定義
type CategoryBreakdownSchema struct {
	Categories   map[string]map[string]int64 `json:"categories"` // category → payment_method_name → amount
	TaxBreakdown TaxBreakdown                `json:"tax_breakdown"`
}

// TaxBreakdown は消費税区分別の集計
type TaxBreakdown struct {
	Standard TaxBreakdownItem `json:"standard"` // 10%
	Reduced  TaxBreakdownItem `json:"reduced"`  // 8%
}

// TaxBreakdownItem は課税額と税額のペア
type TaxBreakdownItem struct {
	TaxableAmount int64 `json:"taxable_amount"`
	TaxAmount     int64 `json:"tax_amount"`
}
