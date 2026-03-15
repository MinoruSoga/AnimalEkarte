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
	findAllFn  func(ctx context.Context, species *string) ([]model.Vaccine, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.Vaccine, error)
	createFn   func(ctx context.Context, vaccine *model.Vaccine) error
	updateFn   func(ctx context.Context, vaccine *model.Vaccine) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockVaccineRepository) FindAll(ctx context.Context, species *string) ([]model.Vaccine, error) {
	return m.findAllFn(ctx, species)
}

func (m *mockVaccineRepository) FindByID(ctx context.Context, id uint64) (*model.Vaccine, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockVaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	return m.createFn(ctx, vaccine)
}

func (m *mockVaccineRepository) Update(ctx context.Context, vaccine *model.Vaccine) error {
	return m.updateFn(ctx, vaccine)
}

func (m *mockVaccineRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockVaccineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return nil
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
				findAllFn: func(_ context.Context, _ *string) ([]model.Vaccine, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			vaccines, err := svc.List(context.Background(), tt.species)

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
				findByIDFn: func(_ context.Context, _ uint64) (*model.Vaccine, error) {
					return tt.repoVaccine, tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			vaccine, err := svc.GetByID(context.Background(), tt.id)

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
	tests := []struct {
		name    string
		vaccine *model.Vaccine
		repoErr error
		wantErr bool
	}{
		{
			name: "creates vaccine successfully",
			vaccine: &model.Vaccine{
				Name:     "新規ワクチン",
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when vaccine already exists",
			vaccine: &model.Vaccine{
				Name:     "重複ワクチン",
				ClinicID: 1,
			},
			repoErr: apperrors.WrapAlreadyExists("vaccine", "重複ワクチン"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			vaccine: &model.Vaccine{
				Name:     "エラーワクチン",
				ClinicID: 1,
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

			err := svc.Create(context.Background(), tt.vaccine)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVaccineService_Update(t *testing.T) {
	tests := []struct {
		name    string
		vaccine *model.Vaccine
		repoErr error
		wantErr bool
	}{
		{
			name: "updates vaccine successfully",
			vaccine: &model.Vaccine{
				ID:   1,
				Name: "更新後ワクチン名",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when vaccine does not exist",
			vaccine: &model.Vaccine{
				ID:   999,
				Name: "存在しないワクチン",
			},
			repoErr: apperrors.WrapNotFound("vaccine", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			vaccine: &model.Vaccine{
				ID:   1,
				Name: "エラーケース",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccineRepository{
				updateFn: func(_ context.Context, _ *model.Vaccine) error {
					return tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

			err := svc.Update(context.Background(), tt.vaccine)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVaccineService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes vaccine successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when vaccine does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("vaccine", "999"),
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
			repo := &mockVaccineRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewVaccineService(repo)

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
