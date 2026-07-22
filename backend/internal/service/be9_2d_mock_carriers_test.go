package service

// BE9-2D ⑥ carriers with real legacy master/owner test consumers. Delete each double with
// its consumer in BE9-2E; BE9-2F is the compatibility backstop.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

func strPtr(s string) *string { return &s }

type mockMedicineRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	createFn                  func(ctx context.Context, medicine *model.Medicine) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockMedicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockMedicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn != nil {
		return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
	}
	return 0, nil
}

func (m *mockMedicineRepository) CountUsageByMedicineID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	return m.createFn(ctx, medicine)
}

func (m *mockMedicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockMedicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

type mockProcedureRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	createFn                  func(ctx context.Context, procedure *model.Procedure) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	countUsageByProcedureIDFn func(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockProcedureRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockProcedureRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	return m.createFn(ctx, procedure)
}

func (m *mockProcedureRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Procedure{ID: id}, nil
}

func (m *mockProcedureRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockProcedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockProcedureRepository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	if m.countUsageByProcedureIDFn == nil {
		return 0, nil
	}
	return m.countUsageByProcedureIDFn(ctx, clinicID, procedureID)
}

func (m *mockProcedureRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

type mockConsultationRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	findByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	createFn                     func(ctx context.Context, consultation *model.Consultation) error
	updateFieldsFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	countUsageByConsultationIDFn func(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	countChildrenByParentIDFn    func(ctx context.Context, clinicID, parentID uint64) (int64, error)
	reorderErr                   error
}

func (m *mockConsultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockConsultationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	return m.createFn(ctx, consultation)
}

func (m *mockConsultationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockConsultationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockConsultationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockConsultationRepository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	if m.countUsageByConsultationIDFn == nil {
		return 0, nil
	}
	return m.countUsageByConsultationIDFn(ctx, clinicID, consultationID)
}

func (m *mockConsultationRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

func uint64Ptr(v uint64) *uint64 { return &v }
