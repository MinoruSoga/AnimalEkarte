package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOutstandingAmount(t *testing.T) {
	t.Parallel()

	t.Run("waiting without payment: full total_amount", func(t *testing.T) {
		t.Parallel()
		b := &model.Billing{Status: model.BillingStatusWaiting, TotalAmount: 1100}
		assert.Equal(t, int64(1100), OutstandingAmount(b))
	})

	t.Run("completed full card settle: residual 0", func(t *testing.T) {
		t.Parallel()
		b := &model.Billing{
			Status:      model.BillingStatusCompleted,
			TotalAmount: 1100,
			Payments: []model.Payment{{
				TotalAmount: 1100, BillingAmount: 1100,
			}},
		}
		assert.Equal(t, int64(0), OutstandingAmount(b))
	})

	t.Run("BUG-007: credit correction underpay residual", func(t *testing.T) {
		t.Parallel()
		// medical 1100, insurance/discount 0, card corrected to 900 → unpaid 200
		b := &model.Billing{
			Status:      model.BillingStatusCompleted,
			TotalAmount: 1100,
			Payments: []model.Payment{{
				TotalAmount: 1100, BillingAmount: 900,
			}},
		}
		assert.Equal(t, int64(200), OutstandingAmount(b))
	})

	t.Run("insurance: residual uses patient_due not medical total", func(t *testing.T) {
		t.Parallel()
		// medical 10000, insurance 5000 → due 5000; collected 4000 → unpaid 1000
		b := &model.Billing{
			Status:      model.BillingStatusCompleted,
			TotalAmount: 10000,
			Payments: []model.Payment{{
				TotalAmount: 10000, InsuranceAmount: 5000, BillingAmount: 4000,
			}},
		}
		assert.Equal(t, int64(1000), OutstandingAmount(b))
	})

	t.Run("over-collection residual is 0 not negative", func(t *testing.T) {
		t.Parallel()
		b := &model.Billing{
			Status:      model.BillingStatusCompleted,
			TotalAmount: 10000,
			Payments: []model.Payment{{
				TotalAmount: 10000, BillingAmount: 12000,
			}},
		}
		assert.Equal(t, int64(0), OutstandingAmount(b))
	})

	t.Run("cancelled is 0", func(t *testing.T) {
		t.Parallel()
		b := &model.Billing{
			Status:      model.BillingStatusCancelled,
			TotalAmount: 1100,
			Payments:    []model.Payment{{TotalAmount: 1100, BillingAmount: 0}},
		}
		assert.Equal(t, int64(0), OutstandingAmount(b))
	})

	t.Run("nil billing is 0", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, int64(0), OutstandingAmount(nil))
	})
}
