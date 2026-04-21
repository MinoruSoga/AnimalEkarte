package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- HospitalizationPlan モック ----

type mockHospitalizationPlanRepository struct {
	findAllFn                    func(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error)
	findByIDFn                   func(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error)
	createFn                     func(ctx context.Context, plan *model.HospitalizationPlan) error
	updateFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	reorderFn                    func(ctx context.Context, clinicID uint64, ids []uint64) error
	countCarePlanItemsByPlanIDFn func(ctx context.Context, clinicID, planID uint64) (int64, error)
}

func (m *mockHospitalizationPlanRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockHospitalizationPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.HospitalizationPlan{ID: id, ClinicID: clinicID}, nil
}

func (m *mockHospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockHospitalizationPlanRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.HospitalizationPlan, error) {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockHospitalizationPlanRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationPlanRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockHospitalizationPlanRepository) CountCarePlanItemsByPlanID(ctx context.Context, clinicID, planID uint64) (int64, error) {
	if m.countCarePlanItemsByPlanIDFn != nil {
		return m.countCarePlanItemsByPlanIDFn(ctx, clinicID, planID)
	}
	return 0, nil
}

// ---- Tests ----

func TestHospitalizationPlanService_List(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		repoPlan []model.HospitalizationPlan
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns hospitalization plans for clinic",
			clinicID: 1,
			repoPlan: []model.HospitalizationPlan{
				{ID: 1, ClinicID: 1, Name: "Plan 1", IsActive: true},
				{ID: 2, ClinicID: 1, Name: "Plan 2", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no plans exist",
			clinicID: 999,
			repoPlan: []model.HospitalizationPlan{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			clinicID: 1,
			repoPlan: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.HospitalizationPlan, error) {
					return tt.repoPlan, tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plans, err := svc.List(context.Background(), tt.clinicID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, plans, tt.wantLen)
			}
		})
	}
}

func TestHospitalizationPlanService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		repoPlan *model.HospitalizationPlan
		repoErr  error
		wantErr  bool
	}{
		{
			name: "returns plan when found",
			id:   1,
			repoPlan: &model.HospitalizationPlan{
				ID:       1,
				ClinicID: 1,
				Name:     "Hospitalization Plan",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when plan not found",
			id:       999,
			repoPlan: nil,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.HospitalizationPlan, error) {
					return tt.repoPlan, tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.repoPlan, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateHospitalizationPlanInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates hospitalization plan successfully",
			input: &CreateHospitalizationPlanInput{
				Name:     "New Plan",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when repository fails",
			input: &CreateHospitalizationPlanInput{
				Name: "Failed Plan",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				createFn: func(_ context.Context, _ *model.HospitalizationPlan) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Update(t *testing.T) {
	name := "Updated Plan"
	isActive := false
	tests := []struct {
		name    string
		input   UpdateHospitalizationPlanInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates hospitalization plan successfully",
			input: UpdateHospitalizationPlanInput{
				Name:     &name,
				IsActive: &isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateHospitalizationPlanInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when repository fails",
			input: UpdateHospitalizationPlanInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.HospitalizationPlan, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.HospitalizationPlan{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewHospitalizationPlanService(repo)

			plan, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

func TestHospitalizationPlanService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		carePlanCount int64
		countErr      error
		repoErr       error
		wantErr       bool
		wantConflict  bool
	}{
		{
			name:    "deletes hospitalization plan successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when plan not found",
			id:      999,
			repoErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:          "使用中の入院プランは削除できない",
			id:            2,
			carePlanCount: 3,
			wantErr:       true,
			wantConflict:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationPlanRepository{
				countCarePlanItemsByPlanIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.carePlanCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHospitalizationPlanService_Reorder(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		ids      []uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "reorders plans successfully",
			clinicID: 1,
			ids:      []uint64{2, 1, 3},
			repoErr:  nil,
			wantErr:  false,
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
			repo := &mockHospitalizationPlanRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := NewHospitalizationPlanService(repo)

			err := svc.Reorder(context.Background(), tt.clinicID, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
