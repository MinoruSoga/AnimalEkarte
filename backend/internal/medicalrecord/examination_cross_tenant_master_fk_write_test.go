package medicalrecord

// examination_cross_tenant_master_fk_write_test.go — BE9-2D ⑦: internal/service
// cross_tenant_master_fk_write_test.go の examination 節（ExamType FK guard 2テスト）を
// 同名のまま縦移動（ok/rejectExamTypeRepo は本 package 既存 builder を使用）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ── examination (CRITICAL #124): exam_type_id ──

func TestExaminationService_Create_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	newSvc := func(created *bool) ExaminationService {
		repo := &mockExaminationRepository{
			createFn: func(_ context.Context, _ *model.Examination) error { *created = true; return nil },
		}
		return NewExaminationService(repo, &mockMedicalRecordRepository{}, rejectExamTypeRepo(ownedExamTypeID), &mockAuditTxLogger{}, &mockCheckupTransactor{})
	}

	t.Run("rejects cross-clinic exam_type_id and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{ExamTypeID: foreignExamTypeID, ActorID: ptrUint64(1)})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created, "examination must NOT be persisted referencing another clinic's exam_type")
	})

	t.Run("accepts same-clinic exam_type_id (no false-reject)", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{ExamTypeID: ownedExamTypeID, ActorID: ptrUint64(1)})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestExaminationService_Update_RejectsCrossClinicExamType(t *testing.T) {
	const clinicID = uint64(1)
	const ownedExamTypeID = uint64(10)
	const foreignExamTypeID = uint64(999)

	newSvc := func(updated *bool) ExaminationService {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
				return &model.Examination{ID: id, ClinicID: clinicID, Status: model.ExaminationStatusPending}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
				*updated = true
				return &model.Examination{ID: 1}, nil
			},
		}
		return NewExaminationService(repo, &mockMedicalRecordRepository{}, rejectExamTypeRepo(ownedExamTypeID), &mockAuditTxLogger{}, &mockCheckupTransactor{})
	}

	t.Run("rejects cross-clinic exam_type_id on update and does not persist", func(t *testing.T) {
		updated := false
		svc := newSvc(&updated)
		foreign := foreignExamTypeID
		out, err := svc.Update(context.Background(), clinicID, 1, UpdateExaminationInput{ExamTypeID: &foreign, ActorID: ptrUint64(1)})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated, "examination must NOT be updated to reference another clinic's exam_type")
	})
}

func TestExaminationService_Create_RejectsCrossClinicExamTypeField(t *testing.T) {
	const clinicID = uint64(1)
	const examTypeID = uint64(10)
	foreignFieldID := uint64(999)
	replaced := false
	repo := &mockExaminationRepository{
		createFn: func(_ context.Context, exam *model.Examination) error {
			exam.ID = 1
			return nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaced = true
			return nil, 0, nil
		},
	}
	examTypeRepo := &mockExamTypeRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{
				ID:    id,
				Items: []model.ExamTypeField{{ID: 100, ClinicID: clinicID}},
			}, nil
		},
	}
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		examTypeRepo,
		&mockAuditTxLogger{},
		&mockCheckupTransactor{},
	)
	items := []UpsertExamItemInput{{ExamTypeFieldID: &foreignFieldID, Name: "foreign field"}}

	out, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
		ExamTypeID: examTypeID,
		Items:      &items,
		ActorID:    ptrUint64(1),
	})

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.False(t, replaced, "foreign exam_type_field must be rejected before item persistence")
}

func TestExaminationService_Update_RejectsCrossClinicExamTypeField(t *testing.T) {
	const clinicID = uint64(1)
	const examTypeID = uint64(10)
	foreignFieldID := uint64(999)
	replaced := false
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{
				ID:         id,
				ClinicID:   clinicID,
				ExamTypeID: examTypeID,
				Status:     model.ExaminationStatusPending,
			}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaced = true
			return nil, 0, nil
		},
	}
	examTypeRepo := &mockExamTypeRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{
				ID:    id,
				Items: []model.ExamTypeField{{ID: 100, ClinicID: clinicID}},
			}, nil
		},
	}
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		examTypeRepo,
		&mockAuditTxLogger{},
		&mockCheckupTransactor{},
	)
	items := []UpsertExamItemInput{{ExamTypeFieldID: &foreignFieldID, Name: "foreign field"}}

	out, err := svc.Update(context.Background(), clinicID, 1, UpdateExaminationInput{Items: &items, ActorID: ptrUint64(1)})

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.False(t, replaced, "foreign exam_type_field must be rejected before item persistence")
}

// ── checkup (MEDIUM, clinical screening record): checkup_type_id ──
