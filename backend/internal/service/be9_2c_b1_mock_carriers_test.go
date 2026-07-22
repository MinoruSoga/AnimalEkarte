package service

// BE9-2C B① carrier with real pet/master test consumers. Delete this repository double and
// ptrString with those consumers in BE9-2E; BE9-2F is the compatibility backstop.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockInsuranceRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	createFn                  func(ctx context.Context, insurance *model.Insurance) error
	updateFn                  func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderErr                error
	countUsageByInsuranceIDFn func(ctx context.Context, clinicID, insuranceID uint64) (int64, error)
}

func (m *mockInsuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockInsuranceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	return m.createFn(ctx, insurance)
}

func (m *mockInsuranceRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockInsuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockInsuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByInsuranceIDFn != nil {
		return m.countUsageByInsuranceIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func ptrString(v string) *string { return &v }
