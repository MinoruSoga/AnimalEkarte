package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation モック ----

type mockConsultationRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	findByIDFn                   func(ctx context.Context, id uint64) (*model.Consultation, error)
	createFn                     func(ctx context.Context, consultation *model.Consultation) error
	updateFieldsFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	deleteFn                     func(ctx context.Context, id uint64) error
	countUsageByConsultationIDFn func(ctx context.Context, consultationID uint64) (int64, error)
	reorderErr                   error
}

func (m *mockConsultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockConsultationRepository) FindByID(ctx context.Context, id uint64) (*model.Consultation, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockConsultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	return m.createFn(ctx, consultation)
}

func (m *mockConsultationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockConsultationRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockConsultationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockConsultationRepository) CountUsageByConsultationID(ctx context.Context, consultationID uint64) (int64, error) {
	if m.countUsageByConsultationIDFn == nil {
		return 0, nil
	}
	return m.countUsageByConsultationIDFn(ctx, consultationID)
}

// ---- Tests ----

func TestConsultationService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Consultation
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns consultations list",
			repoData: []model.Consultation{
				{ID: 1, ClinicID: 1, Name: "相談1", SortOrder: 1, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "相談2", SortOrder: 2, IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no consultations exist",
			repoData: []model.Consultation{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Consultation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			consultations, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, consultations, tt.wantLen)
			}
		})
	}
}

func TestConsultationService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoData     *model.Consultation
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns consultation when found",
			id:   1,
			repoData: &model.Consultation{
				ID:        1,
				ClinicID:  1,
				Name:      "相談1",
				SortOrder: 1,
				IsActive:  true,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when consultation does not exist",
			id:           999,
			repoData:     nil,
			repoErr:      apperrors.WrapNotFound("consultation", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoData:     nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Consultation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, consultation)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoData, consultation)
			}
		})
	}
}

func TestConsultationService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *model.Consultation
		repoErr error
		wantErr bool
	}{
		{
			name: "creates consultation successfully",
			input: &model.Consultation{
				ClinicID:  1,
				Name:      "新規相談",
				SortOrder: 3,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when consultation already exists",
			input: &model.Consultation{
				ClinicID: 1,
				Name:     "既存相談",
			},
			repoErr: apperrors.WrapAlreadyExists("consultation", "既存相談"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &model.Consultation{
				ClinicID: 1,
				Name:     "エラー相談",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				createFn: func(_ context.Context, _ *model.Consultation) error {
					return tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsultationService_Update(t *testing.T) {
	name := "更新後相談"
	isActive := true
	tests := []struct {
		name    string
		input   UpdateConsultationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates consultation successfully",
			input: UpdateConsultationInput{
				Name:     &name,
				IsActive: &isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateConsultationInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when consultation does not exist",
			input: UpdateConsultationInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("consultation", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateConsultationInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Consultation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Consultation{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewConsultationService(repo)

			consultation, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, consultation)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, consultation)
			}
		})
	}
}

func TestConsultationService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		usageCount    int64
		countUsageErr error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes consultation successfully when no medical records use it",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when consultation is used in medical records",
			id:            1,
			usageCount:    2,
			countUsageErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when usage count check fails",
			id:            1,
			usageCount:    0,
			countUsageErr: errors.New("db error"),
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:          "returns not found error when consultation does not exist",
			id:            999,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       apperrors.WrapNotFound("consultation", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			usageCount:    0,
			countUsageErr: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
				countUsageByConsultationIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			err := svc.Delete(context.Background(), tt.id)

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

func TestConsultationService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{2, 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when ids is empty",
			ids:     []uint64{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{reorderErr: tt.repoErr}
			svc := NewConsultationService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
