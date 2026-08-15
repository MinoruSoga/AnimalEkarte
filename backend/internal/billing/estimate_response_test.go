package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToEstimateItemResponse(t *testing.T) {
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		item          *model.EstimateItem
		wantSubtotal  int64
		wantTaxAmount int64
	}{
		{
			name: "excluded tax computes subtotal and tax amount",
			item: &model.EstimateItem{
				ID:                    1,
				EstimateID:            10,
				Name:                  "血液検査",
				Category:              model.ItemCategoryExamination,
				UnitPrice:             1000,
				Quantity:              2,
				TaxType:               model.TaxTypeExcluded,
				TaxRate:               0.10,
				DiscountRate:          0,
				DiscountAmount:        0,
				IsInsuranceApplicable: true,
				SortOrder:             1,
				CreatedAt:             createdAt,
			},
			wantSubtotal:  2000,
			wantTaxAmount: 200,
		},
		{
			name: "included tax computes tax amount from gross subtotal",
			item: &model.EstimateItem{
				ID:         2,
				EstimateID: 10,
				Name:       "診察料",
				UnitPrice:  1100,
				Quantity:   1,
				TaxType:    model.TaxTypeIncluded,
				TaxRate:    0.10,
				SortOrder:  2,
				CreatedAt:  createdAt,
			},
			wantSubtotal:  1100,
			wantTaxAmount: 100,
		},
		{
			name: "discount reduces item subtotal and tax amount (MDL-01)",
			item: &model.EstimateItem{
				ID:             3,
				EstimateID:     10,
				Name:           "処置",
				UnitPrice:      1000,
				Quantity:       2,
				TaxType:        model.TaxTypeExcluded,
				TaxRate:        0.10,
				DiscountAmount: 500,
				SortOrder:      3,
				CreatedAt:      createdAt,
			},
			wantSubtotal:  1500,
			wantTaxAmount: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toEstimateItemResponse(tt.item)

			assert.Equal(t, tt.item.ID, got.ID)
			assert.Equal(t, tt.item.EstimateID, got.EstimateID)
			assert.Equal(t, tt.item.Name, got.Name)
			assert.Equal(t, string(tt.item.Category), got.Category)
			assert.Equal(t, tt.item.UnitPrice, got.UnitPrice)
			assert.Equal(t, tt.item.Quantity, got.Quantity)
			assert.Equal(t, tt.wantSubtotal, got.Subtotal)
			assert.Equal(t, string(tt.item.TaxType), got.TaxType)
			assert.Equal(t, tt.item.TaxRate, got.TaxRate)
			assert.Equal(t, tt.wantTaxAmount, got.TaxAmount)
			assert.Equal(t, tt.item.DiscountRate, got.DiscountRate)
			assert.Equal(t, tt.item.DiscountAmount, got.DiscountAmount)
			assert.Equal(t, tt.item.IsInsuranceApplicable, got.IsInsuranceApplicable)
			assert.Equal(t, tt.item.SortOrder, got.SortOrder)
		})
	}
}

func TestToEstimateResponse(t *testing.T) {
	validUntil := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	createdBy := uint64(3)
	ownerID := uint64(5)

	t.Run("maps all fields with items and owner", func(t *testing.T) {
		estimate := &model.Estimate{
			ID:              1,
			ClinicID:        1,
			EstimateNo:      "EST-0001",
			MedicalRecordID: nil,
			Title:           "見積書",
			OwnerID:         &ownerID,
			Owner:           &model.Owner{ID: ownerID, Name: "田中太郎"},
			Status:          model.EstimateStatusSent,
			Subtotal:        1000,
			TaxTotal:        100,
			TotalAmount:     1100,
			InsuranceAmount: 0,
			DiscountAmount:  0,
			ValidUntil:      &validUntil,
			Comment:         "comment",
			Notes:           "notes",
			CreatedBy:       &createdBy,
			Items: []model.EstimateItem{
				{ID: 1, EstimateID: 1, Name: "血液検査", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.1},
			},
		}

		got := toEstimateResponse(estimate)

		assert.Equal(t, estimate.ID, got.ID)
		assert.Equal(t, estimate.ClinicID, got.ClinicID)
		assert.Equal(t, estimate.EstimateNo, got.EstimateNo)
		assert.Equal(t, estimate.Title, got.Title)
		require.NotNil(t, got.Owner)
		assert.Equal(t, "田中太郎", got.Owner.OwnerName)
		assert.Equal(t, string(estimate.Status), got.Status)
		assert.Equal(t, estimate.Subtotal, got.Subtotal)
		assert.Equal(t, estimate.TaxTotal, got.TaxTotal)
		assert.Equal(t, estimate.TotalAmount, got.TotalAmount)
		assert.Equal(t, estimate.Comment, got.Comment)
		assert.Equal(t, estimate.Notes, got.Notes)
		require.NotNil(t, got.CreatedBy)
		assert.Equal(t, createdBy, *got.CreatedBy)
		require.Len(t, got.Items, 1)
		assert.Equal(t, "血液検査", got.Items[0].Name)
	})

	t.Run("nil owner and empty items produce nil owner and empty item slice", func(t *testing.T) {
		estimate := &model.Estimate{ID: 2, Title: "見積書2"}

		got := toEstimateResponse(estimate)

		assert.Nil(t, got.Owner)
		assert.NotNil(t, got.Items)
		assert.Empty(t, got.Items)
	})
}
