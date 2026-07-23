package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockInventoryRepository is a BE9-2E compatibility carrier for
// cross_tenant_master_fk_write_test.go. Delete in BE9-2F when that legacy
// service-package test constructs the inventory domain service directly.
type mockInventoryRepository struct {
	findAllFn        func(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	findByIDFn       func(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	createFn         func(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	updateFieldsFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
	countUsageByIDFn func(ctx context.Context, clinicID, inventoryID uint64) (int64, error)
	decreaseStockFn  func(ctx context.Context, clinicID, id uint64, quantity float64) error
}

func (m *mockInventoryRepository) FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	return m.findAllFn(ctx, clinicID, category, status, page, limit)
}

func (m *mockInventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockInventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	return m.createFn(ctx, clinicID, item)
}

func (m *mockInventoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockInventoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInventoryRepository) DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error {
	if m.decreaseStockFn != nil {
		return m.decreaseStockFn(ctx, clinicID, id, quantity)
	}
	return nil
}

func (m *mockInventoryRepository) CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error) {
	if m.countUsageByIDFn != nil {
		return m.countUsageByIDFn(ctx, clinicID, inventoryID)
	}
	return 0, nil
}

func (m *mockInventoryRepository) DeleteByNameAndMedicineCategory(_ context.Context, _ uint64, _ string) error {
	return nil
}

func (m *mockInventoryRepository) UpdateNameByMedicineCategory(_ context.Context, _ uint64, _, _ string) error {
	return nil
}
