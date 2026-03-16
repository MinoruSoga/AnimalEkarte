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
	findAllFn func(ctx context.Context) ([]model.AnimalSpecies, error)
}

func (m *mockAnimalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
	return m.findAllFn(ctx)
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
			svc := NewAnimalSpeciesService(repo)

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
