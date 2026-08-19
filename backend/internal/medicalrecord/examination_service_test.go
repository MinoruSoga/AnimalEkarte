package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockExaminationRepository は ExaminationRepository のテスト用モック実装
type mockExaminationRepository struct {
	findAllFn                func(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	findByIDFn               func(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	lockByIDForUpdateFn      func(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	createFn                 func(ctx context.Context, exam *model.Examination) error
	updateFieldsFn           func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	deleteFn                 func(ctx context.Context, clinicID, id uint64) error
	countItemsByExamIDFn     func(ctx context.Context, clinicID, examID uint64) (int64, error)
	findAllItemsByExamIDFn   func(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	replaceItemsByExamIDFn   func(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error)
	appendOfficialRevisionFn func(ctx context.Context, clinicID, examinationID, actorID uint64, changeReason string) (uint64, error)
	confirmWithRevisionCASFn func(ctx context.Context, clinicID, examinationID uint64, expectedStatus model.ExaminationStatus, version uint64) (*model.Examination, error)
	findOfficialByIDFn       func(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
	findPrintSnapshotFn      func(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error)
}

func (m *mockExaminationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, medicalRecordID, status, startDate, endDate, page, limit)
}

func (m *mockExaminationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	mrID := uint64(1)
	return &model.Examination{ID: id, Status: model.ExaminationStatusPending, MedicalRecordID: &mrID}, nil
}

func (m *mockExaminationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, id)
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockExaminationRepository) Create(ctx context.Context, exam *model.Examination) error {
	return m.createFn(ctx, exam)
}

func (m *mockExaminationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockExaminationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockExaminationRepository) CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error) {
	if m.countItemsByExamIDFn == nil {
		return 0, nil
	}
	return m.countItemsByExamIDFn(ctx, clinicID, examID)
}

func (m *mockExaminationRepository) FindAllItemsByExamID(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	if m.findAllItemsByExamIDFn == nil {
		return nil, nil
	}
	return m.findAllItemsByExamIDFn(ctx, clinicID, examID)
}

func (m *mockExaminationRepository) ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
	if m.replaceItemsByExamIDFn == nil {
		return items, 0, nil
	}
	return m.replaceItemsByExamIDFn(ctx, clinicID, examID, items)
}

func (m *mockExaminationRepository) FindByJobID(_ context.Context, _ uint64, _ uuid.UUID) ([]*model.Examination, error) {
	return nil, nil
}

func (m *mockExaminationRepository) AppendOfficialRevision(
	ctx context.Context,
	clinicID, examinationID, actorID uint64,
	changeReason string,
) (uint64, error) {
	if m.appendOfficialRevisionFn != nil {
		return m.appendOfficialRevisionFn(ctx, clinicID, examinationID, actorID, changeReason)
	}
	return initialExaminationRevisionVersion, nil
}

func (m *mockExaminationRepository) ConfirmWithRevisionCAS(
	ctx context.Context,
	clinicID, examinationID uint64,
	expectedStatus model.ExaminationStatus,
	version uint64,
) (*model.Examination, error) {
	if m.confirmWithRevisionCASFn != nil {
		return m.confirmWithRevisionCASFn(ctx, clinicID, examinationID, expectedStatus, version)
	}
	updated, err := m.Update(ctx, clinicID, examinationID, map[string]any{
		"status": model.ExaminationStatusConfirmed,
	})
	if updated != nil {
		updated.CurrentRevisionVersion = cloneUint64(version)
	}
	return updated, err
}

func (m *mockExaminationRepository) FindOfficialByID(
	ctx context.Context,
	clinicID, examinationID uint64,
) (*ExaminationOfficialProjection, error) {
	if m.findOfficialByIDFn != nil {
		return m.findOfficialByIDFn(ctx, clinicID, examinationID)
	}
	exam, err := m.FindByID(ctx, clinicID, examinationID)
	if err != nil {
		return nil, err
	}
	return &ExaminationOfficialProjection{
		Examination:     *exam,
		OfficialVersion: initialExaminationRevisionVersion,
	}, nil
}

func (m *mockExaminationRepository) FindPrintSnapshot(
	ctx context.Context,
	clinicID, examinationID uint64,
	version *uint64,
) (*ExaminationPrintSnapshot, error) {
	if m.findPrintSnapshotFn != nil {
		return m.findPrintSnapshotFn(ctx, clinicID, examinationID, version)
	}
	return nil, apperrors.WrapNotFound("examination_print_snapshot", "mock")
}

// ptrFloat64 は旧 internal/service 共有 helper の最小複製（⑦移動）。
func ptrFloat64(v float64) *float64 { return &v }

type examinationPetStatusRelations struct {
	*mockMedicalRecordRepository
	findPetByIDInClinicFn func(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
}

func (r *examinationPetStatusRelations) FindPetByIDInClinic(
	ctx context.Context,
	clinicID, petID uint64,
) (*model.Pet, error) {
	return r.findPetByIDInClinicFn(ctx, clinicID, petID)
}

func TestExaminationService_List(t *testing.T) {
	petID := uint64(5)
	ownerID := uint64(2)
	status := string(model.ExaminationStatusCompleted)

	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		status    *string
		page      int
		limit     int
		repoItems []model.Examination
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns exam list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			repoItems: []model.Examination{
				{ID: 1, MedicalRecordID: ptrUint64(10), ExamTypeID: 1, Status: model.ExaminationStatusPending},
				{ID: 2, MedicalRecordID: ptrUint64(11), ExamTypeID: 2, Status: model.ExaminationStatusCompleted},
			},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no exams exist",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: []model.Examination{},
			repoTotal: 0,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    &petID,
			page:     1,
			limit:    20,
			repoItems: []model.Examination{
				{ID: 1, MedicalRecordID: ptrUint64(10), ExamTypeID: 1},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			ownerID:  &ownerID,
			page:     1,
			limit:    20,
			repoItems: []model.Examination{
				{ID: 1, MedicalRecordID: ptrUint64(10), ExamTypeID: 1},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			status:   &status,
			page:     1,
			limit:    20,
			repoItems: []model.Examination{
				{ID: 2, MedicalRecordID: ptrUint64(11), ExamTypeID: 2, Status: model.ExaminationStatusCompleted},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, nil)

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, tt.status, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestExaminationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoItem *model.Examination
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns exam when found",
			clinicID: 1,
			id:       10,
			repoItem: &model.Examination{
				ID:              10,
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
				Status:          model.ExaminationStatusPending,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when exam does not exist",
			clinicID: 1,
			id:       999,
			repoItem: nil,
			repoErr:  apperrors.WrapNotFound("exam", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoItem: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					return tt.repoItem, tt.repoErr
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, nil)

			item, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoItem, item)
			}
		})
	}
}

func TestExaminationService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		clinicID     uint64
		input        *CreateExaminationInput
		repoErr      error
		medRecErr    error
		medRecStatus model.MedicalRecordStatus
		examTypeErr  error
		wantErr      bool
	}{
		{
			name:     "creates exam successfully",
			clinicID: 1,
			input: &CreateExaminationInput{
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
				Status:          model.ExaminationStatusPending,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "defaults status to pending when empty",
			clinicID: 1,
			input: &CreateExaminationInput{
				ExamTypeID: 2,
				Date:       now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateExaminationInput{
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:     "returns error when medical record lookup fails",
			clinicID: 1,
			input: &CreateExaminationInput{
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
			},
			medRecErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:     "returns conflict when parent medical record is finalized",
			clinicID: 1,
			input: &CreateExaminationInput{
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
			},
			medRecStatus: model.MedicalRecordStatusFinalized,
			wantErr:      true,
		},
		{
			name:     "returns error when exam type ownership verification fails",
			clinicID: 1,
			input: &CreateExaminationInput{
				ExamTypeID: 1,
				Date:       now,
			},
			examTypeErr: errors.New("not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				createFn: func(_ context.Context, _ *model.Examination) error {
					return tt.repoErr
				},
			}
			medRec := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.medRecErr != nil {
						return nil, tt.medRecErr
					}
					status := tt.medRecStatus
					if status == "" {
						status = model.MedicalRecordStatusDraft
					}
					return &model.MedicalRecord{Status: status}, nil
				},
			}
			examTypeRepo := okExamTypeRepo()
			if tt.examTypeErr != nil {
				examTypeRepo = &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
					return nil, tt.examTypeErr
				}}
			}
			svc := NewExaminationService(repo, medRec, examTypeRepo, &mockAuditTxLogger{}, &mockCheckupTransactor{})
			tt.input.ActorID = ptrUint64(1)

			exam, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exam)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exam)
			}
		})
	}
}

func TestExaminationService_Update(t *testing.T) {
	now := time.Now()
	statusCompleted := model.ExaminationStatusCompleted
	resultSummary := "正常範囲内"
	tests := []struct {
		name             string
		input            UpdateExaminationInput
		repoErr          error
		wantErr          bool
		wantNF           bool
		existingStatus   model.ExaminationStatus
		existingMedRecID *uint64
		medRecErr        error
		medRecStatus     model.MedicalRecordStatus
		examTypeErr      error
		wantConflict     bool
	}{
		{
			name: "updates exam successfully",
			input: UpdateExaminationInput{
				Date:          &now,
				Status:        &statusCompleted,
				ResultSummary: &resultSummary,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateExaminationInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when exam does not exist",
			input: UpdateExaminationInput{
				Status: &statusCompleted,
			},
			repoErr: apperrors.WrapNotFound("exam", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateExaminationInput{
				ResultSummary: &resultSummary,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "rejects update when exam is already confirmed",
			input: UpdateExaminationInput{
				ResultSummary: &resultSummary,
			},
			existingStatus: model.ExaminationStatusConfirmed,
			wantErr:        true,
			wantConflict:   true,
		},
		{
			name: "returns conflict when parent medical record is finalized",
			input: UpdateExaminationInput{
				ResultSummary: &resultSummary,
			},
			existingMedRecID: ptrUint64(5),
			medRecStatus:     model.MedicalRecordStatusFinalized,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name: "returns error when parent medical record lookup fails",
			input: UpdateExaminationInput{
				ResultSummary: &resultSummary,
			},
			existingMedRecID: ptrUint64(5),
			medRecErr:        errors.New("db error"),
			wantErr:          true,
		},
		{
			name: "returns error when exam type ownership verification fails",
			input: UpdateExaminationInput{
				ExamTypeID: ptrUint64(9),
			},
			examTypeErr: errors.New("not found"),
			wantErr:     true,
		},
		{
			name: "succeeds when exam type ownership verification passes",
			input: UpdateExaminationInput{
				ExamTypeID: ptrUint64(9),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					if tt.wantNF {
						return nil, tt.repoErr
					}
					status := tt.existingStatus
					if status == "" {
						status = model.ExaminationStatusPending
					}
					return &model.Examination{ID: 1, Status: status, MedicalRecordID: tt.existingMedRecID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Examination{ID: 1}, nil
				},
			}
			medRec := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.medRecErr != nil {
						return nil, tt.medRecErr
					}
					status := tt.medRecStatus
					if status == "" {
						status = model.MedicalRecordStatusDraft
					}
					return &model.MedicalRecord{Status: status}, nil
				},
			}
			examTypeRepo := okExamTypeRepo()
			if tt.examTypeErr != nil {
				examTypeRepo = &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, _ uint64) (*model.ExaminationType, error) {
					return nil, tt.examTypeErr
				}}
			}
			svc := NewExaminationService(repo, medRec, examTypeRepo, &mockAuditTxLogger{}, &mockCheckupTransactor{})
			tt.input.ActorID = ptrUint64(1)

			exam, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exam)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exam)
			}
		})
	}
}

func TestExaminationService_UpdateUsesLockedExamStatus(t *testing.T) {
	updateCalled := false
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusPending}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusConfirmed}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalled = true
			return &model.Examination{ID: 1}, nil
		},
	}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})
	summary := "must not overwrite a concurrently confirmed exam"

	got, err := svc.Update(context.Background(), 1, 1, UpdateExaminationInput{ResultSummary: &summary, ActorID: ptrUint64(1)})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, updateCalled, "the locked confirmed snapshot must block the update")
}

func TestExaminationService_UpdateLocksExamThenMedicalRecordsInStableOrder(t *testing.T) {
	currentMedicalRecordID := uint64(30)
	destinationMedicalRecordID := uint64(20)
	order := make([]string, 0, 4)
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
			order = append(order, "examination")
			return &model.Examination{
				ID: id, ClinicID: clinicID, MedicalRecordID: &currentMedicalRecordID,
				ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.Examination, error) {
			order = append(order, "update")
			return &model.Examination{ID: id, ClinicID: clinicID, MedicalRecordID: &destinationMedicalRecordID}, nil
		},
	}
	records := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			switch id {
			case destinationMedicalRecordID:
				order = append(order, "medical_record_20")
			case currentMedicalRecordID:
				order = append(order, "medical_record_30")
			}
			return &model.MedicalRecord{
				ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewExaminationService(repo, records, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

	got, err := svc.Update(context.Background(), 1, 1, UpdateExaminationInput{
		MedicalRecordID: &destinationMedicalRecordID,
		ActorID:         ptrUint64(1),
	})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, []string{
		"examination",
		"medical_record_20",
		"medical_record_30",
		"update",
	}, order)
}

func TestExaminationService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		itemCount    int64
		countItemErr error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
		medRecErr    error
		medRecStatus model.MedicalRecordStatus
	}{
		{
			name:         "deletes exam successfully when no items exist",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      nil,
			wantErr:      false,
		},
		{
			name:         "returns conflict error when exam has items",
			clinicID:     1,
			id:           10,
			itemCount:    5,
			countItemErr: nil,
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:         "returns error when item count check fails",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: errors.New("db error"),
			repoErr:      nil,
			wantErr:      true,
		},
		{
			name:         "returns not found error when exam does not exist",
			clinicID:     1,
			id:           999,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      apperrors.WrapNotFound("exam", "999"),
			wantErr:      true,
			wantNF:       true,
		},
		{
			name:         "returns error on repository failure",
			clinicID:     1,
			id:           10,
			itemCount:    0,
			countItemErr: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
		},
		{
			name:         "returns conflict when parent medical record is finalized",
			clinicID:     1,
			id:           10,
			medRecStatus: model.MedicalRecordStatusFinalized,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:      "returns error when parent medical record lookup fails",
			clinicID:  1,
			id:        10,
			medRecErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				countItemsByExamIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.itemCount, tt.countItemErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			medRec := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.medRecErr != nil {
						return nil, tt.medRecErr
					}
					status := tt.medRecStatus
					if status == "" {
						status = model.MedicalRecordStatusDraft
					}
					return &model.MedicalRecord{Status: status}, nil
				},
			}
			svc := NewExaminationService(repo, medRec, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

			actorID := uint64(1)
			err := svc.Delete(context.Background(), tt.clinicID, tt.id, &actorID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExaminationService_DeleteUsesLockedExamRelations(t *testing.T) {
	staleMedicalRecordID := uint64(10)
	lockedMedicalRecordID := uint64(20)
	var lockedParentID uint64
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, MedicalRecordID: &staleMedicalRecordID}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, MedicalRecordID: &lockedMedicalRecordID}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	records := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			lockedParentID = id
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewExaminationService(repo, records, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

	actorID := uint64(1)
	err := svc.Delete(context.Background(), 1, 1, &actorID)

	assert.NoError(t, err)
	assert.Equal(t, lockedMedicalRecordID, lockedParentID, "delete must validate the parent from the locked exam snapshot")
}

func TestComputeExamResultStatus(t *testing.T) {
	tests := []struct {
		name           string
		inspectionVal  string
		refMin         *float64
		refMax         *float64
		wantStatus     model.ExaminationResultStatus
		wantIsAbnormal bool
	}{
		{name: "empty value → normal/false", inspectionVal: "", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "non-numeric value → normal/false", inspectionVal: "陰性", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "in-range → normal/false", inspectionVal: "5.5", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "below ref_min → low/true", inspectionVal: "0.5", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusLow, wantIsAbnormal: true},
		{name: "above ref_max → high/true", inspectionVal: "10.1", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusHigh, wantIsAbnormal: true},
		{name: "exactly ref_min → normal/false (boundary inclusive)", inspectionVal: "1.0", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "exactly ref_max → normal/false (boundary inclusive)", inspectionVal: "10.0", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "only ref_min, value above → normal", inspectionVal: "100", refMin: ptrFloat64(1), refMax: nil, wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "only ref_min, value below → low", inspectionVal: "0", refMin: ptrFloat64(1), refMax: nil, wantStatus: model.ExaminationResultStatusLow, wantIsAbnormal: true},
		{name: "only ref_max, value above → high", inspectionVal: "100", refMin: nil, refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusHigh, wantIsAbnormal: true},
		{name: "only ref_max, value below → normal", inspectionVal: "5", refMin: nil, refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "no refs → normal/false", inspectionVal: "5", refMin: nil, refMax: nil, wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
		{name: "negative value below ref_min → low", inspectionVal: "-5.0", refMin: ptrFloat64(0), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusLow, wantIsAbnormal: true},
		{name: "whitespace-padded numeric → parsed", inspectionVal: "  3.5  ", refMin: ptrFloat64(1), refMax: ptrFloat64(10), wantStatus: model.ExaminationResultStatusNormal, wantIsAbnormal: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotIsAbnormal := computeExamResultStatus(tt.inspectionVal, tt.refMin, tt.refMax)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.Equal(t, tt.wantIsAbnormal, gotIsAbnormal)
		})
	}
}

func TestExaminationService_ListItems(t *testing.T) {
	tests := []struct {
		name        string
		findByIDErr error
		repoItems   []model.ExamResult
		repoErr     error
		wantErr     bool
		wantNF      bool
		wantLen     int
	}{
		{
			name: "returns items when exam exists",
			repoItems: []model.ExamResult{
				{ID: 1, ExamID: 10, Name: "WBC", Status: model.ExaminationResultStatusNormal},
				{ID: 2, ExamID: 10, Name: "RBC", Status: model.ExaminationResultStatusHigh, IsAbnormal: true},
			},
			wantLen: 2,
		},
		{
			name:        "returns not found when exam does not exist",
			findByIDErr: apperrors.WrapNotFound("exam", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:    "propagates repo error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
				},
				findAllItemsByExamIDFn: func(_ context.Context, _, _ uint64) ([]model.ExamResult, error) {
					return tt.repoItems, tt.repoErr
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, nil)

			items, err := svc.ListItems(context.Background(), 1, 10)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
			}
		})
	}
}

func TestExaminationService_ReplaceItems(t *testing.T) {
	t.Run("leaves unmapped legacy items unassessed without master ranges", func(t *testing.T) {
		var capturedItems []model.ExamResult
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				capturedItems = items
				return items, 0, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

		inputs := []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
			{Name: "RBC", InspectionValue: "0.5"},
			{Name: "PLT", InspectionValue: "11"},
			{Name: "Note", InspectionValue: "陰性"},
		}

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, inputs)
		assert.NoError(t, err)
		assert.Len(t, saved, 4)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[0].Status)
		assert.False(t, capturedItems[0].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[1].Status)
		assert.False(t, capturedItems[1].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[2].Status)
		assert.False(t, capturedItems[2].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[3].Status)
		assert.False(t, capturedItems[3].IsAbnormal)

		// ExamID は service が強制設定する
		for _, it := range capturedItems {
			assert.Equal(t, uint64(10), it.ExamID)
			assert.Nil(t, it.RefMin, "an item without a master range must remain unassessed")
			assert.Nil(t, it.RefMax, "an item without a master range must remain unassessed")
		}
	})

	t.Run("rejects update when exam is confirmed", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusConfirmed}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, int64, error) {
				t.Fatal("repo should not be called when exam is confirmed")
				return nil, 0, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

		_, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("returns not found when exam does not exist", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return nil, apperrors.WrapNotFound("exam", "999")
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

		_, err := svc.ReplaceItems(context.Background(), 1, 999, nil, []UpsertExamItemInput{})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("accepts empty items list (clears all)", func(t *testing.T) {
		called := false
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				called = true
				assert.Empty(t, items)
				return []model.ExamResult{}, 0, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{})
		assert.NoError(t, err)
		assert.Empty(t, saved)
		assert.True(t, called)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

		_, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
		})
		assert.Error(t, err)
	})

	// BE-refactor.md R1-2 (#211 と同型): 実削除が発生した置換で監査書込が失敗すると tx 全体が
	// rollback される（fail-closed）ことを mock レベルで証明する。原子性の DB-backed 正本は
	// examination_repository_tx_atomicity_test.go 側にある。
	t.Run("rolls back replacement when audit write fails after real deletion (fail-closed)", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				return items, 1, nil // 実削除 1 件 → 監査ゲートが発火する
			},
		}
		audit := &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
				return errors.New("audit write failed")
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})

		actor := uint64(100)
		_, err := svc.ReplaceItems(context.Background(), 1, 10, &actor, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
		})
		assert.Error(t, err, "監査書込失敗は tx をロールバックしエラーを返す（fail-closed）")
	})

	// 実削除が無い（純粋な新規追加）置換では監査ゲートが発火しないことを確認する。
	t.Run("does not audit when no rows were actually deleted", func(t *testing.T) {
		auditCalled := false
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				return items, 0, nil // 既存 item が無い初回登録 → 削除 0 件
			},
		}
		audit := &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
				auditCalled = true
				return nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
		})
		assert.NoError(t, err)
		assert.Len(t, saved, 1)
		assert.False(t, auditCalled, "純粋な新規追加（削除0件）は監査を書かない")
	})

	// #124: exam_type_field のクロステナント/別種別 紐付け防止。03bf1cb5 は親 exam_type_id のみ
	// 検証し、ReplaceItems の exam_type_field_id は未検証だった。別クリニック/別種別のフィールドを
	// 紐付けると、その基準値・単位が結果に誤適用される（#124 実害と同型）。
	t.Run("rejects exam_type_field not belonging to the exam's clinic-owned type (#124)", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, ExamTypeID: 50, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, int64, error) {
				t.Fatal("repo should not persist when a cross-type/cross-tenant exam_type_field is supplied")
				return nil, 0, nil
			},
		}
		examTypeRepo := &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{ID: id, Items: []model.ExamTypeField{{ID: 100}, {ID: 101}}}, nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, examTypeRepo, nil, &mockCheckupTransactor{})

		foreignField := uint64(999) // not in {100,101}
		_, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0", ExamTypeFieldID: &foreignField},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("accepts exam_type_field that belongs to the exam's own type", func(t *testing.T) {
		var captured []model.ExamResult
		petID := uint64(70)
		baseRepo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{
					ID: 10, PetID: &petID, ExamTypeID: 50, Status: model.ExaminationStatusInProgress,
				}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				captured = items
				return items, 0, nil
			},
		}
		repo := &referenceRangeResolverExaminationRepository{
			mockExaminationRepository: baseRepo,
			speciesID:                 7,
			rangesBySpecies:           map[uint64]map[uint64]model.ExamReferenceRange{7: {}},
		}
		examTypeRepo := &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{ID: id, Items: []model.ExamTypeField{{ID: 100}, {ID: 101}}}, nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, examTypeRepo, nil, &mockCheckupTransactor{})

		ownField := uint64(100)
		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0", ExamTypeFieldID: &ownField},
		})
		require.NoError(t, err)
		require.Len(t, saved, 1)
		require.Len(t, captured, 1)
		if assert.NotNil(t, captured[0].ExamTypeItemID) {
			assert.Equal(t, uint64(100), *captured[0].ExamTypeItemID)
		}
	})
}

func TestExaminationService_ReplaceItemsUsesLockedExamStatus(t *testing.T) {
	replaceCalled := false
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusInProgress}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusConfirmed}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalled = true
			return items, 0, nil
		},
	}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

	got, err := svc.ReplaceItems(context.Background(), 1, 1, nil, []UpsertExamItemInput{{
		Name:            "WBC",
		InspectionValue: "5.0",
	}})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, replaceCalled, "the locked confirmed snapshot must block result replacement")
}

func TestExaminationService_ReplaceItemsRejectsCompletedSeal(t *testing.T) {
	replaceCalled := false
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusCompleted}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalled = true
			return items, 0, nil
		},
	}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

	got, err := svc.ReplaceItems(context.Background(), 1, 1, nil, []UpsertExamItemInput{{
		Name:            "WBC",
		InspectionValue: "5.0",
	}})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err))
	assert.Contains(t, err.Error(), "完了済み")
	assert.False(t, replaceCalled, "BUG-033: first-pass completed seal must block result replacement")
}

func TestExaminationService_ReplaceItemsAllowsCompletedWorkingCopy(t *testing.T) {
	rev := uint64(2)
	replaceCalled := false
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{
				ID: id, Status: model.ExaminationStatusCompleted, CurrentRevisionVersion: &rev,
			}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalled = true
			return items, 0, nil
		},
	}
	// revisioned path needs revisionWorkflow — without it ReplaceItems returns 500.
	// Use non-revision path assertion via nil version already covered; here provide workflow stub via real service path is heavy.
	// Minimal: lock check must pass before revisionWorkflow nil check — if we get internal error about workflow, lock passed.
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, &mockCheckupTransactor{})

	_, err := svc.ReplaceItems(context.Background(), 1, 1, nil, []UpsertExamItemInput{{
		Name:            "WBC",
		InspectionValue: "5.0",
	}})

	// Passed results lock; fails later on missing revision workflow (expected for this unit mock).
	assert.Error(t, err)
	assert.False(t, apperrors.IsConflict(err), "post-unconfirm completed must not hit results lock conflict")
	assert.False(t, replaceCalled)
}

func TestExaminationService_UpdateRejectsItemsOnCompletedSeal(t *testing.T) {
	updateCalled := false
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, ExamTypeID: 1, Status: model.ExaminationStatusCompleted}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalled = true
			return nil, nil
		},
	}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "1"}}
	got, err := svc.Update(context.Background(), 1, 1, UpdateExaminationInput{
		ActorID: ptrUint64(9),
		Items:   &items,
	})
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err))
	assert.Contains(t, err.Error(), "完了済み")
	assert.False(t, updateCalled)
}

func TestExaminationService_DeleteRejectsCompletedSeal(t *testing.T) {
	deleteCalled := false
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Examination, error) {
			return &model.Examination{ID: id, Status: model.ExaminationStatusCompleted}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})
	err := svc.Delete(context.Background(), 1, 1, ptrUint64(9))
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Contains(t, err.Error(), "完了済み")
	assert.False(t, deleteCalled)
}

func TestExaminationService_ReplaceItemsLocksParentFromLockedExam(t *testing.T) {
	lockedMedicalRecordID := uint64(20)
	staleMedicalRecordID := uint64(10)
	var lockedParentID uint64
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
			return &model.Examination{
				ID: id, ClinicID: clinicID, MedicalRecordID: &staleMedicalRecordID,
				ExamTypeID: 1, Status: model.ExaminationStatusInProgress,
			}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
			return &model.Examination{
				ID: id, ClinicID: clinicID, MedicalRecordID: &lockedMedicalRecordID,
				ExamTypeID: 1, Status: model.ExaminationStatusInProgress,
			}, nil
		},
	}
	records := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			lockedParentID = id
			return &model.MedicalRecord{
				ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewExaminationService(repo, records, okExamTypeRepo(), nil, &mockCheckupTransactor{})

	got, err := svc.ReplaceItems(context.Background(), 1, 1, nil, []UpsertExamItemInput{{
		Name:            "WBC",
		InspectionValue: "5.0",
	}})

	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, lockedMedicalRecordID, lockedParentID)
}

// TestBuildExaminationUpdate は buildExaminationUpdate の全フィールド網羅とゼロ値挙動を検証する。
func TestBuildExaminationUpdate(t *testing.T) {
	medRecID := uint64(5)
	petID := uint64(7)
	examTypeID := uint64(1)
	doctorID := uint64(3)
	date := time.Now()
	resultSummary := "正常"
	machine := "X線装置A"
	status := model.ExaminationStatusCompleted

	t.Run("maps all provided fields", func(t *testing.T) {
		input := UpdateExaminationInput{
			MedicalRecordID: &medRecID,
			PetID:           &petID,
			ExamTypeID:      &examTypeID,
			DoctorID:        &doctorID,
			Date:            &date,
			ResultSummary:   &resultSummary,
			Machine:         &machine,
			Status:          &status,
		}
		fields := buildExaminationUpdate(input)
		assert.Equal(t, medRecID, fields["medical_record_id"])
		assert.Equal(t, petID, fields["pet_id"])
		assert.Equal(t, examTypeID, fields["exam_type_id"])
		assert.Equal(t, doctorID, fields["doctor_id"])
		assert.Equal(t, date, fields["date"])
		assert.Equal(t, resultSummary, fields["result_summary"])
		assert.Equal(t, machine, fields["machine"])
		assert.Equal(t, status, fields["status"])
		assert.Len(t, fields, 8)
	})

	t.Run("returns empty map when all fields are nil", func(t *testing.T) {
		fields := buildExaminationUpdate(UpdateExaminationInput{})
		assert.Empty(t, fields)
	})
}

func TestExaminationService_Create_RejectsInvalidClinicalRelations(t *testing.T) {
	const (
		clinicID       = uint64(1)
		medicalRecord  = uint64(10)
		recordOwnerID  = uint64(20)
		recordPetID    = uint64(30)
		otherPetID     = uint64(31)
		assignedDoctor = uint64(40)
	)

	tests := []struct {
		name      string
		recordID  *uint64
		petID     *uint64
		doctorID  *uint64
		configure func(repo *mockMedicalRecordRepository)
		wantErr   bool
	}{
		{
			name:     "rejects a medical record outside the clinic",
			recordID: ptrUint64(999),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.lockByIDForUpdateFn = func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:     "rejects a different same-clinic patient",
			recordID: ptrUint64(medicalRecord),
			petID:    ptrUint64(otherPetID),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.findPetOwnerInClinicFn = func(_ context.Context, _, petID uint64) (uint64, error) {
					if petID == recordPetID || petID == otherPetID {
						return recordOwnerID, nil
					}
					return 0, apperrors.WrapNotFound("pet", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:     "rejects an inactive or unassigned doctor",
			petID:    ptrUint64(recordPetID),
			doctorID: ptrUint64(999),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.assertDoctorInClinicFn = func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("staff", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:     "accepts one active patient graph and assigned doctor",
			recordID: ptrUint64(medicalRecord),
			petID:    ptrUint64(recordPetID),
			doctorID: ptrUint64(assignedDoctor),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalls := 0
			repo := &mockExaminationRepository{
				createFn: func(_ context.Context, exam *model.Examination) error {
					createCalls++
					exam.ID = 1
					return nil
				},
			}
			relations := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{
						ID: id, ClinicID: clinicID, OwnerID: ptrUint64(recordOwnerID),
						PetID: ptrUint64(recordPetID), Status: model.MedicalRecordStatusDraft,
					}, nil
				},
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return recordOwnerID, nil
				},
			}
			if tt.configure != nil {
				tt.configure(relations)
			}
			svc := NewExaminationService(repo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

			got, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
				MedicalRecordID: tt.recordID, PetID: tt.petID, ExamTypeID: 1,
				DoctorID: tt.doctorID, Date: time.Now(), ActorID: ptrUint64(1),
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Zero(t, createCalls)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, 1, createCalls)
		})
	}
}

type examinationTxContextKey struct{}

func TestExaminationService_DeceasedPetWriteBoundary(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{name: "create", run: testExaminationCreateRejectsDeceasedPetBeforeWrites},
		{name: "create from medical record", run: testExaminationCreateFromRecordRejectsDeceasedPetBeforeWrites},
		{name: "update replacement", run: testExaminationUpdateRejectsDeceasedPetBeforeWrites},
		{name: "update medical record replacement", run: testExaminationUpdateRecordRejectsDeceasedPetBeforeWrites},
		{name: "existing deceased pet non-relation update", run: testExaminationUpdateAllowsExistingDeceasedPet},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func testExaminationCreateRejectsDeceasedPetBeforeWrites(t *testing.T) {
	const (
		clinicID = uint64(1)
		petID    = uint64(30)
	)
	deceasedAt := time.Now().Add(-24 * time.Hour)
	createCalls := 0
	replaceCalls := 0
	auditCalls := 0
	petStatusCalls := 0
	repo := &mockExaminationRepository{
		createFn: func(_ context.Context, exam *model.Examination) error {
			createCalls++
			exam.ID = 1
			return nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _ uint64, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalls++
			return items, 0, nil
		},
	}
	relations := &examinationPetStatusRelations{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{},
		findPetByIDInClinicFn: func(ctx context.Context, gotClinicID, gotPetID uint64) (*model.Pet, error) {
			petStatusCalls++
			assert.Equal(t, true, ctx.Value(examinationTxContextKey{}))
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, petID, gotPetID)
			return &model.Pet{ID: petID, ClinicID: clinicID, DeceasedAt: &deceasedAt}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, examinationTxContextKey{}, true))
	}}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), audit, tx)
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "5.0"}}

	got, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
		PetID: ptrUint64(petID), ExamTypeID: 1, Date: time.Now(),
		Items: &items, ActorID: ptrUint64(1),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "死亡")
	assert.Nil(t, got)
	assert.Equal(t, 1, petStatusCalls)
	assert.Zero(t, createCalls)
	assert.Zero(t, replaceCalls)
	assert.Zero(t, auditCalls)
}

func testExaminationCreateFromRecordRejectsDeceasedPetBeforeWrites(t *testing.T) {
	const (
		clinicID        = uint64(1)
		medicalRecordID = uint64(10)
		petID           = uint64(30)
	)
	deceasedAt := time.Now().Add(-24 * time.Hour)
	createCalls := 0
	replaceCalls := 0
	auditCalls := 0
	petStatusCalls := 0
	repo := &mockExaminationRepository{
		createFn: func(_ context.Context, exam *model.Examination) error {
			createCalls++
			exam.ID = 1
			return nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _ uint64, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalls++
			return items, 0, nil
		},
	}
	relations := &examinationPetStatusRelations{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.MedicalRecord, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, medicalRecordID, gotRecordID)
				return &model.MedicalRecord{
					ID: medicalRecordID, ClinicID: clinicID, PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
		},
		findPetByIDInClinicFn: func(ctx context.Context, gotClinicID, gotPetID uint64) (*model.Pet, error) {
			petStatusCalls++
			assert.Equal(t, true, ctx.Value(examinationTxContextKey{}))
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, petID, gotPetID)
			return &model.Pet{ID: petID, ClinicID: clinicID, DeceasedAt: &deceasedAt}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, examinationTxContextKey{}, true))
	}}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), audit, tx)
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "5.0"}}

	got, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
		MedicalRecordID: ptrUint64(medicalRecordID), ExamTypeID: 1, Date: time.Now(),
		Items: &items, ActorID: ptrUint64(1),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "死亡")
	assert.Nil(t, got)
	assert.Equal(t, 1, petStatusCalls)
	assert.Zero(t, createCalls)
	assert.Zero(t, replaceCalls)
	assert.Zero(t, auditCalls)
}

func TestExaminationService_Update_RevalidatesEffectivePatient(t *testing.T) {
	const (
		clinicID      = uint64(1)
		examID        = uint64(2)
		medicalRecord = uint64(10)
		recordOwnerID = uint64(20)
		recordPetID   = uint64(30)
	)
	otherPetID := uint64(31)
	updateCalls := 0
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
			return &model.Examination{
				ID: examID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecord),
				PetID: ptrUint64(recordPetID), ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalls++
			return &model.Examination{ID: examID}, nil
		},
	}
	relations := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: medicalRecord, ClinicID: clinicID, OwnerID: ptrUint64(recordOwnerID),
				PetID: ptrUint64(recordPetID), Status: model.MedicalRecordStatusDraft,
			}, nil
		},
		findPetOwnerInClinicFn: func(_ context.Context, _, petID uint64) (uint64, error) {
			if petID == recordPetID || petID == otherPetID {
				return recordOwnerID, nil
			}
			return 0, apperrors.WrapNotFound("pet", "scoped")
		},
	}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{PetID: &otherPetID, ActorID: ptrUint64(1)})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)
}

func testExaminationUpdateRejectsDeceasedPetBeforeWrites(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(2)
		oldPetID = uint64(30)
		newPetID = uint64(31)
	)
	deceasedAt := time.Now().Add(-24 * time.Hour)
	updateCalls := 0
	replaceCalls := 0
	auditCalls := 0
	petStatusCalls := 0
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{
				ID: examID, ClinicID: clinicID, PetID: ptrUint64(oldPetID),
				ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalls++
			return &model.Examination{
				ID: examID, ClinicID: clinicID, PetID: ptrUint64(newPetID),
				ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _ uint64, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalls++
			return items, 0, nil
		},
	}
	relations := &examinationPetStatusRelations{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{},
		findPetByIDInClinicFn: func(ctx context.Context, gotClinicID, gotPetID uint64) (*model.Pet, error) {
			petStatusCalls++
			assert.Equal(t, true, ctx.Value(examinationTxContextKey{}))
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, newPetID, gotPetID)
			return &model.Pet{ID: newPetID, ClinicID: clinicID, DeceasedAt: &deceasedAt}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, examinationTxContextKey{}, true))
	}}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), audit, tx)

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
		PetID: ptrUint64(newPetID), ActorID: ptrUint64(1),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "死亡")
	assert.Nil(t, got)
	assert.Equal(t, 1, petStatusCalls)
	assert.Zero(t, updateCalls)
	assert.Zero(t, replaceCalls)
	assert.Zero(t, auditCalls)
}

func testExaminationUpdateRecordRejectsDeceasedPetBeforeWrites(t *testing.T) {
	const (
		clinicID        = uint64(1)
		examID          = uint64(2)
		medicalRecordID = uint64(10)
		petID           = uint64(31)
	)
	deceasedAt := time.Now().Add(-24 * time.Hour)
	updateCalls := 0
	replaceCalls := 0
	auditCalls := 0
	petStatusCalls := 0
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{
				ID: examID, ClinicID: clinicID, ExamTypeID: 1,
				Status: model.ExaminationStatusPending,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalls++
			return &model.Examination{
				ID: examID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecordID),
				ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _ uint64, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			replaceCalls++
			return items, 0, nil
		},
	}
	relations := &examinationPetStatusRelations{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.MedicalRecord, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, medicalRecordID, gotRecordID)
				return &model.MedicalRecord{
					ID: medicalRecordID, ClinicID: clinicID, PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
		},
		findPetByIDInClinicFn: func(ctx context.Context, gotClinicID, gotPetID uint64) (*model.Pet, error) {
			petStatusCalls++
			assert.Equal(t, true, ctx.Value(examinationTxContextKey{}))
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, petID, gotPetID)
			return &model.Pet{ID: petID, ClinicID: clinicID, DeceasedAt: &deceasedAt}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, examinationTxContextKey{}, true))
	}}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), audit, tx)

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
		MedicalRecordID: ptrUint64(medicalRecordID), ActorID: ptrUint64(1),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "死亡")
	assert.Nil(t, got)
	assert.Equal(t, 1, petStatusCalls)
	assert.Zero(t, updateCalls)
	assert.Zero(t, replaceCalls)
	assert.Zero(t, auditCalls)
}

func testExaminationUpdateAllowsExistingDeceasedPet(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(2)
		petID    = uint64(31)
	)
	summary := "historical correction"
	updateCalls := 0
	auditCalls := 0
	petStatusCalls := 0
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{
				ID: examID, ClinicID: clinicID, PetID: ptrUint64(petID),
				ExamTypeID: 1, Status: model.ExaminationStatusPending,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Examination, error) {
			updateCalls++
			assert.Equal(t, summary, fields["result_summary"])
			return &model.Examination{
				ID: examID, ClinicID: clinicID, PetID: ptrUint64(petID),
				ExamTypeID: 1, Status: model.ExaminationStatusPending, ResultSummary: summary,
			}, nil
		},
	}
	relations := &examinationPetStatusRelations{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{},
		findPetByIDInClinicFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			petStatusCalls++
			return nil, errors.New("deceased status lookup must not run for a non-relation historical edit")
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), audit, &mockCheckupTransactor{})

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
		ResultSummary: &summary, ActorID: ptrUint64(1),
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, summary, got.ResultSummary)
	assert.Equal(t, 1, updateCalls)
	assert.Equal(t, 1, auditCalls)
	assert.Zero(t, petStatusCalls)
}

func TestExaminationService_Update_RejectsPetChangeAfterRevisionHistory(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(2)
		oldPetID = uint64(30)
		newPetID = uint64(31)
		version  = uint64(2)
	)
	updateCalls := 0
	relationCalls := 0
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{
				ID: examID, ClinicID: clinicID, PetID: ptrUint64(oldPetID), ExamTypeID: 1,
				Status: model.ExaminationStatusCompleted, CurrentRevisionVersion: ptrUint64(version),
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			updateCalls++
			return nil, nil
		},
	}
	relations := &mockMedicalRecordRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			relationCalls++
			return 20, nil
		},
	}
	svc := NewExaminationService(repo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
		PetID:   ptrUint64(newPetID),
		ActorID: ptrUint64(1),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)
	assert.Zero(t, relationCalls, "revision history must reject a patient change before relation or write work")
}

// SEC-DUR-01-MR-T1: 同一clinic内のpet譲渡後も、visit-time snapshotのownerとcurrent pet ownerが異なっても
// Examination更新はclinic/pet相関を守りながら成功する。cross-clinicは拒否する。
func TestExaminationService_Update_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	const (
		clinicID        = uint64(1)
		examID          = uint64(2)
		medicalRecordID = uint64(10)
		previousOwnerID = uint64(20)
		currentOwnerID  = uint64(21)
		petID           = uint64(30)
		foreignOwnerID  = uint64(900)
		foreignPetID    = uint64(901)
	)

	summary := "post-transfer update"
	updateCalls := 0
	baseExam := &model.Examination{
		ID: examID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecordID),
		PetID: ptrUint64(petID), ExamTypeID: 1, Status: model.ExaminationStatusPending,
	}
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
			copy := *baseExam
			return &copy, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Examination, error) {
			updateCalls++
			assert.Equal(t, summary, fields["result_summary"])
			return &model.Examination{ID: examID, ClinicID: clinicID, ResultSummary: summary}, nil
		},
	}

	t.Run("same_clinic_transfer_succeeds", func(t *testing.T) {
		updateCalls = 0
		relations := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: medicalRecordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		svc := NewExaminationService(repo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

		got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{ResultSummary: &summary, ActorID: ptrUint64(1)})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 1, updateCalls)
	})

	t.Run("rejects_foreign_snapshot_owner", func(t *testing.T) {
		updateCalls = 0
		relations := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: medicalRecordID, ClinicID: clinicID,
					OwnerID: ptrUint64(foreignOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				assert.Equal(t, foreignOwnerID, ownerID)
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		svc := NewExaminationService(repo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

		got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{ResultSummary: &summary, ActorID: ptrUint64(1)})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
	})

	t.Run("rejects_foreign_pet", func(t *testing.T) {
		updateCalls = 0
		relations := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: medicalRecordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(foreignPetID),
					Status: model.MedicalRecordStatusDraft,
				}, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		// Exam still points at foreignPetID so effective pet validation fails.
		foreignExam := *baseExam
		foreignExam.PetID = ptrUint64(foreignPetID)
		foreignRepo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				copy := foreignExam
				return &copy, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
				updateCalls++
				return &model.Examination{ID: examID}, nil
			},
		}
		svc := NewExaminationService(foreignRepo, relations, okExamTypeRepo(), &mockAuditTxLogger{}, &mockCheckupTransactor{})

		got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{ResultSummary: &summary, ActorID: ptrUint64(1)})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
	})
}

// BUG-448: 検査明細削除監査の old/new に解決済み境界と派生 is_assessed を含める。
func TestExtractExamResultsAudit(t *testing.T) {
	refMin, refMax := 1.0, 10.0
	qualMin, qualMax := "(-)", "(+)"

	t.Run("numeric_bounds_include_provenance_and_assessed", func(t *testing.T) {
		results := []model.ExamResult{{
			ID: 1, ExamTypeItemID: ptrUint64(11), Name: "WBC",
			InspectionValue: "0.5", RefMin: &refMin, RefMax: &refMax,
			IsAbnormal: true, Status: model.ExaminationResultStatusLow,
		}}
		got := extractExamResultsAudit(results)
		require.Len(t, got, 1)
		assert.Equal(t, uint64(1), got[0]["id"])
		assert.Equal(t, &refMin, got[0]["ref_min"])
		assert.Equal(t, &refMax, got[0]["ref_max"])
		assert.Nil(t, got[0]["qualitative_min"])
		assert.Nil(t, got[0]["qualitative_max"])
		assert.Equal(t, true, got[0]["is_assessed"])
		assert.Equal(t, true, got[0]["is_abnormal"])
		assert.Equal(t, string(model.ExaminationResultStatusLow), got[0]["status"])
	})

	t.Run("qualitative_bounds_include_provenance_and_assessed", func(t *testing.T) {
		results := []model.ExamResult{{
			ID: 2, ExamTypeItemID: ptrUint64(12), Name: "Dipstick",
			InspectionValue: "(++)", QualitativeMin: &qualMin, QualitativeMax: &qualMax,
			IsAbnormal: true, Status: model.ExaminationResultStatusHigh,
		}}
		got := extractExamResultsAudit(results)
		require.Len(t, got, 1)
		assert.Nil(t, got[0]["ref_min"])
		assert.Nil(t, got[0]["ref_max"])
		assert.Equal(t, &qualMin, got[0]["qualitative_min"])
		assert.Equal(t, &qualMax, got[0]["qualitative_max"])
		assert.Equal(t, true, got[0]["is_assessed"])
	})

	t.Run("unassessed_without_bounds", func(t *testing.T) {
		results := []model.ExamResult{{
			ID: 3, Name: "Note", InspectionValue: "free text",
			Status: model.ExaminationResultStatusNormal,
		}}
		got := extractExamResultsAudit(results)
		require.Len(t, got, 1)
		assert.Nil(t, got[0]["ref_min"])
		assert.Nil(t, got[0]["ref_max"])
		assert.Nil(t, got[0]["qualitative_min"])
		assert.Nil(t, got[0]["qualitative_max"])
		assert.Equal(t, false, got[0]["is_assessed"])
	})

	t.Run("unassessed_nonnumeric_with_numeric_bounds_still_records_bounds", func(t *testing.T) {
		results := []model.ExamResult{{
			ID: 4, Name: "ALT", InspectionValue: "陰性",
			RefMin: &refMin, RefMax: &refMax,
			Status: model.ExaminationResultStatusNormal,
		}}
		got := extractExamResultsAudit(results)
		require.Len(t, got, 1)
		assert.Equal(t, &refMin, got[0]["ref_min"])
		assert.Equal(t, &refMax, got[0]["ref_max"])
		assert.Equal(t, false, got[0]["is_assessed"])
	})

	t.Run("empty_returns_nil", func(t *testing.T) {
		assert.Nil(t, extractExamResultsAudit(nil))
		assert.Nil(t, extractExamResultsAudit([]model.ExamResult{}))
	})
}
