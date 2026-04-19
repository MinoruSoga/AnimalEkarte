package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Occupation モック ----

type mockOccupationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	createFn       func(ctx context.Context, occupation *model.Occupation) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	reorderErr     error
}

func (m *mockOccupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockOccupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockOccupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	return m.createFn(ctx, occupation)
}

func (m *mockOccupationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockOccupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockOccupationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockOccupationRepository) CountStaffsByOccupationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// ---- Tests ----

func TestOccupationService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Occupation
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns occupations list",
			repoData: []model.Occupation{
				{ID: 1, ClinicID: 1, Name: "獣医師", SortOrder: 1, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "看護師", SortOrder: 2, IsActive: true},
				{ID: 3, ClinicID: 1, Name: "受付", SortOrder: 3, IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:     "returns empty list when no occupations exist",
			repoData: []model.Occupation{},
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
			repo := &mockOccupationRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Occupation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occupations, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, occupations, tt.wantLen)
			}
		})
	}
}

func TestOccupationService_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		id             uint64
		repoOccupation *model.Occupation
		repoErr        error
		wantErr        bool
		wantNotFound   bool
	}{
		{
			name: "returns occupation when found",
			id:   1,
			repoOccupation: &model.Occupation{
				ID:          1,
				ClinicID:    1,
				Name:        "獣医師",
				Description: "診療を行う医師",
				SortOrder:   1,
				IsActive:    true,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:           "returns not found error when occupation does not exist",
			id:             999,
			repoOccupation: nil,
			repoErr:        apperrors.WrapNotFound("occupation", "999"),
			wantErr:        true,
			wantNotFound:   true,
		},
		{
			name:           "returns error on repository failure",
			id:             1,
			repoOccupation: nil,
			repoErr:        errors.New("db error"),
			wantErr:        true,
			wantNotFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
					return tt.repoOccupation, tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occupation, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, occupation)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOccupation, occupation)
			}
		})
	}
}

func TestOccupationService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateOccupationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates occupation successfully",
			input: &CreateOccupationInput{
				Name:        "新規職種",
				Description: "新しい職種の説明",
				SortOrder:   4,
				IsActive:    true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when occupation already exists",
			input: &CreateOccupationInput{
				Name: "獣医師",
			},
			repoErr: apperrors.WrapAlreadyExists("occupation", "獣医師"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateOccupationInput{
				Name: "エラー職種",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				createFn: func(_ context.Context, _ *model.Occupation) error {
					return tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occ, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, occ)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, occ)
			}
		})
	}
}

func TestOccupationService_Update(t *testing.T) {
	updateName := "更新後職種名"
	updateSortOrder := 2
	anotherName := "更新職種"
	errorName := "エラー職種"

	tests := []struct {
		name           string
		clinicID       uint64
		id             uint64
		input          *UpdateOccupationInput
		repoOccupation *model.Occupation
		repoErr        error
		wantErr        bool
	}{
		{
			name:     "updates occupation successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateOccupationInput{
				Name:      &updateName,
				SortOrder: &updateSortOrder,
			},
			repoOccupation: &model.Occupation{
				ID:        1,
				ClinicID:  1,
				Name:      "更新後職種名",
				SortOrder: 2,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateOccupationInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns not found error when occupation does not exist",
			clinicID: 1,
			id:       999,
			input: &UpdateOccupationInput{
				Name: &anotherName,
			},
			repoErr: apperrors.WrapNotFound("occupation", "999"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: &UpdateOccupationInput{
				Name: &errorName,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Occupation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.repoOccupation, nil
				},
			}
			svc := NewOccupationService(repo)

			occupation, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOccupation, occupation)
			}
		})
	}
}

func TestOccupationService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes occupation successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when occupation does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("occupation", "999"),
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
			repo := &mockOccupationRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

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

func TestOccupationService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{2, 3, 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "reorders single item",
			ids:     []uint64{1},
			repoErr: nil,
			wantErr: false,
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
			repo := &mockOccupationRepository{reorderErr: tt.repoErr}
			svc := NewOccupationService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
