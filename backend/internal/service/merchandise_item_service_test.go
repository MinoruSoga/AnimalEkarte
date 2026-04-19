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
	return m.findByIDFn(ctx, clinicID, id)
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

func (m *mockMerchandiseItemRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MerchandiseItem, error) {
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
