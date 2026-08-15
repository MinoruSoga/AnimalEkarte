package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- BUG-013 source matrix: unbilled-details aggregation ----

type matrixTreatmentRepo struct {
	items []model.Treatment
	err   error
}

func (m *matrixTreatmentRepo) FindUnbilledByPetID(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *matrixTreatmentRepo) CountFinalizedUnconfirmedByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

type matrixVaccinationRepo struct {
	*mockBillingItemRepository
	items       []model.BillingItem
	unbillable  int
	err         error
	trimming    []model.BillingItem
	trimmingErr error
}

func (m *matrixVaccinationRepo) FindUnbilledVaccinationItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.items, m.unbillable, nil
}

func (m *matrixVaccinationRepo) FindUnbilledTrimmingItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	if m.trimmingErr != nil {
		return nil, m.trimmingErr
	}
	return m.trimming, nil
}

func newMatrixService(t *testing.T, repo *matrixVaccinationRepo, treatments *matrixTreatmentRepo) BillingItemService {
	t.Helper()
	billingRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			petID := uint64(7)
			return &model.Billing{ID: id, ClinicID: clinicID, PetID: &petID, Status: model.BillingStatusWaiting}, nil
		},
	}
	return NewBillingItemServiceWithCampaign(
		repo,
		billingRepo,
		treatments,
		&mockTransactor{},
		nil, nil, nil, nil,
	)
}

func TestBillingItemService_UnbilledDetails_SourceMatrix(t *testing.T) {
	t.Run("all sources succeed returns items without warnings", func(t *testing.T) {
		txID := uint64(11)
		vaccID := uint64(22)
		repo := &matrixVaccinationRepo{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			items: []model.BillingItem{{
				ID: vaccID, Name: "混合ワクチン", UnitPrice: 5000, Quantity: 1,
				Category: model.ItemCategoryVaccine, VaccinationID: &vaccID,
			}},
			trimming: []model.BillingItem{{
				ID: 33, Name: "爪切り", UnitPrice: 300, Quantity: 1,
				Category: model.ItemCategoryTrimming,
			}},
		}
		svc := newMatrixService(t, repo, &matrixTreatmentRepo{
			items: []model.Treatment{{ID: txID, Content: "処置A", UnitPrice: 15000, Quantity: 1}},
		})

		details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
		require.NoError(t, err)
		require.NotNil(t, details)
		assert.Len(t, details.Items, 3)
		assert.Empty(t, details.Warnings)

		// legacy all-success invariant
		legacy, err := svc.GetUnbilledItems(context.Background(), 1, 7)
		require.NoError(t, err)
		assert.Len(t, legacy, 3)
	})

	t.Run("vaccination unbillable yields items plus blocking warning not 500", func(t *testing.T) {
		txID := uint64(11)
		repo := &matrixVaccinationRepo{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			items:                     nil,
			unbillable:                2,
			trimming: []model.BillingItem{{
				ID: 33, Name: "爪切り", UnitPrice: 300, Quantity: 1,
				Category: model.ItemCategoryTrimming,
			}},
		}
		svc := newMatrixService(t, repo, &matrixTreatmentRepo{
			items: []model.Treatment{{ID: txID, Content: "処置A", UnitPrice: 15000, Quantity: 1}},
		})

		details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
		require.NoError(t, err)
		require.NotNil(t, details)
		assert.Len(t, details.Items, 2, "treatment + trimming kept")
		require.Len(t, details.Warnings, 1)
		w := details.Warnings[0]
		assert.Equal(t, UnbilledWarningSourceVaccination, w.Source)
		assert.Equal(t, UnbilledWarningCodeVaccinationMasterUnbillable, w.Code)
		assert.Equal(t, 2, w.Count)
		assert.True(t, w.Blocking)
		// payload must only expose source/code/count/blocking (struct fields)
		raw, err := json.Marshal(w)
		require.NoError(t, err)
		var asMap map[string]any
		require.NoError(t, json.Unmarshal(raw, &asMap))
		assert.Equal(t, map[string]any{
			"source":   UnbilledWarningSourceVaccination,
			"code":     UnbilledWarningCodeVaccinationMasterUnbillable,
			"count":    float64(2),
			"blocking": true,
		}, asMap)

		// legacy remains fail-closed (no silent partial array)
		legacy, err := svc.GetUnbilledItems(context.Background(), 1, 7)
		assert.Error(t, err)
		assert.Nil(t, legacy)
	})

	t.Run("infra error on vaccination stays 500-class fail-closed", func(t *testing.T) {
		repo := &matrixVaccinationRepo{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			err:                       errors.New("sql: connection refused"),
		}
		svc := newMatrixService(t, repo, &matrixTreatmentRepo{
			items: []model.Treatment{{ID: 1, Content: "処置", UnitPrice: 1000, Quantity: 1}},
		})

		details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
		assert.Error(t, err)
		assert.Nil(t, details)
		assert.Contains(t, err.Error(), "failed to find unbilled vaccination items")
	})

	t.Run("all sources empty returns empty items and no warnings", func(t *testing.T) {
		repo := &matrixVaccinationRepo{mockBillingItemRepository: defaultMockBillingItemRepo()}
		svc := newMatrixService(t, repo, &matrixTreatmentRepo{})
		details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
		require.NoError(t, err)
		assert.Empty(t, details.Items)
		assert.Empty(t, details.Warnings)
	})
}

func TestBillingItemService_CreateItem_BlocksWhenUnbilledWarning(t *testing.T) {
	createCalled := false
	repo := &matrixVaccinationRepo{
		mockBillingItemRepository: &mockBillingItemRepository{
			createFn: func(_ context.Context, _ *model.BillingItem) error {
				createCalled = true
				return nil
			},
			findByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
				return nil, nil
			},
			validateCreateReferencesFn: func(_ context.Context, _, _ uint64, _, _, _, _, _ *uint64) (model.ItemCategory, error) {
				return model.ItemCategoryProcedure, nil
			},
		},
		unbillable: 1,
	}
	billingRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
			petID := uint64(7)
			return &model.Billing{ID: id, ClinicID: clinicID, PetID: &petID, Status: model.BillingStatusWaiting}, nil
		},
	}
	svc := NewBillingItemServiceWithCampaign(
		repo, billingRepo, &matrixTreatmentRepo{}, &mockTransactor{}, nil, nil, nil, nil,
	)

	item, err := svc.CreateItem(context.Background(), &CreateBillingItemInput{
		ClinicID:  1,
		BillingID: 10,
		Category:  string(model.ItemCategoryProcedure),
		Name:      "有効な処置のみ",
		UnitPrice: 15000,
		Quantity:  1,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected conflict: %v", err)
	assert.Nil(t, item)
	assert.False(t, createCalled, "must not partially commit billing_item")
}

func TestAccountingService_Create_BlocksWhenUnbilledWarning(t *testing.T) {
	createCalled := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	guard := &mockBillingItemService{
		assertNoBlockingUnbilledFn: func(_ context.Context, _, _ uint64) error {
			return apperrors.WrapConflict("未請求候補に請求不能な予防接種が含まれるため会計を確定できません")
		},
	}
	svc := NewAccountingService(
		repo, nil, nil, nil, nil, &mockTransactor{}, nil, &mockPaymentMethodMasterRepository{},
		WithUnbilledWriteGuard(guard),
	)
	petID := uint64(7)
	created, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		PetID:         &petID,
		Subtotal:      1000,
		TaxTotal:      100,
		TotalAmount:   1100,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, created)
	assert.False(t, createCalled)
}

func TestGetUnbilledItemDetails_HandlerEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockBillingItemService{
		getUnbilledItemDetailsFn: func(_ context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), petID)
			txID := uint64(11)
			return &UnbilledDetails{
				Items: []model.BillingItem{{
					ID: txID, Name: "処置", UnitPrice: 1000, Quantity: 1,
					Category: model.ItemCategoryProcedure, TreatmentID: &txID,
				}},
				Warnings: []UnbilledWarning{{
					Source: UnbilledWarningSourceVaccination,
					Code:   UnbilledWarningCodeVaccinationMasterUnbillable,
					Count:  1, Blocking: true,
				}},
			}, nil
		},
	}
	h := newHandlerWithBillingItemSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/billing-items/unbilled-details?pet_id=7", nil)
	setClinicID(c)
	h.GetUnbilledItemDetails(c)
	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	_, hasItems := body["items"]
	_, hasWarnings := body["warnings"]
	assert.True(t, hasItems)
	assert.True(t, hasWarnings)
	// no extra top-level keys
	assert.Len(t, body, 2)
	warnings, ok := body["warnings"].([]any)
	require.True(t, ok)
	require.Len(t, warnings, 1)
	w0 := warnings[0].(map[string]any)
	assert.Equal(t, []string{"blocking", "code", "count", "source"}, sortedKeys(w0))
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// simple insertion sort for stable assert without importing sort for 4 keys
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}
	return keys
}
