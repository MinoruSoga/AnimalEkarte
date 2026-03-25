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
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.Consultation, error)
	createFn   func(ctx context.Context, consultation *model.Consultation) error
	updateFn   func(ctx context.Context, consultation *model.Consultation) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderErr error
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

func (m *mockConsultationRepository) Update(ctx context.Context, consultation *model.Consultation) error {
	return m.updateFn(ctx, consultation)
}

func (m *mockConsultationRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockConsultationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
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
	tests := []struct {
		name    string
		input   *model.Consultation
		repoErr error
		wantErr bool
	}{
		{
			name: "updates consultation successfully",
			input: &model.Consultation{
				ID:        1,
				ClinicID:  1,
				Name:      "更新後相談",
				SortOrder: 2,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when consultation does not exist",
			input: &model.Consultation{
				ID:       999,
				ClinicID: 1,
				Name:     "存在しない相談",
			},
			repoErr: apperrors.WrapNotFound("consultation", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &model.Consultation{
				ID:       1,
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
				updateFn: func(_ context.Context, _ *model.Consultation) error {
					return tt.repoErr
				},
			}
			svc := NewConsultationService(repo)

			err := svc.Update(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsultationService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes consultation successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when consultation does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("consultation", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockConsultationRepository{
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
