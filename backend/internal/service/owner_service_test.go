package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOwnerRepository は OwnerRepository のテスト用モック実装
type mockOwnerRepository struct {
	findAllFn        func(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	findByIDFn       func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	createWithPetsFn func(ctx context.Context, owner *model.Owner, pets []model.Pet) error
	updateFn         func(ctx context.Context, owner *model.Owner) error
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockOwnerRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit, search)
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockOwnerRepository) CreateWithPets(ctx context.Context, owner *model.Owner, pets []model.Pet) error {
	if m.createWithPetsFn != nil {
		return m.createWithPetsFn(ctx, owner, pets)
	}
	return nil
}

func (m *mockOwnerRepository) Update(ctx context.Context, owner *model.Owner) error {
	return m.updateFn(ctx, owner)
}

func (m *mockOwnerRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func TestOwnerService_List(t *testing.T) {
	tests := []struct {
		name       string
		clinicID   uint64
		page       int
		limit      int
		search     string
		repoOwners []model.Owner
		repoTotal  int64
		repoErr    error
		wantLen    int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name:     "returns owner list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			search:   "",
			repoOwners: []model.Owner{
				{ID: 1, ClinicID: 1, OwnerName: "山田 太郎"},
				{ID: 2, ClinicID: 1, OwnerName: "鈴木 花子"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no owners exist",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "",
			repoOwners: []model.Owner{},
			repoTotal:  0,
			repoErr:    nil,
			wantLen:    0,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name:       "filters by search keyword",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "山田",
			repoOwners: []model.Owner{{ID: 1, ClinicID: 1, OwnerName: "山田 太郎"}},
			repoTotal:  1,
			repoErr:    nil,
			wantLen:    1,
			wantTotal:  1,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			clinicID:   1,
			page:       1,
			limit:      20,
			search:     "",
			repoOwners: nil,
			repoTotal:  0,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantTotal:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findAllFn: func(_ context.Context, _ uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return tt.repoOwners, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewOwnerService(repo)

			owners, total, err := svc.List(context.Background(), tt.clinicID, tt.page, tt.limit, tt.search)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, owners, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestOwnerService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		clinicID  uint64
		id        uint64
		repoOwner *model.Owner
		repoErr   error
		wantOwner *model.Owner
		wantErr   error
	}{
		{
			name:      "returns owner when found",
			clinicID:  1,
			id:        10,
			repoOwner: &model.Owner{ID: 10, ClinicID: 1, OwnerName: "山田 太郎"},
			repoErr:   nil,
			wantOwner: &model.Owner{ID: 10, ClinicID: 1, OwnerName: "山田 太郎"},
			wantErr:   nil,
		},
		{
			name:      "returns not found error when owner does not exist",
			clinicID:  1,
			id:        999,
			repoOwner: nil,
			repoErr:   apperrors.WrapNotFound("owner", "999"),
			wantOwner: nil,
			wantErr:   apperrors.ErrNotFound,
		},
		{
			name:      "returns error on repository failure",
			clinicID:  1,
			id:        10,
			repoOwner: nil,
			repoErr:   errors.New("db error"),
			wantOwner: nil,
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return tt.repoOwner, tt.repoErr
				},
			}
			svc := NewOwnerService(repo)

			owner, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
			}
		})
	}
}

func TestOwnerService_GetByID_NotFound(t *testing.T) {
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "999")
		},
	}
	svc := NewOwnerService(repo)

	owner, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, owner)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestOwnerService_Update(t *testing.T) {
	tests := []struct {
		name    string
		owner   *model.Owner
		repoErr error
		wantErr bool
	}{
		{
			name: "updates owner successfully",
			owner: &model.Owner{
				ID:        1,
				ClinicID:  1,
				OwnerName: "更新後 氏名",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns not found error when owner does not exist",
			owner: &model.Owner{
				ID:        999,
				ClinicID:  1,
				OwnerName: "存在しない 飼主",
			},
			repoErr: apperrors.WrapNotFound("owner", "999"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			owner: &model.Owner{
				ID:        1,
				ClinicID:  1,
				OwnerName: "エラー ケース",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				updateFn: func(_ context.Context, _ *model.Owner) error {
					return tt.repoErr
				},
			}
			svc := NewOwnerService(repo)

			err := svc.Update(context.Background(), tt.owner)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOwnerService_CreateWithPets(t *testing.T) {
	tests := []struct {
		name    string
		owner   *model.Owner
		pets    []model.Pet
		repoErr error
		wantErr bool
	}{
		{
			name: "creates owner with pets atomically",
			owner: &model.Owner{
				ClinicID:  1,
				OwnerName: "林 文昭",
				Email:     "hayashi@example.com",
			},
			pets: []model.Pet{
				{Name: "ポチ", AnimalSpeciesID: 1, Gender: model.PetGenderMale},
				{Name: "タマ", AnimalSpeciesID: 2, Gender: model.PetGenderFemale},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates owner without pets (empty slice)",
			owner: &model.Owner{
				ClinicID:  1,
				OwnerName: "鈴木 次郎",
			},
			pets:    []model.Pet{},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "propagates repository error",
			owner: &model.Owner{
				ClinicID:  1,
				OwnerName: "エラー 飼主",
			},
			pets:    []model.Pet{{Name: "ペット"}},
			repoErr: errors.New("transaction failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				createWithPetsFn: func(_ context.Context, _ *model.Owner, _ []model.Pet) error {
					return tt.repoErr
				},
			}
			svc := NewOwnerService(repo)

			err := svc.CreateWithPets(context.Background(), tt.owner, tt.pets)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOwnerService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes owner successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when owner does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("owner", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewOwnerService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

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
