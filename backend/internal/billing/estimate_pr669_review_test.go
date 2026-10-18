package billing

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- PR669: 明細合計への保険適用額控除と legacy 保存済み合計の保持 ----

// newEstimateCapturingRepo は create/update 両方で cmd/estimate を捕捉する共有モック。
func newEstimateCapturingRepo(existing *model.Estimate, saved **model.Estimate, updatedCmd *UpdateEstimateInput) *mockEstimateRepository {
	return &mockEstimateRepository{
		createFn: func(_ context.Context, e *model.Estimate) error {
			e.ID = 12
			cp := *e
			*saved = &cp
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			if *saved != nil {
				cp := **saved
				cp.ID = id
				return &cp, nil
			}
			if existing != nil {
				cp := *existing
				return &cp, nil
			}
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		lockEditableByIDFn: func(_ context.Context, _, id uint64) (*model.Estimate, error) {
			if existing != nil {
				cp := *existing
				return &cp, nil
			}
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
			*updatedCmd = cmd
			if existing != nil {
				cp := *existing
				return &cp, nil
			}
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		replaceItemsFn: func(_ context.Context, _, _ uint64, _ []model.EstimateItem) error {
			return nil
		},
	}
}

func TestEstimateService_PR669_Create_ItemizedDeductsDiscountAndInsurance(t *testing.T) {
	tests := []struct {
		name         string
		item         EstimateItemInput
		discount     int64
		insurance    int64
		wantSubtotal int64
		wantTaxTotal int64
		wantTotal    int64
	}{
		{
			name:      "excluded tax gross1100 discount100 insurance200 yields800",
			item:      EstimateItemInput{Name: "外税", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: ptrFloat64(0.10)},
			discount:  100,
			insurance: 200,
			// subtotal1000/tax100 gross1100 → 1100-100-200=800
			wantSubtotal: 1000,
			wantTaxTotal: 100,
			wantTotal:    800,
		},
		{
			name:      "included tax gross1100 discount100 insurance200 yields800",
			item:      EstimateItemInput{Name: "内税", UnitPrice: 1100, Quantity: 1, TaxType: model.TaxTypeIncluded, TaxRate: ptrFloat64(0.10)},
			discount:  100,
			insurance: 200,
			// gross1100 (subtotal1100 内税100) → 800
			wantSubtotal: 1100,
			wantTaxTotal: 100,
			wantTotal:    800,
		},
		{
			name:         "insurance equal to gross minus discount yields zero",
			item:         EstimateItemInput{Name: "外税", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: ptrFloat64(0.10)},
			discount:     100,
			insurance:    1000,
			wantSubtotal: 1000,
			wantTaxTotal: 100,
			wantTotal:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var saved *model.Estimate
			var updatedCmd UpdateEstimateInput
			repo := newEstimateCapturingRepo(nil, &saved, &updatedCmd)
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

			_, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
				Title:           "調整付き明細",
				DiscountAmount:  tt.discount,
				InsuranceAmount: tt.insurance,
				CreatedBy:       ptrU64(estimateTestCreatedByStaffID),
				Items:           []EstimateItemInput{tt.item},
			})

			require.NoError(t, err)
			require.NotNil(t, saved)
			assert.Equal(t, tt.wantSubtotal, saved.Subtotal)
			assert.Equal(t, tt.wantTaxTotal, saved.TaxTotal)
			assert.Equal(t, tt.wantTotal, saved.TotalAmount)
		})
	}
}

func TestEstimateService_PR669_Create_RejectsExcessAdjustments(t *testing.T) {
	item := EstimateItemInput{Name: "外税", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: ptrFloat64(0.10)}
	tests := []struct {
		name      string
		discount  int64
		insurance int64
	}{
		{name: "insurance above gross", discount: 0, insurance: 1101},
		{name: "combined discount plus insurance above gross", discount: 200, insurance: 901},
		{name: "negative insurance", discount: 0, insurance: -1},
		{name: "negative discount", discount: -1, insurance: 0},
		{name: "discount above gross", discount: 1101, insurance: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeCalled := false
			repo := &mockEstimateRepository{
				createFn:       func(_ context.Context, _ *model.Estimate) error { writeCalled = true; return nil },
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ []model.EstimateItem) error { writeCalled = true; return nil },
			}
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})

			got, err := svc.Create(context.Background(), 1, &CreateEstimateInput{
				Title:           "超過控除",
				DiscountAmount:  tt.discount,
				InsuranceAmount: tt.insurance,
				CreatedBy:       ptrU64(estimateTestCreatedByStaffID),
				Items:           []EstimateItemInput{item},
			})

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
			assert.Nil(t, got)
			assert.False(t, writeCalled, "no write may happen on rejected adjustments")
		})
	}
}

func TestEstimateService_PR669_Update_InsuranceAndDiscount(t *testing.T) {
	existing := &model.Estimate{
		ID:              3,
		Status:          model.EstimateStatusDraft,
		Subtotal:        1000,
		TaxTotal:        100,
		TotalAmount:     1100,
		DiscountAmount:  100,
		InsuranceAmount: 0,
		Items: []model.EstimateItem{
			{ID: 7, EstimateID: 3, Name: "外税処置", UnitPrice: 1000, Quantity: 1,
				TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
		},
	}

	t.Run("insurance-only PATCH with items omitted recomputes total", func(t *testing.T) {
		var saved *model.Estimate
		var updatedCmd UpdateEstimateInput
		repo := newEstimateCapturingRepo(existing, &saved, &updatedCmd)
		svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		insurance := int64(200)
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{InsuranceAmount: &insurance})
		require.NoError(t, err)
		require.NotNil(t, updatedCmd.TotalAmount)
		// gross1100 - stored discount100 - insurance200 = 800
		assert.Equal(t, int64(800), *updatedCmd.TotalAmount)
		require.NotNil(t, updatedCmd.Subtotal)
		assert.Equal(t, int64(1000), *updatedCmd.Subtotal)
	})

	t.Run("items replacement retains omitted stored insurance", func(t *testing.T) {
		withInsurance := *existing
		withInsurance.InsuranceAmount = 200
		var saved *model.Estimate
		var updatedCmd UpdateEstimateInput
		repo := newEstimateCapturingRepo(&withInsurance, &saved, &updatedCmd)
		svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		items := []EstimateItemInput{{Name: "外税処置", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: ptrFloat64(0.10)}}
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Items: &items})
		require.NoError(t, err)
		require.NotNil(t, updatedCmd.TotalAmount)
		// gross1100 - stored discount100 - stored insurance200 = 800
		assert.Equal(t, int64(800), *updatedCmd.TotalAmount)
	})

	t.Run("insurance equal to gross minus discount allowed as zero", func(t *testing.T) {
		var saved *model.Estimate
		var updatedCmd UpdateEstimateInput
		repo := newEstimateCapturingRepo(existing, &saved, &updatedCmd)
		svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		insurance := int64(1000)
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{InsuranceAmount: &insurance})
		require.NoError(t, err)
		require.NotNil(t, updatedCmd.TotalAmount)
		assert.Equal(t, int64(0), *updatedCmd.TotalAmount)
	})

	t.Run("insurance above gross minus discount rejected before write", func(t *testing.T) {
		writeCalled := false
		repo := newEstimateCapturingRepo(existing, new(*model.Estimate), new(UpdateEstimateInput))
		repo.updateIfNotLockedFn = func(_ context.Context, _, id uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
			writeCalled = true
			return &model.Estimate{ID: id}, nil
		}
		svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		insurance := int64(1001)
		got, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{InsuranceAmount: &insurance})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
		assert.Nil(t, got)
		assert.False(t, writeCalled)
	})
}

func TestEstimateService_PR669_Update_PreservesLegacyStoredTotals(t *testing.T) {
	// 保存時点より前の旧データ: stored 割引が明細総額を超過していても、
	// 控除据え置きのメタデータ PATCH は保存済み合計をそのまま保持し再検証しない。
	existing := &model.Estimate{
		ID:              3,
		Status:          model.EstimateStatusDraft,
		Subtotal:        1000,
		TaxTotal:        100,
		TotalAmount:     1100,
		DiscountAmount:  2000, // stored excess deduction (legacy)
		InsuranceAmount: 0,
		Items: []model.EstimateItem{
			{ID: 7, EstimateID: 3, Name: "外税処置", UnitPrice: 1000, Quantity: 1,
				TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
		},
	}
	newTitle := "タイトル変更"
	newStatus := model.EstimateStatusSent

	newRepo := func(replaceCalled *bool, updatedCmd *UpdateEstimateInput) *mockEstimateRepository {
		return &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
				cp := *existing
				return &cp, nil
			},
			lockEditableByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
				cp := *existing
				return &cp, nil
			},
			updateIfNotLockedFn: func(_ context.Context, _, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
				*updatedCmd = cmd
				cp := *existing
				return &cp, nil
			},
			replaceItemsFn: func(_ context.Context, _, _ uint64, _ []model.EstimateItem) error {
				*replaceCalled = true
				return nil
			},
		}
	}

	t.Run("metadata title PATCH preserves stored total and excess discount without error", func(t *testing.T) {
		var updatedCmd UpdateEstimateInput
		replaceCalled := false
		svc := NewEstimateService(newRepo(&replaceCalled, &updatedCmd), nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Title: &newTitle})
		require.NoError(t, err)
		assert.Nil(t, updatedCmd.Subtotal, "metadata-only PATCH must not touch totals")
		assert.Nil(t, updatedCmd.TaxTotal)
		assert.Nil(t, updatedCmd.TotalAmount)
		assert.False(t, replaceCalled, "no ReplaceItems on header-only PATCH")
	})

	t.Run("status-only PATCH preserves stored totals", func(t *testing.T) {
		var updatedCmd UpdateEstimateInput
		replaceCalled := false
		svc := NewEstimateService(newRepo(&replaceCalled, &updatedCmd), nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Status: &newStatus})
		require.NoError(t, err)
		assert.Nil(t, updatedCmd.TotalAmount)
		assert.False(t, replaceCalled)
	})

	t.Run("same-valued discount and insurance fields preserve stored totals", func(t *testing.T) {
		var updatedCmd UpdateEstimateInput
		replaceCalled := false
		svc := NewEstimateService(newRepo(&replaceCalled, &updatedCmd), nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		sameDiscount := int64(2000)
		sameInsurance := int64(0)
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{
			DiscountAmount:  &sameDiscount,
			InsuranceAmount: &sameInsurance,
		})
		require.NoError(t, err)
		assert.Nil(t, updatedCmd.TotalAmount, "same-valued deduction fields must not recompute")
		assert.False(t, replaceCalled)
	})

	t.Run("tampered client totals cannot drift persisted amounts", func(t *testing.T) {
		var updatedCmd UpdateEstimateInput
		replaceCalled := false
		svc := NewEstimateService(newRepo(&replaceCalled, &updatedCmd), nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		tampered := int64(1)
		_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{
			Title:       &newTitle,
			Subtotal:    &tampered,
			TaxTotal:    &tampered,
			TotalAmount: &tampered,
		})
		require.NoError(t, err)
		require.NotNil(t, updatedCmd.Subtotal)
		require.NotNil(t, updatedCmd.TaxTotal)
		require.NotNil(t, updatedCmd.TotalAmount)
		assert.Equal(t, int64(1000), *updatedCmd.Subtotal)
		assert.Equal(t, int64(100), *updatedCmd.TaxTotal)
		assert.Equal(t, int64(1100), *updatedCmd.TotalAmount)
	})
}

func TestEstimateService_PR669_Update_ChangedDeductionStillValidates(t *testing.T) {
	existing := &model.Estimate{
		ID:              3,
		Status:          model.EstimateStatusDraft,
		DiscountAmount:  0,
		InsuranceAmount: 0,
		Items: []model.EstimateItem{
			{ID: 7, EstimateID: 3, Name: "外税処置", UnitPrice: 1000, Quantity: 1,
				TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
		},
	}
	t.Run("discount change above gross is invalid input", func(t *testing.T) {
		writeCalled := false
		repo := &mockEstimateRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
				cp := *existing
				return &cp, nil
			},
			updateIfNotLockedFn: func(_ context.Context, _, id uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
				writeCalled = true
				return &model.Estimate{ID: id}, nil
			},
		}
		svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
		discount := int64(2000)
		got, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{DiscountAmount: &discount})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
		assert.False(t, writeCalled)
	})
}

// ---- PR669: 数量 0.1 刻みの浮動小数点誤判定 ----

func TestEstimateService_PR669_QuantityStepValidation(t *testing.T) {
	validQuantities := []struct {
		name string
		q    float64
	}{
		{"0.1", 0.1},
		{"0.3", 0.3},
		{"1000000.1", 1000000.1},
		{"upper range decimal 999999999.9", 999999999.9},
		{"999999999.8", 999999999.8},
		{"500000000.5", 500000000.5},
	}
	for _, tt := range validQuantities {
		t.Run("valid "+tt.name, func(t *testing.T) {
			err := validateEstimateItemInputs([]EstimateItemInput{
				{Name: "x", UnitPrice: 1, Quantity: tt.q},
			})
			assert.NoError(t, err)
		})
	}

	invalidQuantities := []struct {
		name string
		q    float64
	}{
		{"0.15 not a 0.1 step", 0.15},
		{"1000000.15 not a 0.1 step", 1000000.15},
		{"1.01 not a 0.1 step", 1.01},
		{"0.05 not a 0.1 step", 0.05},
	}
	for _, tt := range invalidQuantities {
		t.Run("invalid "+tt.name, func(t *testing.T) {
			err := validateEstimateItemInputs([]EstimateItemInput{
				{Name: "x", UnitPrice: 1, Quantity: tt.q},
			})
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
		})
	}
}

func TestEstimateService_PR669_Update_ReplacementRejectsExcessAdjustments(t *testing.T) {
	// 明細置換時も控除の上限検証が適用され、拒否時は UpdateIfNotLocked / ReplaceItems
	// ともに呼ばれないことを確認する。
	existing := &model.Estimate{
		ID:              3,
		Status:          model.EstimateStatusDraft,
		DiscountAmount:  0,
		InsuranceAmount: 0,
		Items: []model.EstimateItem{
			{ID: 7, EstimateID: 3, Name: "外税処置", UnitPrice: 1000, Quantity: 1,
				TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
		},
	}
	tests := []struct {
		name      string
		discount  int64
		insurance int64
	}{
		{name: "negative insurance on replacement", discount: 0, insurance: -1},
		{name: "negative discount on replacement", discount: -1, insurance: 0},
		{name: "combined discount plus insurance exceeds gross", discount: 600, insurance: 501},
		{name: "insurance int64 max rejected without overflow", discount: 0, insurance: math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			replaceCalled := false
			repo := &mockEstimateRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					cp := *existing
					return &cp, nil
				},
				lockEditableByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					cp := *existing
					return &cp, nil
				},
				updateIfNotLockedFn: func(_ context.Context, _, id uint64, _ UpdateEstimateInput) (*model.Estimate, error) {
					updateCalled = true
					return &model.Estimate{ID: id}, nil
				},
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ []model.EstimateItem) error {
					replaceCalled = true
					return nil
				},
			}
			svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
			items := []EstimateItemInput{{Name: "外税処置", UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: ptrFloat64(0.10)}}
			got, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{
				Items:           &items,
				DiscountAmount:  &tt.discount,
				InsuranceAmount: &tt.insurance,
			})
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input, got %v", err)
			assert.Nil(t, got)
			assert.False(t, updateCalled, "no UpdateIfNotLocked on rejected adjustments")
			assert.False(t, replaceCalled, "no ReplaceItems on rejected adjustments")
		})
	}
}

func TestEstimateService_PR669_Update_ClearAllRowsWithStoredInsurance(t *testing.T) {
	// 保存済み保険適用額がある状態での全行クリアも合計 0（控除はメタデータとして保持）。
	existing := &model.Estimate{
		ID:              3,
		Status:          model.EstimateStatusDraft,
		DiscountAmount:  100,
		InsuranceAmount: 200,
		Items: []model.EstimateItem{
			{ID: 7, EstimateID: 3, Name: "外税処置", UnitPrice: 1000, Quantity: 1,
				TaxType: model.TaxTypeExcluded, TaxRate: 0.10},
		},
	}
	replaced := []model.EstimateItem(nil)
	var updatedCmd UpdateEstimateInput
	repo := &mockEstimateRepository{
		lockEditableByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			cp := *existing
			return &cp, nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
			cp := *existing
			cp.Items = replaced
			return &cp, nil
		},
		updateIfNotLockedFn: func(_ context.Context, _, id uint64, cmd UpdateEstimateInput) (*model.Estimate, error) {
			updatedCmd = cmd
			return &model.Estimate{ID: id, Status: model.EstimateStatusDraft}, nil
		},
		replaceItemsFn: func(_ context.Context, _, _ uint64, _ []model.EstimateItem) error {
			replaced = []model.EstimateItem{}
			return nil
		},
	}
	svc := NewEstimateService(repo, nil, nil, estimateTestMembershipCounter(), nil, noopTransactor{})
	empty := []EstimateItemInput{}
	_, err := svc.Update(context.Background(), 1, 3, &UpdateEstimateInput{Items: &empty})
	require.NoError(t, err)
	require.NotNil(t, updatedCmd.TotalAmount)
	assert.Equal(t, int64(0), *updatedCmd.TotalAmount)
	require.NotNil(t, updatedCmd.Subtotal)
	assert.Equal(t, int64(0), *updatedCmd.Subtotal)
	// 保存済み割引・保険は cmd に載せず保持。
	assert.Nil(t, updatedCmd.DiscountAmount)
	assert.Nil(t, updatedCmd.InsuranceAmount)
}
