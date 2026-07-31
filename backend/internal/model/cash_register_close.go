package model

import (
	"encoding/json"
	"time"
)

// CashRegisterClose はレジ締めレコード。
// W-013 FINAL B: append-only。app は Update/Delete/soft-delete しない。
// DB の deleted_at 列は migration 003 以降未使用（完全 UNIQUE と immutability trigger で再オープン不可）。
// 締め後の会計訂正は CashRegisterCloseAdjustment への追記で表現する（close 自体の reverse は productize しない）。
type CashRegisterClose struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement"          json:"id"`
	ClinicID          uint64          `gorm:"not null;uniqueIndex:uq_cash_register_closes_date_period" json:"clinic_id"`
	CloseDate         time.Time       `gorm:"type:date;not null;uniqueIndex:uq_cash_register_closes_date_period" json:"close_date"`
	Period            string          `gorm:"type:varchar(3);not null;uniqueIndex:uq_cash_register_closes_date_period" json:"period"` // "am", "pm", or "emg"
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
	// UnclassifiedOtherCount は category=other 明細を1件以上持つ会計の distinct 件数（DEC-40）。
	// ポインタ + omitempty により旧 snapshot ではフィールド欠落（FE は「記録なし」表示）。
	// 0 件の場合もポインタ先 0 として保存し、欠落と区別する。
	UnclassifiedOtherCount *int64 `json:"unclassified_other_count,omitempty"`
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
