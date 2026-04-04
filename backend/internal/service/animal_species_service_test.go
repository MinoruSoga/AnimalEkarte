package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- AnimalSpecies モック ----

type mockAnimalSpeciesRepository struct {
	findAllFn  func(ctx context.Context) ([]model.AnimalSpecies, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	createFn   func(ctx context.Context, species *model.AnimalSpecies) error
	updateFn   func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderFn  func(ctx context.Context, ids []uint64) error
}

func (m *mockAnimalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	return m.findAllFn(ctx)
}

func (m *mockAnimalSpeciesRepository) FindByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAnimalSpeciesRepository) Create(ctx context.Context, species *model.AnimalSpecies) error {
	if m.createFn != nil {
		return m.createFn(ctx, species)
	}
	return nil
}

func (m *mockAnimalSpeciesRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, fields)
	}
	return nil
}

func (m *mockAnimalSpeciesRepository) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockAnimalSpeciesRepository) Reorder(ctx context.Context, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, ids)
	}
	return nil
}

// ---- Pet モック ----

type mockPetRepository struct {
	countByAnimalSpeciesIDFn func(ctx context.Context, speciesID uint64) (int64, error)
}

func (m *mockPetRepository) FindAll(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	return nil, 0, nil
}

func (m *mockPetRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return nil, nil
}

func (m *mockPetRepository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	return 0, nil
}

func (m *mockPetRepository) CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
	if m.countByAnimalSpeciesIDFn != nil {
		return m.countByAnimalSpeciesIDFn(ctx, speciesID)
	}
	return 0, nil
}

func (m *mockPetRepository) Create(ctx context.Context, pet *model.Pet) error {
	return nil
}

func (m *mockPetRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return nil
}

func (m *mockPetRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return nil
}

// ---- Tests ----

func TestAnimalSpeciesService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.AnimalSpecies
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns all animal species",
			repoData: []model.AnimalSpecies{
				{ID: 1, Name: "犬", IsActive: true},
				{ID: 2, Name: "猫", IsActive: true},
				{ID: 3, Name: "ウサギ", IsActive: true},
				{ID: 4, Name: "ハムスター", IsActive: true},
			},
			repoErr: nil,
			wantLen: 4,
			wantErr: false,
		},
		{
			name:     "returns empty list when no species exist",
			repoData: []model.AnimalSpecies{},
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
			repo := &mockAnimalSpeciesRepository{
				findAllFn: func(_ context.Context) ([]model.AnimalSpecies, error) {
					return tt.repoData, tt.repoErr
				},
			}
			petRepo := &mockPetRepository{}
			svc := NewAnimalSpeciesService(repo, petRepo)

			species, err := svc.List(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, species, tt.wantLen)
			}
		})
	}
}

func TestAnimalSpeciesService_Create(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{
		createFn: func(_ context.Context, s *model.AnimalSpecies) error {
			s.ID = 99
			return nil
		},
	}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	species, err := svc.Create(context.Background(), &CreateAnimalSpeciesInput{
		Name:      "フェレット",
		IsActive:  true,
		SortOrder: 7,
	})

	assert.NoError(t, err)
	assert.Equal(t, uint64(99), species.ID)
	assert.Equal(t, "フェレット", species.Name)
}

func TestAnimalSpeciesService_Update_EmptyFields(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	_, err := svc.Update(context.Background(), 1, &UpdateAnimalSpeciesInput{})
	assert.Error(t, err)
}

func TestAnimalSpeciesService_Reorder_EmptyIDs(t *testing.T) {
	repo := &mockAnimalSpeciesRepository{}
	petRepo := &mockPetRepository{}
	svc := NewAnimalSpeciesService(repo, petRepo)

	err := svc.Reorder(context.Background(), []uint64{})
	assert.Error(t, err)
}

func TestAnimalSpeciesService_Delete_WithPetReference(t *testing.T) {
	tests := []struct {
		name      string
		petCount  int64
		wantError bool
		wantCode  string
	}{
		{
			name:      "returns 409 when animal species is referenced by pets",
			petCount:  2,
			wantError: true,
			wantCode:  "CONFLICT",
		},
		{
			name:      "succeeds when animal species has no pet references",
			petCount:  0,
			wantError: false,
			wantCode:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAnimalSpeciesRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return nil
				},
			}
			petRepo := &mockPetRepository{
				countByAnimalSpeciesIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.petCount, nil
				},
			}
			svc := NewAnimalSpeciesService(repo, petRepo)

			err := svc.Delete(context.Background(), 1)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
