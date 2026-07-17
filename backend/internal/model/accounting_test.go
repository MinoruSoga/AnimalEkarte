package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBillingItem_CalculateTaxAmount pins the #85 discount-then-tax calculation:
// tax base is (unit_price * quantity - discount_amount), clamped at 0.
func TestBillingItem_CalculateTaxAmount(t *testing.T) {
	tests := []struct {
		name string
		item BillingItem
		want int64
	}{
		{
			name: "excluded tax: subtotal * rate",
			item: BillingItem{UnitPrice: 1000, Quantity: 2, DiscountAmount: 0, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 200, // (1000*2 - 0) * 0.10
		},
		{
			name: "excluded tax with discount reduces base",
			item: BillingItem{UnitPrice: 1000, Quantity: 2, DiscountAmount: 500, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 150, // (2000 - 500) * 0.10
		},
		{
			name: "included tax: subtotal * rate / (1+rate)",
			item: BillingItem{UnitPrice: 1100, Quantity: 1, DiscountAmount: 0, TaxType: TaxTypeIncluded, TaxRate: 0.10},
			want: 100, // 1100 * 0.10 / 1.10
		},
		{
			name: "exempt tax always zero",
			item: BillingItem{UnitPrice: 1000, Quantity: 3, DiscountAmount: 0, TaxType: TaxTypeExempt, TaxRate: 0.10},
			want: 0,
		},
		{
			name: "discount larger than subtotal clamps base at zero",
			item: BillingItem{UnitPrice: 100, Quantity: 1, DiscountAmount: 500, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 0,
		},
		{
			name: "unknown tax type defaults to zero",
			item: BillingItem{UnitPrice: 1000, Quantity: 1, DiscountAmount: 0, TaxType: TaxType("unknown"), TaxRate: 0.10},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			assert.Equal(t, tt.want, item.CalculateTaxAmount())
		})
	}
}
