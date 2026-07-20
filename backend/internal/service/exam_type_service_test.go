package service

// mockExamTypeRepository is a shared test double for repository.ExamTypeRepository. Its own
// ExamTypeService test suite moved to internal/medicalrecord/exam_type_service_test.go
// (BE9-2C — internal/service/exam_type_service.go was deleted, zero remaining fan-in). This
// mock stays here because cross_tenant_master_fk_write_test.go's ExaminationService and
// LabImportExaminationService tests (both still in internal/service, out of this batch's
// scope) construct it via okExamTypeRepo()/rejectExamTypeRepo() and need it to keep
// compiling.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockExamTypeRepository は ExamTypeRepository のテスト用モック実装
type mockExamTypeRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	createFn                  func(ctx context.Context, exType *model.ExaminationType) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
	countUsageByExamTypeIDFn  func(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

func (m *mockExamTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockExamTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockExamTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	return m.createFn(ctx, exType)
}

func (m *mockExamTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockExamTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockExamTypeRepository) ReplaceItems(ctx context.Context, examTypeID uint64, items []model.ExamTypeField) error {
	return nil
}

func (m *mockExamTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockExamTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
	if m.countUsageByExamTypeIDFn == nil {
		return 0, nil
	}
	return m.countUsageByExamTypeIDFn(ctx, clinicID, examTypeID)
}

func (m *mockExamTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}
