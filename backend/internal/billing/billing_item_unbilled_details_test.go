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
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

func TestBillingItemService_CreateItem_AllowsWhenUnbilledWarning(t *testing.T) {
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
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.True(t, createCalled, "明細追加は未請求予防接種 warning で止めない（確定時のみ拒否）")
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
	c.Request = httptest.NewRequest(http.MethodGet, "/billing-items/unbilled-details?pet_id=7", http.NoBody)
	setClinicID(c)
	h.GetUnbilledItemDetails(c)
	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	_, hasItems := body["items"]
	_, hasWarnings := body["warnings"]
	_, hasRevision := body["revision"]
	assert.True(t, hasItems)
	assert.True(t, hasWarnings)
	// EMR-196②: complete の expected_unbilled_revision へ返送する集約版を公開する
	assert.True(t, hasRevision)
	// no extra top-level keys
	assert.Len(t, body, 3)
	warnings, ok := body["warnings"].([]any)
	require.True(t, ok)
	require.Len(t, warnings, 1)
	w0 := warnings[0].(map[string]any)
	assert.Equal(t, []string{"blocking", "code", "count", "source"}, sortedKeys(w0))
}

// EMR-65 / BUG-BILLING-TAX-TYPE-DROPPED: リンク済みマスタの税区分・税率が
// 未請求候補明細へ伝播する。内税・非課税・非既定税率を外税10%へ一律潰しする
// 回帰（過課金）を防ぐ。税フィールドを持たない inventory / 未リンク行は
// 従来の外税既定を維持し、候補行のスキップや0円フォールバックは禁止。
func TestBillingItemService_UnbilledDetails_MasterTaxPropagation(t *testing.T) {
	repo := &matrixVaccinationRepo{mockBillingItemRepository: defaultMockBillingItemRepo()}
	svc := newMatrixService(t, repo, &matrixTreatmentRepo{
		items: []model.Treatment{
			{
				ID: 11, ItemType: model.TreatmentItemTypeProcedure, Content: "内税処置",
				UnitPrice: 1000, Quantity: 1,
				Procedure: &model.Procedure{TaxType: model.TaxTypeIncluded, TaxRate: 0.08},
			},
			{
				ID: 12, ItemType: model.TreatmentItemTypeMedicine, Content: "非課税薬",
				UnitPrice: 500, Quantity: 2,
				Medicine: &model.Medicine{TaxType: model.TaxTypeExempt, TaxRate: 0},
			},
			{
				ID: 13, ItemType: model.TreatmentItemTypeConsultation, Content: "外税診察",
				UnitPrice: 3000, Quantity: 1,
				Consultation: &model.Consultation{TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
			},
			{
				ID: 14, ItemType: model.TreatmentItemTypeOther, Content: "物販",
				UnitPrice: 800, Quantity: 1,
				Inventory: &model.InventoryItem{},
			},
			{
				ID: 15, ItemType: model.TreatmentItemTypeOther, Content: "マスタ未リンク",
				UnitPrice: 200, Quantity: 1,
			},
		},
	})

	details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
	require.NoError(t, err)
	require.Len(t, details.Items, 5, "税メタデータの有無にかかわらず候補行は全件返す")

	byID := make(map[uint64]model.BillingItem, len(details.Items))
	for _, item := range details.Items {
		byID[item.ID] = item
		assert.Greater(t, item.UnitPrice, int64(0), "zero-yen fallback は禁止")
	}

	assert.Equal(t, model.TaxTypeIncluded, byID[11].TaxType, "procedure マスタの内税を伝播")
	assert.InDelta(t, 0.08, byID[11].TaxRate, 1e-9)
	assert.Equal(t, model.TaxTypeExempt, byID[12].TaxType, "medicine マスタの非課税を伝播")
	assert.InDelta(t, 0, byID[12].TaxRate, 1e-9)
	assert.Equal(t, model.TaxTypeExcluded, byID[13].TaxType, "consultation マスタの外税を伝播")
	assert.InDelta(t, 0.10, byID[13].TaxRate, 1e-9)
	assert.Equal(t, model.TaxTypeExcluded, byID[14].TaxType, "inventory は税フィールドを持たないため既定維持")
	assert.InDelta(t, sharedkernel.DefaultTaxRate, byID[14].TaxRate, 1e-9)
	assert.Equal(t, model.TaxTypeExcluded, byID[15].TaxType, "未リンク行も既定維持")
	assert.InDelta(t, sharedkernel.DefaultTaxRate, byID[15].TaxRate, 1e-9)
}

// EMR-196②: AssertUnbilledForComplete は complete 確定時の版照合。
// blocking warning（BUG-013）を先に検査し、その後 expected revision と
// tx 内再集計の版を比較する（不一致・空指定は Conflict）。
func TestBillingItemService_AssertUnbilledForComplete(t *testing.T) {
	buildService := func() BillingItemService {
		repo := &matrixVaccinationRepo{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			items: []model.BillingItem{{
				ID: 22, Name: "混合ワクチン", UnitPrice: 5000, Quantity: 1,
				Category: model.ItemCategoryVaccine,
			}},
		}
		return newMatrixService(t, repo, &matrixTreatmentRepo{
			items: []model.Treatment{{ID: 11, Content: "処置A", UnitPrice: 15000, Quantity: 1}},
		})
	}

	currentRevision := func(t *testing.T, svc BillingItemService) string {
		t.Helper()
		details, err := svc.GetUnbilledItemDetails(context.Background(), 1, 7)
		require.NoError(t, err)
		require.NotEmpty(t, details.Revision, "集約版が空文字列では token にならない")
		return details.Revision
	}

	t.Run("現在の集約版と一致すれば確定を許可する", func(t *testing.T) {
		svc := buildService()
		expected := currentRevision(t, svc)
		err := svc.AssertUnbilledForComplete(context.Background(), 1, 7, expected)
		require.NoError(t, err)
	})

	t.Run("別端末の追記相当（items が増えた集約版）とは不一致で Conflict", func(t *testing.T) {
		svc := buildService()
		err := svc.AssertUnbilledForComplete(context.Background(), 1, 7, "u1:stale")
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "want 409 conflict, got %v", err)

		var conflict *unbilledRevisionConflictError
		require.True(t, errors.As(err, &conflict), "want unbilledRevisionConflictError, got %T", err)
		assert.Equal(t, currentRevision(t, svc), conflict.CurrentRevision)

		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, AccountingCodeUnbilledConflict, appErr.Code)
	})

	t.Run("空の expected は必ず不一致（未指定のすり抜け防止）", func(t *testing.T) {
		svc := buildService()
		err := svc.AssertUnbilledForComplete(context.Background(), 1, 7, "")
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("blocking warning は版照合より先に BUG-013 の Conflict を返す", func(t *testing.T) {
		repo := &matrixVaccinationRepo{
			mockBillingItemRepository: defaultMockBillingItemRepo(),
			unbillable:                1,
		}
		svc := newMatrixService(t, repo, &matrixTreatmentRepo{})
		err := svc.AssertUnbilledForComplete(context.Background(), 1, 7, "u1:any")
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "請求不能な予防接種")
		var conflict *unbilledRevisionConflictError
		assert.False(t, errors.As(err, &conflict), "blocking は UNBILLED_ITEMS_CHANGED ではなく BUG-013 の既存契約")
	})
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
