package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure モック ----

type mockProcedureRepository struct {
	findAllFn      func(ctx context.Context) ([]model.Procedure, error)
	findByIDFn     func(ctx context.Context, id uint64) (*model.Procedure, error)
	createFn       func(ctx context.Context, procedure *model.Procedure) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	deleteFn       func(ctx context.Context, id uint64) error
	reorderFn      func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockProcedureRepository) FindAll(ctx context.Context) ([]model.Procedure, error) {
	return m.findAllFn(ctx)
}

func (m *mockProcedureRepository) FindByID(ctx context.Context, id uint64) (*model.Procedure, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockProcedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	return m.createFn(ctx, procedure)
}

func (m *mockProcedureRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Procedure{ID: id}, nil
}

func (m *mockProcedureRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockProcedureRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- Tests ----

func TestProcedureService_List(t *testing.T) {
	tests := []struct {
		name           string
		repoProcedures []model.Procedure
		repoErr        error
		wantLen        int
		wantErr        bool
	}{
		{
			name: "returns all procedures",
			repoProcedures: []model.Procedure{
				{ID: 1, Name: "Surgery A", IsActive: true},
				{ID: 2, Name: "Surgery B", IsActive: true},
				{ID: 3, Name: "Examination", IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:           "returns empty list when no procedures exist",
			repoProcedures: []model.Procedure{},
			repoErr:        nil,
			wantLen:        0,
			wantErr:        false,
		},
		{
			name:           "propagates repository error",
			repoProcedures: nil,
			repoErr:        errors.New("db error"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findAllFn: func(_ context.Context) ([]model.Procedure, error) {
					return tt.repoProcedures, tt.repoErr
				},
			}
			svc := NewProcedureService(repo)

			procedures, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, procedures, tt.wantLen)
			}
		})
	}
}

func TestProcedureService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		repoProcedure *model.Procedure
		repoErr       error
		wantErr       bool
	}{
		{
			name: "returns procedure when found",
			id:   1,
			repoProcedure: &model.Procedure{
				ID:       1,
				Name:     "Surgical Procedure",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:          "returns error when procedure not found",
			id:            999,
			repoProcedure: nil,
			repoErr:       errors.New("not found"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Procedure, error) {
					return tt.repoProcedure, tt.repoErr
				},
			}
			svc := NewProcedureService(repo)

			procedure, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, procedure)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, procedure)
				assert.Equal(t, tt.repoProcedure, procedure)
			}
		})
	}
}

func TestProcedureService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *model.Procedure
		repoErr error
		wantErr bool
	}{
		{
			name: "creates procedure successfully",
			input: &model.Procedure{
				Name:     "New Procedure",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &model.Procedure{
				Name: "Failed Procedure",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				createFn: func(_ context.Context, _ *model.Procedure) error {
					return tt.repoErr
				},
			}
			svc := NewProcedureService(repo)

			err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcedureService_Update(t *testing.T) {
	name := "Updated Procedure"
	isActive := false
	tests := []struct {
		name    string
		input   UpdateProcedureInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates procedure successfully",
			input: UpdateProcedureInput{
				Name:     &name,
				IsActive: &isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateProcedureInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when procedure does not exist",
			input: UpdateProcedureInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("procedure", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error when repository fails",
			input: UpdateProcedureInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Procedure, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Procedure{ID: 1}, nil
				},
			}
			svc := NewProcedureService(repo)

			procedure, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, procedure)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, procedure)
			}
		})
	}
}

func TestProcedureService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "deletes procedure successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when procedure not found",
			id:      999,
			repoErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewProcedureService(repo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcedureService_Reorder(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		ids      []uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "reorders procedures successfully",
			clinicID: 1,
			ids:      []uint64{3, 1, 2},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when ids is empty",
			clinicID: 1,
			ids:      []uint64{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error when repository fails",
			clinicID: 1,
			ids:      []uint64{1, 2},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProcedureRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewProcedureService(repo)

			err := svc.Reorder(context.Background(), tt.clinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				// Validate empty ids error is about invalid input
				if len(tt.ids) == 0 {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
