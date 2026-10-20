package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-454: 元診療明細（treatment）に保存済みの保険設定は、後から master を
// 引き直さず来歴どおり未請求候補へ引き継ぐ。未保存・旧欠落（nil）は
// false（保険対象）へ補完しない。
func TestTreatmentToUnbilledBillingItem_InsuranceSetting(t *testing.T) {
	isInsurance := true
	notInsurance := false

	tests := []struct {
		name  string
		input *bool
		want  *bool
	}{
		{name: "unset stays nil", input: nil, want: nil},
		{name: "explicit applicable is preserved", input: &isInsurance, want: &isInsurance},
		{name: "explicit not applicable is preserved", input: &notInsurance, want: &notInsurance},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := treatmentToUnbilledBillingItem(&model.Treatment{
				ID:          9,
				Content:     "薬剤A",
				UnitPrice:   500,
				Quantity:    1,
				IsInsurance: tt.input,
			})

			if tt.want == nil {
				assert.Nil(t, item.IsInsuranceApplicable)
				return
			}
			require.NotNil(t, item.IsInsuranceApplicable)
			assert.Equal(t, *tt.want, *item.IsInsuranceApplicable)
		})
	}
}
