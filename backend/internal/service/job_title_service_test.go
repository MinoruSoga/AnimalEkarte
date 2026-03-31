package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- JobTitle モック ----

type mockJobTitleRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.JobTitle, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.JobTitle, error)
	createFn   func(ctx context.Context, jobTitle *model.JobTitle) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderErr error
}

func (m *mockJobTitleRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.JobTitle, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockJobTitleRepository) FindByID(ctx context.Context, id uint64) (*model.JobTitle, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockJobTitleRepository) Create(ctx context.Context, jobTitle *model.JobTitle) error {
	return m.createFn(ctx, jobTitle)
}

func (m *mockJobTitleRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockJobTitleRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockJobTitleRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

// ---- Tests ----

func TestJobTitleService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.JobTitle
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns job titles list",
			repoData: []model.JobTitle{
				{ID: 1, ClinicID: 1, Name: "獣医師", SortOrder: 1, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "看護師", SortOrder: 2, IsActive: true},
				{ID: 3, ClinicID: 1, Name: "受付", SortOrder: 3, IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:     "returns empty list when no job titles exist",
			repoData: []model.JobTitle{},
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
			repo := &mockJobTitleRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.JobTitle, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewJobTitleService(repo)

			jobTitles, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, jobTitles, tt.wantLen)
			}
		})
	}
}

func TestJobTitleService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoJobTitle *model.JobTitle
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name: "returns job title when found",
			id:   1,
			repoJobTitle: &model.JobTitle{
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
			name:         "returns not found error when job title does not exist",
			id:           999,
			repoJobTitle: nil,
			repoErr:      apperrors.WrapNotFound("job_title", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoJobTitle: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobTitleRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.JobTitle, error) {
					return tt.repoJobTitle, tt.repoErr
				},
			}
			svc := NewJobTitleService(repo)

			jobTitle, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, jobTitle)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoJobTitle, jobTitle)
			}
		})
	}
}

func TestJobTitleService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *model.JobTitle
		repoErr error
		wantErr bool
	}{
		{
			name: "creates job title successfully",
			input: &model.JobTitle{
				ClinicID:    1,
				Name:        "新規職種",
				Description: "新しい職種の説明",
				SortOrder:   4,
				IsActive:    true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when job title already exists",
			input: &model.JobTitle{
				ClinicID: 1,
				Name:     "獣医師",
			},
			repoErr: apperrors.WrapAlreadyExists("job_title", "獣医師"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &model.JobTitle{
				ClinicID: 1,
				Name:     "エラー職種",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobTitleRepository{
				createFn: func(_ context.Context, _ *model.JobTitle) error {
					return tt.repoErr
				},
			}
			svc := NewJobTitleService(repo)

			err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJobTitleService_Update(t *testing.T) {
	updateName := "更新後職種名"
	updateSortOrder := 2
	anotherName := "更新職種"
	errorName := "エラー職種"

	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		input        *UpdateJobTitleInput
		repoJobTitle *model.JobTitle
		repoErr      error
		wantErr      bool
	}{
		{
			name:     "updates job title successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateJobTitleInput{
				Name:      &updateName,
				SortOrder: &updateSortOrder,
			},
			repoJobTitle: &model.JobTitle{
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
			input:    &UpdateJobTitleInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns not found error when job title does not exist",
			clinicID: 1,
			id:       999,
			input: &UpdateJobTitleInput{
				Name: &anotherName,
			},
			repoErr: apperrors.WrapNotFound("job_title", "999"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: &UpdateJobTitleInput{
				Name: &errorName,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobTitleRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _ uint64) (*model.JobTitle, error) {
					if tt.repoErr != nil && apperrors.IsNotFound(tt.repoErr) {
						return nil, tt.repoErr
					}
					return tt.repoJobTitle, nil
				},
			}
			svc := NewJobTitleService(repo)

			jobTitle, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoJobTitle, jobTitle)
			}
		})
	}
}

func TestJobTitleService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes job title successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when job title does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("job_title", "999"),
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
			repo := &mockJobTitleRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewJobTitleService(repo)

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

func TestJobTitleService_Reorder(t *testing.T) {
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
			repo := &mockJobTitleRepository{reorderErr: tt.repoErr}
			svc := NewJobTitleService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
