package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMedicineRepository は MedicineRepository のテスト用モック実装
type mockMedicineRepository struct {
	findAllFn  func(ctx context.Context) ([]model.Medicine, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.Medicine, error)
	createFn   func(ctx context.Context, medicine *model.Medicine) error
	updateFn   func(ctx context.Context, medicine *model.Medicine) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockMedicineRepository) FindAll(ctx context.Context) ([]model.Medicine, error) {
	return m.findAllFn(ctx)
}

func (m *mockMedicineRepository) FindByID(ctx context.Context, id uint64) (*model.Medicine, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockMedicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	return m.createFn(ctx, medicine)
}

func (m *mockMedicineRepository) Update(ctx context.Context, medicine *model.Medicine) error {
	return m.updateFn(ctx, medicine)
}

func (m *mockMedicineRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func TestMedicineService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Medicine
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns medicine list",
			repoData: []model.Medicine{
				{ID: 1, Name: "アモキシシリン"},
				{ID: 2, Name: "メトロニダゾール"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no medicines exist",
			repoData: []model.Medicine{},
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
			repo := &mockMedicineRepository{
				findAllFn: func(_ context.Context) ([]model.Medicine, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewMedicineService(repo)

			medicines, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, medicines, tt.wantLen)
			}
		})
	}
}

func TestMedicineService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoMedicine *model.Medicine
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns medicine when found",
			id:           1,
			repoMedicine: &model.Medicine{ID: 1, Name: "アモキシシリン"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when medicine does not exist",
			id:           999,
			repoMedicine: nil,
			repoErr:      apperrors.WrapNotFound("medicine", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoMedicine: nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Medicine, error) {
					return tt.repoMedicine, tt.repoErr
				},
			}
			svc := NewMedicineService(repo)

			medicine, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, medicine)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoMedicine, medicine)
			}
		})
	}
}

func TestMedicineService_Create(t *testing.T) {
	tests := []struct {
		name     string
		medicine *model.Medicine
		repoErr  error
		wantErr  bool
	}{
		{
			name: "creates medicine successfully",
			medicine: &model.Medicine{
				Name:     "新規薬剤",
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when medicine already exists",
			medicine: &model.Medicine{
				Name:     "重複薬剤",
				ClinicID: 1,
			},
			repoErr: apperrors.WrapAlreadyExists("medicine", "重複薬剤"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			medicine: &model.Medicine{
				Name:     "エラー薬剤",
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				createFn: func(_ context.Context, _ *model.Medicine) error {
					return tt.repoErr
				},
			}
			svc := NewMedicineService(repo)

			err := svc.Create(context.Background(), tt.medicine)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicineService_Update(t *testing.T) {
	tests := []struct {
		name     string
		medicine *model.Medicine
		repoErr  error
		wantErr  bool
	}{
		{
			name: "updates medicine successfully",
			medicine: &model.Medicine{
				ID:   1,
				Name: "更新後薬剤名",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when medicine does not exist",
			medicine: &model.Medicine{
				ID:   999,
				Name: "存在しない薬剤",
			},
			repoErr: apperrors.WrapNotFound("medicine", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			medicine: &model.Medicine{
				ID:   1,
				Name: "エラーケース",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicineRepository{
				updateFn: func(_ context.Context, _ *model.Medicine) error {
					return tt.repoErr
				},
			}
			svc := NewMedicineService(repo)

			err := svc.Update(context.Background(), tt.medicine)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicineService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes medicine successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when medicine does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("medicine", "999"),
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
			repo := &mockMedicineRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewMedicineService(repo)

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
