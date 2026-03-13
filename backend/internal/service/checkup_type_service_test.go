package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockCheckupTypeRepository は CheckupTypeRepository のテスト用モック実装
type mockCheckupTypeRepository struct {
	findAllFn  func(ctx context.Context) ([]model.CheckupType, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.CheckupType, error)
	createFn   func(ctx context.Context, checkupType *model.CheckupType) error
	updateFn   func(ctx context.Context, checkupType *model.CheckupType) error
	deleteFn   func(ctx context.Context, id uint64) error
}

func (m *mockCheckupTypeRepository) FindAll(ctx context.Context) ([]model.CheckupType, error) {
	return m.findAllFn(ctx)
}

func (m *mockCheckupTypeRepository) FindByID(ctx context.Context, id uint64) (*model.CheckupType, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockCheckupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return m.createFn(ctx, checkupType)
}

func (m *mockCheckupTypeRepository) Update(ctx context.Context, checkupType *model.CheckupType) error {
	return m.updateFn(ctx, checkupType)
}

func (m *mockCheckupTypeRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func TestCheckupTypeService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.CheckupType
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns checkup type list",
			repoData: []model.CheckupType{
				{ID: 1, Name: "定期健診"},
				{ID: 2, Name: "ワクチン前健診"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no checkup types exist",
			repoData: []model.CheckupType{},
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
			repo := &mockCheckupTypeRepository{
				findAllFn: func(_ context.Context) ([]model.CheckupType, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			checkupTypes, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, checkupTypes, tt.wantLen)
			}
		})
	}
}

func TestCheckupTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint64
		repoCheckupType *model.CheckupType
		repoErr         error
		wantErr         bool
		wantNotFound    bool
	}{
		{
			name:            "returns checkup type when found",
			id:              1,
			repoCheckupType: &model.CheckupType{ID: 1, Name: "定期健診"},
			repoErr:         nil,
			wantErr:         false,
			wantNotFound:    false,
		},
		{
			name:            "returns not found error when checkup type does not exist",
			id:              999,
			repoCheckupType: nil,
			repoErr:         apperrors.WrapNotFound("checkup_type", "999"),
			wantErr:         true,
			wantNotFound:    true,
		},
		{
			name:            "returns error on repository failure",
			id:              1,
			repoCheckupType: nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
			wantNotFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.CheckupType, error) {
					return tt.repoCheckupType, tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			checkupType, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, checkupType)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCheckupType, checkupType)
			}
		})
	}
}

func TestCheckupTypeService_Create(t *testing.T) {
	tests := []struct {
		name        string
		checkupType *model.CheckupType
		repoErr     error
		wantErr     bool
	}{
		{
			name: "creates checkup type successfully",
			checkupType: &model.CheckupType{
				Name:     "新規健診種別",
				ClinicID: 1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when checkup type already exists",
			checkupType: &model.CheckupType{
				Name:     "重複健診種別",
				ClinicID: 1,
			},
			repoErr: apperrors.WrapAlreadyExists("checkup_type", "重複健診種別"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			checkupType: &model.CheckupType{
				Name:     "エラー健診種別",
				ClinicID: 1,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				createFn: func(_ context.Context, _ *model.CheckupType) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			err := svc.Create(context.Background(), tt.checkupType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckupTypeService_Update(t *testing.T) {
	tests := []struct {
		name        string
		checkupType *model.CheckupType
		repoErr     error
		wantErr     bool
	}{
		{
			name: "updates checkup type successfully",
			checkupType: &model.CheckupType{
				ID:   1,
				Name: "更新後健診種別",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when checkup type does not exist",
			checkupType: &model.CheckupType{
				ID:   999,
				Name: "存在しない健診種別",
			},
			repoErr: apperrors.WrapNotFound("checkup_type", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			checkupType: &model.CheckupType{
				ID:   1,
				Name: "エラーケース",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeRepository{
				updateFn: func(_ context.Context, _ *model.CheckupType) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

			err := svc.Update(context.Background(), tt.checkupType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckupTypeService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes checkup type successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when checkup type does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("checkup_type", "999"),
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
			repo := &mockCheckupTypeRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewCheckupTypeService(repo)

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
