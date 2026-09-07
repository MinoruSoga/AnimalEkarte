package staff

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

func strPtr(value string) *string {
	return &value
}

type mockPermissionGroupRepository struct {
	findAllGroupIDsByStaffIDFn func(context.Context, uint64, uint64) ([]uint64, error)
	updateStaffGroupsFn        func(context.Context, uint64, uint64, []uint64) error
}

func (m *mockPermissionGroupRepository) FindAllGroupIDsByStaffID(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]uint64, error) {
	if m.findAllGroupIDsByStaffIDFn != nil {
		return m.findAllGroupIDsByStaffIDFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) UpdateStaffGroups(
	ctx context.Context,
	clinicID, staffID uint64,
	groupIDs []uint64,
) error {
	if m.updateStaffGroupsFn != nil {
		return m.updateStaffGroupsFn(ctx, clinicID, staffID, groupIDs)
	}
	return nil
}

type mockOccupationRepository struct {
	findAllFn                  func(context.Context, uint64) ([]model.Occupation, error)
	findByIDFn                 func(context.Context, uint64, uint64) (*model.Occupation, error)
	lockForShareFn             func(context.Context, uint64, uint64) (*model.Occupation, error)
	lockForUpdateFn            func(context.Context, uint64, uint64) (*model.Occupation, error)
	createFn                   func(context.Context, *model.Occupation) error
	updateFieldsFn             func(context.Context, uint64, uint64, UpdateOccupationInput) (*model.Occupation, error)
	deleteFn                   func(context.Context, uint64, uint64) error
	reorderErr                 error
	countUsageByOccupationIDFn func(context.Context, uint64, uint64) (int64, error)
	withTxFn                   func(context.Context, func(context.Context) error) error
	withTxCalls                int
}

func (m *mockOccupationRepository) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	m.withTxCalls++
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockOccupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.Occupation{}, nil
}

func (m *mockOccupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Occupation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOccupationRepository) LockActiveByIDForShare(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	if m.lockForShareFn != nil {
		return m.lockForShareFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Occupation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOccupationRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	if m.lockForUpdateFn != nil {
		return m.lockForUpdateFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Occupation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOccupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, occupation)
	}
	return nil
}

func (m *mockOccupationRepository) Update(
	ctx context.Context,
	clinicID, id uint64,
	cmd UpdateOccupationInput,
) (*model.Occupation, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
	}
	return &model.Occupation{}, nil
}

func (m *mockOccupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockOccupationRepository) Reorder(context.Context, uint64, []uint64) error {
	return m.reorderErr
}

func (m *mockOccupationRepository) CountUsageByOccupationID(
	ctx context.Context,
	clinicID, id uint64,
) (int64, error) {
	if m.countUsageByOccupationIDFn != nil {
		return m.countUsageByOccupationIDFn(ctx, clinicID, id)
	}
	return 0, nil
}
