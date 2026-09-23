package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestPatientOutstanding_AbsInsuranceAmount は EMR-62:
// insurance_amount / discount_amount の正規契約は正の magnitude。旧クライアント由来の
// 負値レガシー行でも ABS 解釈し、未収残高を過大計上しないことを固定する。
func TestPatientOutstanding_AbsInsuranceAmount(t *testing.T) {
	tests := []struct {
		name    string
		payment *model.Payment
		want    int64
	}{
		{
			name:    "nil payment は 0",
			payment: nil,
			want:    0,
		},
		{
			name: "正の保険額: due=1100-500, paid=600 → 未収 0",
			payment: &model.Payment{
				TotalAmount:     1100,
				InsuranceAmount: 500,
				BillingAmount:   600,
			},
			want: 0,
		},
		{
			name: "負の保険額レガシー行も magnitude 解釈: due=1100-500, paid=600 → 未収 0",
			payment: &model.Payment{
				TotalAmount:     1100,
				InsuranceAmount: -500,
				BillingAmount:   600,
			},
			want: 0,
		},
		{
			name: "負値を素通しすると due=1600 となり過大計上する回帰防止",
			payment: &model.Payment{
				TotalAmount:     1100,
				InsuranceAmount: -500,
				DiscountAmount:  -100,
				BillingAmount:   500,
			},
			// due = 1100 - 500 - 100 = 500, residual = 0
			want: 0,
		},
		{
			name: "クレジット訂正で支払額が due 未満: 残差を未収として返す",
			payment: &model.Payment{
				TotalAmount:     1100,
				InsuranceAmount: 0,
				BillingAmount:   400,
			},
			want: 700,
		},
		{
			name: "過払い（residual 負）は 0 に丸める",
			payment: &model.Payment{
				TotalAmount:   1100,
				BillingAmount: 1500,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, patientOutstanding(tt.payment))
		})
	}
}
