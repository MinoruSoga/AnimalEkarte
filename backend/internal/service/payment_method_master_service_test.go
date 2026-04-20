package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockPaymentMethodMasterRepository は PaymentMethodMasterRepository のテスト用モック実装
type mockPaymentMethodMasterRepository struct {
	findAllFn        func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	findByIDFn       func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	createFn         func(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
	updateFieldsFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
	countUsageByIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	reorderFn        func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockPaymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Create(ctx context.Context, pm *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
	if m.createFn != nil {
		return m.createFn(ctx, pm)
	}
	return pm, nil
}

func (m *mockPaymentMethodMasterRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPaymentMethodMasterRepository) CountUsageByID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByIDFn != nil {
		return m.countUsageByIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockPaymentMethodMasterRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- テスト ----

func TestPaymentMethodMasterService_List(t *testing.T) {
	methods := []model.PaymentMethodMaster{
		{ID: 1, ClinicID: 1, Name: "現金", DisplayOrder: 1},
		{ID: 2, ClinicID: 1, Name: "クレジット", DisplayOrder: 2},
	}

	tests := []struct {
		name       string
		findAllFn  func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		wantErr    bool
		wantResult []model.PaymentMethodMaster
	}{
		{
			name: "正常: 全件を返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return methods, nil
			},
			wantResult: methods,
		},
		{
			name: "エラー: DB エラーを返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{findAllFn: tt.findAllFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.List(context.Background(), 1)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestPaymentMethodMasterService_Create(t *testing.T) {
	created := &model.PaymentMethodMaster{
		ID:           1,
		ClinicID:     1,
		Name:         "クレジット",
		DisplayOrder: 2,
	}

	tests := []struct {
		name       string
		input      *CreatePaymentMethodInput
		createFn   func(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
		wantErr    bool
		wantResult *model.PaymentMethodMaster
	}{
		{
			name: "正常: 作成済みレコードを返す",
			input: &CreatePaymentMethodInput{
				Name:         "クレジット",
				DisplayOrder: 2,
			},
			createFn: func(_ context.Context, _ *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
				return created, nil
			},
			wantResult: created,
		},
		{
			name: "エラー: DB エラーを返す",
			input: &CreatePaymentMethodInput{
				Name:         "クレジット",
				DisplayOrder: 2,
			},
			createFn: func(_ context.Context, _ *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{createFn: tt.createFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.Create(context.Background(), 1, tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestPaymentMethodMasterService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		countUsageByIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
		deleteFn         func(ctx context.Context, clinicID, id uint64) error
		wantErr          bool
		wantErrIs        error
	}{
		{
			name: "正常: 未使用の支払方法を削除",
			id:   1,
			countUsageByIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		},
		{
			name: "エラー: 使用中の支払方法 → ErrConflict",
			id:   2,
			countUsageByIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 5, nil
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrConflict,
		},
		{
			name: "エラー: CountUsageByID がエラーを返す",
			id:   3,
			countUsageByIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "エラー: Delete がエラーを返す",
			id:   4,
			countUsageByIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{
				countUsageByIDFn: tt.countUsageByIDFn,
				deleteFn:         tt.deleteFn,
			}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			err := svc.Delete(context.Background(), 1, tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPaymentMethodMasterService_Reorder(t *testing.T) {
	tests := []struct {
		name      string
		ids       []uint64
		reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
		wantErr   bool
	}{
		{
			name: "正常: 並び順を更新する",
			ids:  []uint64{2, 1, 3},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return nil
			},
		},
		{
			name:    "エラー: IDリストが空",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name: "エラー: DB エラーを返す",
			ids:  []uint64{1, 2},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{reorderFn: tt.reorderFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			err := svc.Reorder(context.Background(), 1, tt.ids)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
