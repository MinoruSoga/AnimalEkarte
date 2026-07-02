package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEstimateItem_CalculateTaxAmount mirrors BillingItem's tax calc but without a
// discount field: tax base is simply unit_price * quantity.
func TestEstimateItem_CalculateTaxAmount(t *testing.T) {
	tests := []struct {
		name string
		item EstimateItem
		want int64
	}{
		{
			name: "excluded tax: subtotal * rate",
			item: EstimateItem{UnitPrice: 1000, Quantity: 2, TaxType: TaxTypeExcluded, TaxRate: 0.10},
			want: 200,
		},
		{
			name: "included tax: subtotal * rate / (1+rate)",
			item: EstimateItem{UnitPrice: 1100, Quantity: 1, TaxType: TaxTypeIncluded, TaxRate: 0.10},
			want: 100,
		},
		{
			name: "exempt tax always zero",
			item: EstimateItem{UnitPrice: 1000, Quantity: 3, TaxType: TaxTypeExempt, TaxRate: 0.10},
			want: 0,
		},
		{
			name: "unknown tax type defaults to zero",
			item: EstimateItem{UnitPrice: 1000, Quantity: 1, TaxType: TaxType("unknown"), TaxRate: 0.10},
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
