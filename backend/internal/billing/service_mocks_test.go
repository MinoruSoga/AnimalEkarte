package billing

// service_mocks_test.go — def残存（inventory系mock）→移動先で再宣言する規約の複製（B①）。

import (
	"context"

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
