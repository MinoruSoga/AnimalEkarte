package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockExaminationRepository は ExaminationRepository のテスト用モック実装
type mockExaminationRepository struct {
	findAllFn              func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	createFn               func(ctx context.Context, exam *model.Examination) error
	updateFieldsFn         func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	countItemsByExamIDFn   func(ctx context.Context, clinicID, examID uint64) (int64, error)
	findAllItemsByExamIDFn func(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	replaceItemsByExamIDFn func(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error)
}

func (m *mockExaminationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockExaminationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	mrID := uint64(1)
	return &model.Examination{ID: id, Status: model.ExaminationStatusCompleted, MedicalRecordID: &mrID}, nil
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
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), nil, nil)

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)

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
			svc := NewExaminationService(repo, medRec, examTypeRepo, nil, &mockCheckupTransactor{})

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
			svc := NewExaminationService(repo, medRec, examTypeRepo, nil, &mockCheckupTransactor{})

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
			svc := NewExaminationService(repo, medRec, okExamTypeRepo(), nil, &mockCheckupTransactor{})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

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
	t.Run("computes status from ref_min/ref_max and overrides FE-supplied values", func(t *testing.T) {
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

		min1 := 1.0
		max10 := 10.0
		inputs := []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0", RefMin: &min1, RefMax: &max10}, // normal
			{Name: "RBC", InspectionValue: "0.5", RefMin: &min1, RefMax: &max10}, // low
			{Name: "PLT", InspectionValue: "11", RefMin: &min1, RefMax: &max10},  // high
			{Name: "Note", InspectionValue: "陰性"},                                // non-numeric → normal
		}

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, inputs)
		assert.NoError(t, err)
		assert.Len(t, saved, 4)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[0].Status)
		assert.False(t, capturedItems[0].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusLow, capturedItems[1].Status)
		assert.True(t, capturedItems[1].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusHigh, capturedItems[2].Status)
		assert.True(t, capturedItems[2].IsAbnormal)

		assert.Equal(t, model.ExaminationResultStatusNormal, capturedItems[3].Status)
		assert.False(t, capturedItems[3].IsAbnormal)

		// ExamID は service が強制設定する
		for _, it := range capturedItems {
			assert.Equal(t, uint64(10), it.ExamID)
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
		assert.True(t, apperrors.IsInvalidInput(err))
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
			logEntryTxFn: func(_ context.Context, _ *AuditLogInput) error {
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
			logEntryTxFn: func(_ context.Context, _ *AuditLogInput) error {
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
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, ExamTypeID: 50, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
				captured = items
				return items, 0, nil
			},
		}
		examTypeRepo := &mockExamTypeRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.ExaminationType, error) {
			return &model.ExaminationType{ID: id, Items: []model.ExamTypeField{{ID: 100}, {ID: 101}}}, nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, examTypeRepo, nil, &mockCheckupTransactor{})

		ownField := uint64(100)
		saved, err := svc.ReplaceItems(context.Background(), 1, 10, nil, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0", ExamTypeFieldID: &ownField},
		})
		assert.NoError(t, err)
		assert.Len(t, saved, 1)
		if assert.NotNil(t, captured[0].ExamTypeItemID) {
			assert.Equal(t, uint64(100), *captured[0].ExamTypeItemID)
		}
	})
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
