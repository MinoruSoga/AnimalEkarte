package service

// This reservation-staff carrier has real trimming-service test consumers. Delete it with the
// trimming consumer in BE9-2E; BE9-2F is the compatibility backstop.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// testIntegrationKeyHex is a deterministic test-only key consumed by
// line_reservation_setting_cipher_wiring_test.go. Remove it with that legacy
// service wiring test in BE9-2F.
const testIntegrationKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type mockReservationStaffRepository struct {
	findAllFn                              func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                             func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	createFn                               func(ctx context.Context, staff *model.Staff, clinicID uint64) error
	updateFieldsFn                         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                               func(ctx context.Context, clinicID, id uint64) error
	countUsageByStaffIDFn                  func() (int64, error)
	swapSortOrderFn                        func(ctx context.Context, clinicID, id uint64, direction string) error
	findExcludedReservationTypesFn         func(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	findExcludedReservationTypesByStaffIDs func(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	replaceExcludedReservationTypesFn      func(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error
	findCapabilitiesFn                     func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	findCapabilitiesByStaffIDsFn           func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
	replaceCapabilitiesFn                  func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	supportsReservationTypeFn              func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return m.createFn(ctx, staff, clinicID)
}

func (m *mockReservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockReservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	if m.countUsageByStaffIDFn != nil {
		return m.countUsageByStaffIDFn()
	}
	return 0, nil
}

func (m *mockReservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.swapSortOrderFn != nil {
		return m.swapSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, staffID)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesByStaffIDs != nil {
		return m.findExcludedReservationTypesByStaffIDs(ctx, staffIDs)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	if m.replaceExcludedReservationTypesFn != nil {
		return m.replaceExcludedReservationTypesFn(ctx, clinicID, staffID, courseIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesFn != nil {
		return m.findCapabilitiesFn(ctx, clinicID, staffID)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesByStaffIDsFn != nil {
		return m.findCapabilitiesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.replaceCapabilitiesFn != nil {
		return m.replaceCapabilitiesFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}
