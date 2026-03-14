package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockInsuranceRepository は InsuranceRepository のテスト用モック実装
type mockInsuranceRepository struct {
	findAllFn  func(ctx context.Context) ([]model.Insurance, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.Insurance, error)
	createFn   func(ctx context.Context, insurance *model.Insurance) error
	updateFn   func(ctx context.Context, insurance *model.Insurance) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockInsuranceRepository) FindAll(ctx context.Context) ([]model.Insurance, error) {
	return m.findAllFn(ctx)
}

func (m *mockInsuranceRepository) FindByID(ctx context.Context, id uint64) (*model.Insurance, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockInsuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	return m.createFn(ctx, insurance)
}

func (m *mockInsuranceRepository) Update(ctx context.Context, insurance *model.Insurance) error {
	return m.updateFn(ctx, insurance)
}

func (m *mockInsuranceRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func TestInsuranceService_List(t *testing.T) {
	tests := []struct {
		name           string
		repoInsurances []model.Insurance
		repoErr        error
		wantLen        int
		wantErr        bool
	}{
		{
			name: "returns insurance list",
			repoInsurances: []model.Insurance{
				{ID: 1, Name: "アニコム損保", IsActive: true},
				{ID: 2, Name: "アイペット損保", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:           "returns empty list when no insurances exist",
			repoInsurances: []model.Insurance{},
			repoErr:        nil,
			wantLen:        0,
			wantErr:        false,
		},
		{
			name:           "propagates repository error",
			repoInsurances: nil,
			repoErr:        errors.New("db connection error"),
			wantLen:        0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findAllFn: func(_ context.Context) ([]model.Insurance, error) {
					return tt.repoInsurances, tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			insurances, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, insurances, tt.wantLen)
			}
		})
	}
}

func TestInsuranceService_GetByID(t *testing.T) {
	coverageRate := 70
	tests := []struct {
		name          string
		id            uint64
		repoInsurance *model.Insurance
		repoErr       error
		wantErr       bool
		wantNF        bool
	}{
		{
			name: "returns insurance when found",
			id:   1,
			repoInsurance: &model.Insurance{
				ID:           1,
				Name:         "アニコム損保",
				IsActive:     true,
				CoverageRate: coverageRate,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:          "returns not found error when insurance does not exist",
			id:            999,
			repoInsurance: nil,
			repoErr:       apperrors.WrapNotFound("insurance", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		{
			name:          "returns error on repository failure",
			id:            1,
			repoInsurance: nil,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Insurance, error) {
					return tt.repoInsurance, tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			insurance, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoInsurance, insurance)
			}
		})
	}
}

func TestInsuranceService_Create(t *testing.T) {
	coverageRate := 70
	tests := []struct {
		name      string
		insurance *model.Insurance
		repoErr   error
		wantErr   bool
	}{
		{
			name: "creates insurance successfully",
			insurance: &model.Insurance{
				Name:         "新規保険",
				IsActive:     true,
				CoverageRate: coverageRate,
				ContactPhone: "03-1234-5678",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when insurance already exists",
			insurance: &model.Insurance{
				Name:     "既存保険",
				IsActive: true,
			},
			repoErr: apperrors.WrapAlreadyExists("insurance", "既存保険"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			insurance: &model.Insurance{
				Name:     "エラー保険",
				IsActive: true,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				createFn: func(_ context.Context, _ *model.Insurance) error {
					return tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			err := svc.Create(context.Background(), tt.insurance)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsuranceService_Update(t *testing.T) {
	coverageRate := 80
	tests := []struct {
		name      string
		insurance *model.Insurance
		repoErr   error
		wantErr   bool
	}{
		{
			name: "updates insurance successfully",
			insurance: &model.Insurance{
				ID:           1,
				Name:         "更新後保険",
				IsActive:     true,
				CoverageRate: coverageRate,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error on repository failure",
			insurance: &model.Insurance{
				ID:   999,
				Name: "エラー保険",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				updateFn: func(_ context.Context, _ *model.Insurance) error {
					return tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

			err := svc.Update(context.Background(), tt.insurance)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsuranceService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes insurance successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when insurance does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("insurance", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsuranceRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewInsuranceService(repo)

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
