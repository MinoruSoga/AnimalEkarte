package service

// mockDiagnosisTypeRepository / mockDiagnosisNameRepository are shared test doubles for
// repository.DiagnosisTypeRepository / repository.DiagnosisNameRepository. Their own
// DiagnosisTypeService/DiagnosisNameService test suite moved to
// internal/medicalrecord/diagnosis_service_test.go (BE9-2C —
// internal/service/diagnosis_service.go was deleted, zero remaining fan-in). These mocks
// stay here because cross_tenant_master_fk_write_test.go's ClinicalPlanService tests (still
// in internal/service, out of this batch's scope) construct them via
// okDiagnosisTypeRepo()/rejectDiagnosisTypeRepo()/okDiagnosisNameRepo()/
// rejectDiagnosisNameRepo() and need them to keep compiling.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DiagnosisType モック ----

type mockDiagnosisTypeRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	createFn                  func(ctx context.Context, category *model.DiagnosisType) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
	countChildrenByParentIDFn func(ctx context.Context, clinicID, categoryID uint64) (int64, error)
}

func (m *mockDiagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.DiagnosisType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisTypeRepository) Create(ctx context.Context, category *model.DiagnosisType) error {
	return m.createFn(ctx, category)
}

func (m *mockDiagnosisTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.DiagnosisType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockDiagnosisTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, categoryID)
}

// ---- DiagnosisName モック ----

type mockDiagnosisNameRepository struct {
	findAllFn                             func(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findByCategoryIDFn                    func(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	findAllByFilterFn                     func(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	findByIDFn                            func(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	createFn                              func(ctx context.Context, name *model.DiagnosisName) error
	updateFieldsFn                        func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error)
	deleteFn                              func(ctx context.Context, clinicID, id uint64) error
	reorderFn                             func(ctx context.Context, clinicID uint64, ids []uint64) error
	countClinicalPlansByDiagnosisNameIDFn func(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error)
}

func (m *mockDiagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindAllByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	return m.findByCategoryIDFn(ctx, clinicID, categoryID, page, limit)
}

func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	if m.findAllByFilterFn != nil {
		return m.findAllByFilterFn(ctx, clinicID, typeID)
	}
	return []model.DiagnosisName{}, nil
}

func (m *mockDiagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.DiagnosisName{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	return m.createFn(ctx, name)
}

func (m *mockDiagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.DiagnosisName{ID: id, ClinicID: clinicID}, nil
}

func (m *mockDiagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockDiagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockDiagnosisNameRepository) CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error) {
	if m.countClinicalPlansByDiagnosisNameIDFn != nil {
		return m.countClinicalPlansByDiagnosisNameIDFn(ctx, clinicID, diagnosisNameID)
	}
	return 0, nil
}

func (m *mockDiagnosisNameRepository) FindAllActive(_ context.Context, _ uint64, _ *uint64) ([]model.DiagnosisName, error) {
	return []model.DiagnosisName{}, nil
}
