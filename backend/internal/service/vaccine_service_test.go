package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockVaccineRepository は VaccineRepository のテスト用モック実装
type mockVaccineRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	createFn                  func(ctx context.Context, vaccine *model.Vaccine) error
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	countUsageByVaccineIDFn   func(ctx context.Context, clinicID, vaccineID uint64) (int64, error)
	countChildrenByParentIDFn func(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

func (m *mockVaccineRepository) FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	return m.findAllFn(ctx, clinicID, species)
}

func (m *mockVaccineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockVaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	return m.createFn(ctx, vaccine)
}

func (m *mockVaccineRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockVaccineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockVaccineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return nil
}

func (m *mockVaccineRepository) CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error) {
	if m.countUsageByVaccineIDFn == nil {
		return 0, nil
	}
	return m.countUsageByVaccineIDFn(ctx, clinicID, vaccineID)
}

func (m *mockVaccineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	if m.countChildrenByParentIDFn == nil {
		return 0, nil
	}
	return m.countChildrenByParentIDFn(ctx, clinicID, parentID)
}

func TestVaccineService_List(t *testing.T) {
	dogSpecies := "dog"

	tests := []struct {
		name     string
		species  *string
		repoData []model.Vaccine
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name:    "returns all vaccines without species filter",
			species: nil,
			repoData: []model.Vaccine{
				{ID: 1, Name: "混合ワクチン（犬）"},
				{ID: 2, Name: "混合ワクチン（猫）"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "returns vaccines filtered by species",
			species: &dogSpecies,
			repoData: []model.Vaccine{
				{ID: 1, Name: "混合ワクチン（犬）"},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:     "returns empty list when no vaccines exist",
			species:  nil,
			repoData: []model.Vaccine{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			species:  nil,
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *string) ([]model.Vaccine, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			vaccines, err := svc.List(context.Background(), 1, tt.species)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vaccines, tt.wantLen)
			}
		})
	}
}

func TestVaccineService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoVaccine  *model.Vaccine
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns vaccine when found",
			id:           1,
			repoVaccine:  &model.Vaccine{ID: 1, Name: "混合ワクチン（犬）"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when vaccine does not exist",
			id:           999,
			repoVaccine:  nil,
			repoErr:      apperrors.WrapNotFound("vaccine", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoVaccine:  nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccine, error) {
					return tt.repoVaccine, tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			vaccine, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, vaccine)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoVaccine, vaccine)
			}
		})
	}
}

func TestVaccineService_Create(t *testing.T) {
	negativePrice := int64(-1)
	tests := []struct {
		name    string
		input   *CreateVaccineInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates vaccine successfully",
			input: &CreateVaccineInput{
				Name:     "新規ワクチン",
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when name is empty (BUG-396)",
			input: &CreateVaccineInput{
				Name: "",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when price is negative (BUG-380)",
			input: &CreateVaccineInput{
				Name:  "価格エラーワクチン",
				Price: &negativePrice,
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error when vaccine already exists",
			input: &CreateVaccineInput{
				Name: "重複ワクチン",
			},
			repoErr: apperrors.WrapAlreadyExists("vaccine", "重複ワクチン"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateVaccineInput{
				Name: "エラーワクチン",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				createFn: func(_ context.Context, _ *model.Vaccine) error {
					return tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			vaccine, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccine)
			}
		})
	}
}

func TestVaccineService_Update(t *testing.T) {
	name := "更新後ワクチン名"
	tests := []struct {
		name    string
		input   UpdateVaccineInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates vaccine successfully",
			input: UpdateVaccineInput{
				Name: &name,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateVaccineInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns not found error when vaccine does not exist",
			input: UpdateVaccineInput{
				Name: &name,
			},
			repoErr: apperrors.WrapNotFound("vaccine", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateVaccineInput{
				Name: &name,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccine, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Vaccine{ID: 1, ClinicID: 1}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccine, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Vaccine{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewVaccineService(repo)

			vaccine, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccine)
			}
		})
	}
}

func TestVaccineService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		childCount       int64
		countChildrenErr error
		usageCount       int64
		countUsageErr    error
		findByIDErr      error
		repoErr          error
		wantErr          bool
		wantNF           bool
		wantConflict     bool
	}{
		{
			name:             "deletes vaccine successfully when no children and no vaccinations use it",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          false,
		},
		{
			name:             "returns conflict error when vaccine has children (BUG-390)",
			id:               1,
			childCount:       2,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name:             "returns error when children count check fails (BUG-390)",
			id:               1,
			childCount:       0,
			countChildrenErr: errors.New("db error"),
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
		},
		{
			name:             "returns conflict error when vaccine is used in vaccination records",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       5,
			countUsageErr:    nil,
			repoErr:          nil,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name:             "returns error when usage count check fails",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    errors.New("db error"),
			repoErr:          nil,
			wantErr:          true,
		},
		{
			name:        "returns not found error when vaccine does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("vaccine", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:             "returns error on repository failure",
			id:               1,
			childCount:       0,
			countChildrenErr: nil,
			usageCount:       0,
			countUsageErr:    nil,
			repoErr:          errors.New("db error"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Vaccine{ID: id}, nil
				},
				countChildrenByParentIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.childCount, tt.countChildrenErr
				},
				countUsageByVaccineIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			err := svc.Delete(context.Background(), 1, tt.id)

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
