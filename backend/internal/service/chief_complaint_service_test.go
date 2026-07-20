package service

// mockChiefComplaintTypeRepository is a shared test double for
// repository.ChiefComplaintTypeRepository. Its own ChiefComplaintTypeService test suite
// moved to internal/medicalrecord/chief_complaint_service_test.go (BE9-2C —
// internal/service/chief_complaint_service.go was deleted, zero remaining fan-in). This mock
// stays here because cross_tenant_master_fk_write_test.go's MedicalRecordService test and
// medical_record_subrecords_test.go (still in internal/service) construct it via
// okChiefComplaintTypeRepo()/rejectChiefComplaintTypeRepo() and need it to keep compiling.
// The InquiryService cross-tenant test moved to internal/medicalrecord in BE9-2D (which has its
// own mockChiefComplaintTypeRepository + helper copies).

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintType モック ----

type mockChiefComplaintTypeRepository struct {
	findAllFn                        func(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	findByIDFn                       func(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	countUsageByChiefComplaintTypeFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	createFn                         func(ctx context.Context, category *model.ChiefComplaintType) error
	updateFieldsFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error)
	deleteFn                         func(ctx context.Context, clinicID, id uint64) error
	reorderErr                       error
}

func (m *mockChiefComplaintTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockChiefComplaintTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ChiefComplaintType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockChiefComplaintTypeRepository) CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByChiefComplaintTypeFn != nil {
		return m.countUsageByChiefComplaintTypeFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockChiefComplaintTypeRepository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	return m.createFn(ctx, category)
}

func (m *mockChiefComplaintTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.ChiefComplaintType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockChiefComplaintTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockChiefComplaintTypeRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}
