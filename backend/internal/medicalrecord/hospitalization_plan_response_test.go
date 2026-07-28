package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToHospitalizationPlanResponse(t *testing.T) {
	now := time.Now()
	price := int64(5000)
	bodySize := model.BodySizeMedium
	billingUnit := model.BillingUnitPerDay

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
				TaxRate:     0.1,
				SortOrder:   3,
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
			assert.Equal(t, tt.in.SortOrder, got.SortOrder)

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
