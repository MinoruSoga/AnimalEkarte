package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMerchandiseItemRepository は MerchandiseItemRepository のテスト用モック実装
type mockMerchandiseItemRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	countUsageByMerchandiseItemFn func(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	createFn                      func(ctx context.Context, item *model.MerchandiseItem) error
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMerchandiseItemRepository) FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, category)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MerchandiseItem{ID: id, ClinicID: clinicID}, nil
}

func (m *mockMerchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error) {
	if m.countUsageByMerchandiseItemFn != nil {
		return m.countUsageByMerchandiseItemFn(ctx, clinicID, merchandiseItemID)
	}
	return 0, nil
}

func (m *mockMerchandiseItemRepository) Create(ctx context.Context, item *model.MerchandiseItem) error {
	return m.createFn(ctx, item)
}

func (m *mockMerchandiseItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockMerchandiseItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMerchandiseItemRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func newTestMerchandiseItemService(repo *mockMerchandiseItemRepository) MerchandiseItemService {
	return NewMerchandiseItemService(repo)
}

// ---- Create テスト ----

func TestMerchandiseItemService_Create(t *testing.T) {
	price100 := int64(100)
	priceNeg := int64(-1)

	tests := []struct {
		name    string
		input   *CreateMerchandiseItemInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates merchandise item successfully",
			input: &CreateMerchandiseItemInput{
				Name:      "ドッグフード",
				Category:  "food",
				UnitPrice: price100,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "applies default tax type and rate when not specified",
			input: &CreateMerchandiseItemInput{
				Name:      "デフォルト税率商品",
				UnitPrice: price100,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when name is empty",
			input: &CreateMerchandiseItemInput{
				Name: "",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when unit_price is negative",
			input: &CreateMerchandiseItemInput{
				Name:      "マイナス商品",
				UnitPrice: priceNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateMerchandiseItemInput{
				Name:      "エラー商品",
				UnitPrice: price100,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{
				createFn: func(_ context.Context, _ *model.MerchandiseItem) error {
					return tt.repoErr
				},
			}
			svc := newTestMerchandiseItemService(repo)

			item, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

// ---- Update テスト ----

func TestMerchandiseItemService_Update(t *testing.T) {
	name := "更新後商品名"
	isActive := false
	price200 := int64(200)
	priceNeg := int64(-1)
	category := "food"

	tests := []struct {
		name    string
		input   *UpdateMerchandiseItemInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates merchandise item successfully",
			input: &UpdateMerchandiseItemInput{
				Name:      &name,
				IsActive:  &isActive,
				UnitPrice: &price200,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "updates category only",
			input: &UpdateMerchandiseItemInput{
				Category: &category,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateMerchandiseItemInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when unit_price is negative",
			input: &UpdateMerchandiseItemInput{
				UnitPrice: &priceNeg,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &UpdateMerchandiseItemInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns not found error when item does not exist",
			input: &UpdateMerchandiseItemInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("merchandise_item", "999"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMerchandiseItemRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MerchandiseItem, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.MerchandiseItem{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := newTestMerchandiseItemService(repo)

			item, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
			}
		})
	}
}

// ---- Delete テスト ----

func TestMerchandiseItemService_Delete_Success(t *testing.T) {
	item := &model.MerchandiseItem{ID: 1, ClinicID: 1, Name: "ドッグフード"}
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return item, nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.NoError(t, err)
}

func TestMerchandiseItemService_Delete_NotFound(t *testing.T) {
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return nil, apperrors.WrapNotFound("merchandise_item", "999")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestMerchandiseItemService_Delete_ConflictWhenInUse(t *testing.T) {
	item := &model.MerchandiseItem{ID: 1, ClinicID: 1, Name: "ドッグフード"}
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return item, nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 2, nil // 2件の billing_items から参照
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestMerchandiseItemService_Delete_CountUsageError(t *testing.T) {
	item := &model.MerchandiseItem{ID: 1, ClinicID: 1, Name: "ドッグフード"}
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return item, nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db connection error")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	// 依存チェック失敗はNotFoundでもConflictでもなく一般エラー
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestMerchandiseItemService_Delete_RepositoryError(t *testing.T) {
	item := &model.MerchandiseItem{ID: 1, ClinicID: 1, Name: "ドッグフード"}
	repo := &mockMerchandiseItemRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MerchandiseItem, error) {
			return item, nil
		},
		countUsageByMerchandiseItemFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	svc := newTestMerchandiseItemService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}
