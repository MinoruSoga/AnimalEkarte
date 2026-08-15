package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestChronicConditionService_UpdateRejectsConditionOwnedByAnotherPet(t *testing.T) {
	updateCalled := false
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
		},
	}
	repo := &mockPetChronicConditionRepository{
		findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(100), petID)
			assert.Equal(t, uint64(10), id)
			return nil, apperrors.WrapNotFound("pet_chronic_condition", "10")
		},
		updateFn: func(_ context.Context, _, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	svc := NewChronicConditionService(repo, petRepo, nil)

	got, err := svc.Update(context.Background(), 1, 100, 10, UpdateChronicConditionInput{})

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.False(t, updateCalled)
}

func TestChronicConditionService_DeleteRejectsConditionOwnedByAnotherPet(t *testing.T) {
	deleteCalled := false
	petRepo := &mockPetRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
		},
	}
	repo := &mockPetChronicConditionRepository{
		findByIDFn: func(_ context.Context, clinicID, petID, id uint64) (*model.PetChronicCondition, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(100), petID)
			assert.Equal(t, uint64(10), id)
			return nil, apperrors.WrapNotFound("pet_chronic_condition", "10")
		},
		deleteFn: func(_ context.Context, _, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewChronicConditionService(repo, petRepo, nil)

	err := svc.Delete(context.Background(), 1, 100, 10)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.False(t, deleteCalled)
}
