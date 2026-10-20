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
	ID                uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	CloseDate         time.Time       `gorm:"column:close_date;type:date;not null" json:"close_date"`
	Period            string          `gorm:"column:period;type:varchar(3);not null" json:"period"`
	TheoreticalCash   int64           `gorm:"column:theoretical_cash;type:bigint;not null;default:0" json:"theoretical_cash"`
	ActualCash        int64           `gorm:"column:actual_cash;type:bigint;not null;default:0" json:"actual_cash"`
	CategoryBreakdown json.RawMessage `gorm:"column:category_breakdown;type:jsonb;not null;default:'{}'" json:"category_breakdown"`
	Memo              string          `gorm:"column:memo;type:text;not null;default:''" json:"memo"`
	ClosedBy          *uint64         `gorm:"column:closed_by;type:bigint" json:"closed_by,omitempty"`
	ClosedAt          time.Time       `gorm:"column:closed_at;type:timestamptz;not null;default:now()" json:"closed_at"`
	CreatedAt         time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	ClosedByStaff     *Staff          `gorm:"foreignKey:ClosedBy;references:ID" json:"closed_by_staff,omitempty" tygo:"-"`
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
