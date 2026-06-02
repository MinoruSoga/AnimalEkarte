package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	replaceItemsByExamIDFn func(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, error)
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

func (m *mockExaminationRepository) ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, error) {
	if m.replaceItemsByExamIDFn == nil {
		return items, nil
	}
	return m.replaceItemsByExamIDFn(ctx, clinicID, examID, items)
}

// mockMedicalRecordRepositoryForExam は MedicalRecordRepository のテスト用 stub（全メソッド実装）
type mockMedicalRecordRepositoryForExam struct{}

func (m *mockMedicalRecordRepositoryForExam) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}

func (m *mockMedicalRecordRepositoryForExam) Create(ctx context.Context, record *model.MedicalRecord) error {
	return nil
}

func (m *mockMedicalRecordRepositoryForExam) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) Delete(ctx context.Context, clinicID, id uint64) error {
	return nil
}

func (m *mockMedicalRecordRepositoryForExam) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicalRecordRepositoryForExam) CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicalRecordRepositoryForExam) CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error) {
	return 0, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*repository.OwnerVisitSummary, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnersByLTVRange(ctx context.Context, clinicID uint64, minLTV, maxLTV int64, excludeOwnerIDs []uint64) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindOwnersByPetAgeRange(ctx context.Context, clinicID uint64, minMonths, maxMonths int, excludeOwnerIDs []uint64) ([]uint64, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepositoryForExam) FindLastVisitDateByOwner(ctx context.Context, clinicID, ownerID uint64) (*time.Time, error) {
	return nil, nil
}

// mockAuditServiceForExam は AuditService のテスト用 stub（全メソッド実装）
type mockAuditServiceForExam struct{}

func (m *mockAuditServiceForExam) LogEntry(ctx context.Context, input *AuditLogInput) error {
	return nil
}

func (m *mockAuditServiceForExam) Log(ctx context.Context, audit *model.AuditLog) error {
	return nil
}

func (m *mockAuditServiceForExam) LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue any) error {
	return nil
}

func (m *mockAuditServiceForExam) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resourceType string, resourceID *uint64) error {
	return nil
}

func (m *mockAuditServiceForExam) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resourceType string, resourceID *uint64, metadata map[string]any) error {
	return nil
}

func (m *mockAuditServiceForExam) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID uint64, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	return nil
}

func (m *mockAuditServiceForExam) LogAddendumDelete(ctx context.Context, clinicID uint64, actorID *uint64, addendumID uint64, medicalRecordID uint64) error {
	return nil
}

func (m *mockAuditServiceForExam) LogVaccinationAdministration(ctx context.Context, clinicID uint64, actorID *uint64, action string, vaccinationID uint64, oldValue, newValue any) error {
	return nil
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
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

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
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

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
		name     string
		clinicID uint64
		input    *CreateExaminationInput
		repoErr  error
		wantErr  bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				createFn: func(_ context.Context, _ *model.Examination) error {
					return tt.repoErr
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

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
		name    string
		input   UpdateExaminationInput
		repoErr error
		wantErr bool
		wantNF  bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExaminationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					if tt.wantNF {
						return nil, tt.repoErr
					}
					return &model.Examination{ID: 1, Status: model.ExaminationStatusPending}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Examination{ID: 1}, nil
				},
			}
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

			exam, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exam)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
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
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

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
			svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

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
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, error) {
				capturedItems = items
				return items, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

		min1 := 1.0
		max10 := 10.0
		inputs := []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0", RefMin: &min1, RefMax: &max10}, // normal
			{Name: "RBC", InspectionValue: "0.5", RefMin: &min1, RefMax: &max10}, // low
			{Name: "PLT", InspectionValue: "11", RefMin: &min1, RefMax: &max10},  // high
			{Name: "Note", InspectionValue: "陰性"},                                // non-numeric → normal
		}

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, inputs)
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
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, error) {
				t.Fatal("repo should not be called when exam is confirmed")
				return nil, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

		_, err := svc.ReplaceItems(context.Background(), 1, 10, []UpsertExamItemInput{
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
		svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

		_, err := svc.ReplaceItems(context.Background(), 1, 999, []UpsertExamItemInput{})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("accepts empty items list (clears all)", func(t *testing.T) {
		called := false
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, error) {
				called = true
				assert.Empty(t, items)
				return []model.ExamResult{}, nil
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

		saved, err := svc.ReplaceItems(context.Background(), 1, 10, []UpsertExamItemInput{})
		assert.NoError(t, err)
		assert.Empty(t, saved)
		assert.True(t, called)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{ID: 10, Status: model.ExaminationStatusInProgress}, nil
			},
			replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, _ []model.ExamResult) ([]model.ExamResult, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewExaminationService(repo, &mockMedicalRecordRepositoryForExam{}, nil)

		_, err := svc.ReplaceItems(context.Background(), 1, 10, []UpsertExamItemInput{
			{Name: "WBC", InspectionValue: "5.0"},
		})
		assert.Error(t, err)
	})
}
