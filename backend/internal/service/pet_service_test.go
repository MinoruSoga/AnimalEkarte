package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockPetRepository は PetRepository のテスト用モック実装
type mockPetRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	createFn   func(ctx context.Context, pet *model.Pet) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockPetRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	return m.findAllFn(ctx, clinicID, ownerID, page, limit, search)
}

func (m *mockPetRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockPetRepository) Create(ctx context.Context, pet *model.Pet) error {
	return m.createFn(ctx, pet)
}

func (m *mockPetRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockPetRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func ptrUint64(v uint64) *uint64 { return &v }

func TestPetService_List(t *testing.T) {
	tests := []struct {
		name      string
		clinicID  uint64
		ownerID   *uint64
		page      int
		limit     int
		search    string
		repoPets  []model.Pet
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns all pets for clinic",
			clinicID: 1,
			ownerID:  nil,
			page:     1,
			limit:    20,
			search:   "",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
				{ID: 2, ClinicID: 1, OwnerID: 11, Name: "タマ"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters pets by owner_id",
			clinicID: 1,
			ownerID:  ptrUint64(10),
			page:     1,
			limit:    20,
			search:   "",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no pets exist",
			clinicID:  1,
			ownerID:   nil,
			page:      1,
			limit:     20,
			search:    "",
			repoPets:  []model.Pet{},
			repoTotal: 0,
			repoErr:   nil,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by search keyword",
			clinicID: 1,
			ownerID:  nil,
			page:     1,
			limit:    20,
			search:   "ポチ",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			ownerID:   nil,
			page:      1,
			limit:     20,
			search:    "",
			repoPets:  nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedOwnerID := (*uint64)(nil)
			repo := &mockPetRepository{
				findAllFn: func(_ context.Context, _ uint64, ownerID *uint64, _, _ int, _ string) ([]model.Pet, int64, error) {
					capturedOwnerID = ownerID
					return tt.repoPets, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewPetService(repo)

			pets, total, err := svc.List(context.Background(), tt.clinicID, tt.ownerID, tt.page, tt.limit, tt.search)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, pets, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestPetService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoPet  *model.Pet
		repoErr  error
		wantPet  *model.Pet
		wantErr  error
	}{
		{
			name:     "returns pet when found",
			clinicID: 1,
			id:       10,
			repoPet:  &model.Pet{ID: 10, ClinicID: 1, OwnerID: 5, Name: "ポチ"},
			repoErr:  nil,
			wantPet:  &model.Pet{ID: 10, ClinicID: 1, OwnerID: 5, Name: "ポチ"},
			wantErr:  nil,
		},
		{
			name:     "returns not found error when pet does not exist",
			clinicID: 1,
			id:       999,
			repoPet:  nil,
			repoErr:  apperrors.WrapNotFound("pet", "999"),
			wantPet:  nil,
			wantErr:  apperrors.ErrNotFound,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoPet:  nil,
			repoErr:  errors.New("db error"),
			wantPet:  nil,
			wantErr:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPetRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					return tt.repoPet, tt.repoErr
				},
			}
			svc := NewPetService(repo)

			pet, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPet, pet)
			}
		})
	}
}

func TestPetService_GetByID_NotFound(t *testing.T) {
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return nil, apperrors.WrapNotFound("pet", "999")
		},
	}
	svc := NewPetService(repo)

	pet, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, pet)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestPetService_Create(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		input    CreatePetInput
		repoErr  error
		wantErr  bool
		wantPet  bool
	}{
		{
			name:     "creates pet successfully",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "新しいペット",
				Gender:          "male",
			},
			repoErr: nil,
			wantErr: false,
			wantPet: true,
		},
		{
			name:     "rejects invalid gender",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "ペット",
				Gender:          "invalid",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "rejects invalid status",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "ペット",
				Status:          "invalid",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "returns already exists error on duplicate",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "既存ペット",
			},
			repoErr: apperrors.WrapAlreadyExists("pet", "既存ペット"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "エラーペット",
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPetRepository{
				createFn: func(_ context.Context, _ *model.Pet) error {
					return tt.repoErr
				},
			}
			svc := NewPetService(repo)

			pet, err := svc.Create(context.Background(), tt.clinicID, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pet)
			} else {
				assert.NoError(t, err)
				if tt.wantPet {
					assert.NotNil(t, pet)
					assert.Equal(t, tt.clinicID, pet.ClinicID)
					assert.Equal(t, tt.input.Name, pet.Name)
					assert.Equal(t, tt.input.OwnerID, pet.OwnerID)
				}
			}
		})
	}
}

func TestPetService_Update(t *testing.T) {
	updatedPet := &model.Pet{ID: 1, ClinicID: 1, Name: "更新後ペット名"}

	tests := []struct {
		name       string
		clinicID   uint64
		id         uint64
		input      UpdatePetInput
		updateErr  error
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
		wantErr    bool
		wantNF     bool
		wantPet    bool
	}{
		{
			name:     "updates pet successfully",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name:   ptrString("更新後ペット名"),
				Gender: ptrString("female"),
			},
			updateErr: nil,
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return updatedPet, nil
			},
			wantErr: false,
			wantPet: true,
		},
		{
			name:     "rejects invalid gender",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name:   ptrString("ペット"),
				Gender: ptrString("invalid"),
			},
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:     "rejects invalid status",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name:   ptrString("ペット"),
				Status: ptrString("invalid"),
			},
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:      "returns error when no fields provided",
			clinicID:  1,
			id:        1,
			input:     UpdatePetInput{}, // 全 nil
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:     "returns not found error when pet does not exist",
			clinicID: 1,
			id:       999,
			input: UpdatePetInput{
				Name: ptrString("存在しないペット"),
			},
			updateErr: apperrors.WrapNotFound("pet", "999"),
			wantErr:   true,
			wantNF:    true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name: ptrString("エラーケース"),
			},
			updateErr: errors.New("db error"),
			wantErr:   true,
			wantNF:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					return nil, errors.New("findByID should not be called")
				}
			}
			repo := &mockPetRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
				findByIDFn: findByIDFn,
			}
			svc := NewPetService(repo)

			pet, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pet)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				if tt.wantPet {
					assert.NotNil(t, pet)
				}
			}
		})
	}
}

func TestPetService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes pet successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when pet does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("pet", "999"),
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
			repo := &mockPetRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewPetService(repo)

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
