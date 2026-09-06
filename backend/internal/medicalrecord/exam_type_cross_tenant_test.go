package medicalrecord

// Cross-tenant master FK write isolation tests for ExamTypeService's self-referential
// parent_id guard (X-14 batch3), ported from internal/service/cross_tenant_master_fk_write_test.go
// (BE9-2C). That shared file spans many domains beyond this slice — examination_service.go,
// clinical_plan_service.go, lab_import_examination_service.go — all of which stay in
// internal/service pending BE9-2D, so the shared file was NOT split; instead these two
// specific tests were removed from it (they would no longer compile there once
// internal/service/exam_type_service.go is deleted by this batch — ExaminationTypeService /
// NewExamTypeService / CreateExamTypeInput cease to exist in that package) and recreated
// here against the moved implementation. mockExamTypeRepository is reused from this
// package's exam_type_service_test.go (ported from the sibling internal/service file of the
// same name) — not redefined here.
//
// This is the master_fk_write_inventory_lint_test.go allowlist evidence for
// "examTypeService.Create"/"examTypeService.Update" (see
// internal/service/master_fk_write_inventory_lint_test.go's masterFKWriteAllowlist entries,
// updated by this batch to note the new source location).

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestExamTypeService_Create_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(created *bool) ExaminationTypeService {
		repo := &mockExamTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
				if id != ownedParentID {
					return nil, apperrors.WrapNotFound("exam_type", "foreign")
				}
				return &model.ExaminationType{ID: id}, nil
			},
			createFn: func(_ context.Context, _ *model.ExaminationType) error { *created = true; return nil },
		}
		return NewExamTypeService(repo, passthroughExamTypeTransactor{})
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		foreign := foreignParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateExamTypeInput{Name: "x", ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "exam type must NOT be persisted referencing another clinic's parent exam type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		owned := ownedParentID
		out, err := svc.Create(context.Background(), clinicID, &CreateExamTypeInput{Name: "x", ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestExamTypeService_Update_RejectsCrossClinicParentFK(t *testing.T) {
	const clinicID = uint64(1)
	const entityID = uint64(1)
	const ownedParentID = uint64(10)
	const foreignParentID = uint64(999)

	newSvc := func(updated *bool) ExaminationTypeService {
		repo := &mockExamTypeRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
				if id == entityID || id == ownedParentID {
					return &model.ExaminationType{ID: id}, nil
				}
				return nil, apperrors.WrapNotFound("exam_type", "foreign")
			},
			updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateExamTypeInput) (*model.ExaminationType, error) {
				*updated = true
				return &model.ExaminationType{ID: id}, nil
			},
		}
		return NewExamTypeService(repo, passthroughExamTypeTransactor{})
	}

	t.Run("rejects cross-clinic parent_id and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateExamTypeInput{ParentID: &foreign})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "exam type must NOT be updated to reference another clinic's parent exam type")
	})

	t.Run("accepts same-clinic parent_id (no false-reject)", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		owned := ownedParentID
		out, err := svc.Update(context.Background(), clinicID, entityID, &UpdateExamTypeInput{ParentID: &owned})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, updated)
	})
}
