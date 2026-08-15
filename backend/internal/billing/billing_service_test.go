package billing_test

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCalculateTaxAmount(t *testing.T) {
	tests := []struct {
		name      string
		unitPrice int64
		quantity  float64
		taxType   model.TaxType
		taxRate   float64
		want      int64
	}{
		{
			name:      "外税10% - 1000円 × 1個",
			unitPrice: 1000,
			quantity:  1,
			taxType:   model.TaxTypeExcluded,
			taxRate:   0.10,
			want:      100,
		},
		{
			name:      "外税10% - 1000円 × 2.5個",
			unitPrice: 1000,
			quantity:  2.5,
			taxType:   model.TaxTypeExcluded,
			taxRate:   0.10,
			want:      250,
		},
		{
			name:      "内税10% - 1100円 × 1個（税額は100円）",
			unitPrice: 1100,
			quantity:  1,
			taxType:   model.TaxTypeIncluded,
			taxRate:   0.10,
			want:      100,
		},
		{
			name:      "内税8% - 1080円 × 1個（税額は80円）",
			unitPrice: 1080,
			quantity:  1,
			taxType:   model.TaxTypeIncluded,
			taxRate:   0.08,
			want:      80,
		},
		{
			name:      "非課税 - 税額は0",
			unitPrice: 5000,
			quantity:  3,
			taxType:   model.TaxTypeExempt,
			taxRate:   0.10,
			want:      0,
		},
		{
			name:      "外税 - 端数四捨五入",
			unitPrice: 333,
			quantity:  1,
			taxType:   model.TaxTypeExcluded,
			taxRate:   0.10,
			want:      33, // 33.3 → 33
		},
		{
			name:      "不正な課税区分 - 税額は0",
			unitPrice: 1000,
			quantity:  1,
			taxType:   "invalid",
			taxRate:   0.10,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := billing.CalculateTaxAmount(tt.unitPrice, tt.quantity, tt.taxType, tt.taxRate)
			if got != tt.want {
				t.Errorf("CalculateTaxAmount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateBillingTotals(t *testing.T) {
	tests := []struct {
		name            string
		items           []model.BillingItem
		wantSubtotal    int64
		wantTaxTotal    int64
		wantTotalAmount int64
	}{
		{
			name: "外税のみ - subtotal + 外税 = totalAmount",
			items: []model.BillingItem{
				{UnitPrice: 1000, Quantity: 2, TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
			},
			wantSubtotal:    2000,
			wantTaxTotal:    200,
			wantTotalAmount: 2200,
		},
		{
			name: "内税のみ - totalAmount = subtotal（内税は価格に含まれる）",
			items: []model.BillingItem{
				{UnitPrice: 1100, Quantity: 1, TaxType: model.TaxTypeIncluded, TaxRate: 0.10},
			},
			wantSubtotal:    1100,
			wantTaxTotal:    100,
			wantTotalAmount: 1100,
		},
		{
			name: "非課税のみ",
			items: []model.BillingItem{
				{UnitPrice: 500, Quantity: 3, TaxType: model.TaxTypeExempt, TaxRate: 0},
			},
			wantSubtotal:    1500,
			wantTaxTotal:    0,
			wantTotalAmount: 1500,
		},
		{
			name: "外税 + 内税 + 非課税の混在",
			items: []model.BillingItem{
				{UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10}, // subtotal=1000, tax=100, excluded=100
				{UnitPrice: 1100, Quantity: 1, TaxType: model.TaxTypeIncluded, TaxRate: 0.10}, // subtotal=1100, tax=100, excluded=0
				{UnitPrice: 500, Quantity: 1, TaxType: model.TaxTypeExempt, TaxRate: 0},       // subtotal=500,  tax=0,   excluded=0
			},
			wantSubtotal:    2600,
			wantTaxTotal:    200,
			wantTotalAmount: 2700, // 2600 + 外税100のみ
		},
		{
			name: "#85 外税 + 割引額 - 割引後の課税ベースで計算",
			items: []model.BillingItem{
				{UnitPrice: 1000, Quantity: 2, DiscountAmount: 500, TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
			},
			wantSubtotal:    1500, // 2000 - 500
			wantTaxTotal:    150,  // 1500 * 10%
			wantTotalAmount: 1650, // 1500 + 外税150
		},
		{
			name: "#85 内税 + 割引額 - 割引後の税込価格から内数",
			items: []model.BillingItem{
				{UnitPrice: 1100, Quantity: 1, DiscountAmount: 100, TaxType: model.TaxTypeIncluded, TaxRate: 0.10},
			},
			wantSubtotal:    1000, // 1100 - 100
			wantTaxTotal:    91,   // 1000 * 10% 内数
			wantTotalAmount: 1000, // 内税は加算なし
		},
		{
			name: "#85 割引額が小計を超過 - 0 にクランプ",
			items: []model.BillingItem{
				{UnitPrice: 500, Quantity: 1, DiscountAmount: 800, TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
			},
			wantSubtotal:    0,
			wantTaxTotal:    0,
			wantTotalAmount: 0,
		},
		{
			name:            "明細なし",
			items:           []model.BillingItem{},
			wantSubtotal:    0,
			wantTaxTotal:    0,
			wantTotalAmount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubtotal, gotTaxTotal, gotTotalAmount := billing.CalculateBillingTotals(tt.items)
			if gotSubtotal != tt.wantSubtotal {
				t.Errorf("subtotal = %d, want %d", gotSubtotal, tt.wantSubtotal)
			}
			if gotTaxTotal != tt.wantTaxTotal {
				t.Errorf("taxTotal = %d, want %d", gotTaxTotal, tt.wantTaxTotal)
			}
			if gotTotalAmount != tt.wantTotalAmount {
				t.Errorf("totalAmount = %d, want %d", gotTotalAmount, tt.wantTotalAmount)
			}
		})
	}
}
