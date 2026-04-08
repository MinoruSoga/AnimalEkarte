package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockExaminationRepository は ExaminationRepository のテスト用モック実装
type mockExaminationRepository struct {
	findAllFn            func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	findByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	createFn             func(ctx context.Context, exam *model.Examination) error
	updateFieldsFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	deleteFn             func(ctx context.Context, clinicID, id uint64) error
	countItemsByExamIDFn func(ctx context.Context, clinicID, examID uint64) (int64, error)
}

func (m *mockExaminationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockExaminationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockExaminationRepository) Create(ctx context.Context, exam *model.Examination) error {
	return m.createFn(ctx, exam)
}

func (m *mockExaminationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
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
			svc := NewExaminationService(repo)

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
			svc := NewExaminationService(repo)

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
		name    string
		exam    *model.Examination
		repoErr error
		wantErr bool
	}{
		{
			name: "creates exam successfully",
			exam: &model.Examination{
				MedicalRecordID: ptrUint64(5),
				ExamTypeID:      1,
				Date:            now,
				Status:          model.ExaminationStatusPending,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error on repository failure",
			exam: &model.Examination{
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
			svc := NewExaminationService(repo)

			err := svc.Create(context.Background(), tt.exam)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
			svc := NewExaminationService(repo)

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
			svc := NewExaminationService(repo)

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
