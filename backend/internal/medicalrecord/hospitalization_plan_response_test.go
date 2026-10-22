package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToHospitalizationPlanResponse(t *testing.T) {
	now := time.Now()
	price := int64(5000)
	bodySize := model.BodySizeMedium
	billingUnit := model.BillingUnitPerDay
	sortOrder := 3

	tests := []struct {
		name string
		in   *model.HospitalizationPlan
	}{
		{
			name: "converts full plan with all pointer fields set",
			in: &model.HospitalizationPlan{
				ID:          1,
				ClinicID:    2,
				Name:        "スタンダード",
				Price:       &price,
				IsActive:    true,
				Description: "標準入院プラン",
				BodySize:    &bodySize,
				BillingUnit: &billingUnit,
				TaxType:     model.TaxType("included"),
				TaxRate:     mustNumeric(t, "0.10"),
				SortOrder:   &sortOrder,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name: "converts zero-value plan with nil pointer fields",
			in:   &model.HospitalizationPlan{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toHospitalizationPlanResponse(tt.in)

			assert.Equal(t, tt.in.ID, got.ID)
			assert.Equal(t, tt.in.ClinicID, got.ClinicID)
			assert.Equal(t, tt.in.Name, got.Name)
			assert.Equal(t, tt.in.Price, got.Price)
			assert.Equal(t, tt.in.IsActive, got.IsActive)
			assert.Equal(t, tt.in.Description, got.Description)
			assert.Equal(t, string(tt.in.TaxType), got.TaxType)
			assert.Equal(t, tt.in.TaxRate, got.TaxRate)
			if tt.in.SortOrder == nil {
				assert.Equal(t, 0, got.SortOrder)
			} else {
				assert.Equal(t, *tt.in.SortOrder, got.SortOrder)
			}

			if tt.in.BodySize == nil {
				assert.Nil(t, got.BodySize)
			} else {
				assert.NotNil(t, got.BodySize)
				assert.Equal(t, string(*tt.in.BodySize), *got.BodySize)
			}

			if tt.in.BillingUnit == nil {
				assert.Nil(t, got.BillingUnit)
			} else {
				assert.NotNil(t, got.BillingUnit)
				assert.Equal(t, string(*tt.in.BillingUnit), *got.BillingUnit)
			}
		})
	}
}

// EMR-528: model の ExactNumeric / *int を写した response が既存 wire（tax_rate number, sort_order integer）を保つ。
func TestToHospitalizationPlanResponse_WireKeepsNumberTaxRateAndIntegerSortOrder(t *testing.T) {
	sortOrder := 0
	plan := &model.HospitalizationPlan{
		ID:        1,
		Name:      "0円プラン",
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   mustNumeric(t, "0"),
		SortOrder: &sortOrder,
	}

	raw, err := json.Marshal(toHospitalizationPlanResponse(plan))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))
	assert.InDelta(t, 0.0, got["tax_rate"], 0)
	assert.InDelta(t, 0.0, got["sort_order"], 0)
	_, hasBodySize := got["body_size"]
	assert.False(t, hasBodySize, "DB NULL の body_size は既存 omitempty の意味を保つ")
}
