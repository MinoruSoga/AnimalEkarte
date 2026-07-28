package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEstimateItem_CalculateTaxAmount pins #85 discount-then-tax (aligned with BillingItem / MDL-01).
func TestEstimateItem_CalculateTaxAmount(t *testing.T) {
	tests := []struct {
		name string
		item EstimateItem
		want int64
	}{
		{
			name: "excluded tax: subtotal * rate",
			item: EstimateItem{UnitPrice: 1000, Quantity: 2, DiscountAmount: 0, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 200,
		},
		{
			name: "excluded tax: discount reduces taxable base",
			item: EstimateItem{UnitPrice: 1000, Quantity: 2, DiscountAmount: 500, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 150,
		},
		{
			name: "included tax: subtotal * rate / (1+rate)",
			item: EstimateItem{UnitPrice: 1100, Quantity: 1, DiscountAmount: 0, TaxType: TaxTypeIncluded, TaxRate: 0.10},
			want: 100,
		},
		{
			name: "exempt tax always zero",
			item: EstimateItem{UnitPrice: 1000, Quantity: 3, DiscountAmount: 0, TaxType: TaxTypeExempt, TaxRate: 0.10},
			want: 0,
		},
		{
			name: "discount exceeding subtotal floors taxable base at zero",
			item: EstimateItem{UnitPrice: 100, Quantity: 1, DiscountAmount: 500, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 0,
		},
		{
			name: "unknown tax type defaults to zero",
			item: EstimateItem{UnitPrice: 1000, Quantity: 1, DiscountAmount: 0, TaxType: TaxType("unknown"), TaxRate: 0.10},
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
