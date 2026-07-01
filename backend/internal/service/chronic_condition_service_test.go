package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockPetChronicConditionRepository struct {
	repository.PetChronicConditionRepository
	findByPetIDFn                    func(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error)
	findByIDFn                       func(ctx context.Context, clinicID, id uint64) (*model.PetChronicCondition, error)
	createFn                         func(ctx context.Context, record *model.PetChronicCondition) error
	updateFn                         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                         func(ctx context.Context, clinicID, id uint64) error
	getActiveConditionCodesByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) ([]string, error)
}

func (m *mockPetChronicConditionRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
	if m.findByPetIDFn != nil {
		return m.findByPetIDFn(ctx, clinicID, petID)
	}
	return []model.PetChronicCondition{}, nil
}

func (m *mockPetChronicConditionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
}

func (m *mockPetChronicConditionRepository) Create(ctx context.Context, record *model.PetChronicCondition) error {
	if m.createFn != nil {
		return m.createFn(ctx, record)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPetChronicConditionRepository) GetActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error) {
	if m.getActiveConditionCodesByOwnerFn != nil {
		return m.getActiveConditionCodesByOwnerFn(ctx, clinicID, ownerID)
	}
	return []string{}, nil
}

func TestChronicConditionService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockPetChronicConditionRepository{
			findByPetIDFn: func(_ context.Context, clinicID, petID uint64) ([]model.PetChronicCondition, error) {
				return []model.PetChronicCondition{{ID: 1, ClinicID: clinicID, PetID: petID}}, nil
			},
		}
		svc := NewChronicConditionService(repo, nil, nil)
		res, err := svc.List(ctx, 1, 100)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockPetChronicConditionRepository{
			findByPetIDFn: func(_ context.Context, _, _ uint64) ([]model.PetChronicCondition, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewChronicConditionService(repo, nil, nil)
		res, err := svc.List(ctx, 1, 100)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestChronicConditionService_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			createFn: func(_ context.Context, record *model.PetChronicCondition) error {
				record.ID = 10
				return nil
			},
			getActiveConditionCodesByOwnerFn: func(_ context.Context, _, _ uint64) ([]string, error) {
				return []string{"C01"}, nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		notes := "some notes"
		input := CreateChronicConditionInput{
			ConditionCode: "C01",
			ConditionName: "Disease 1",
			DiagnosedAt:   now,
			Notes:         &notes,
			IsActive:      true,
		}
		res, err := svc.Create(ctx, 1, 100, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), res.ID)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		_, err := svc.Create(ctx, 1, 100, CreateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("db create error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			createFn: func(_ context.Context, _ *model.PetChronicCondition) error {
				return errors.New("db error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		_, err := svc.Create(ctx, 1, 100, CreateChronicConditionInput{})
		assert.Error(t, err)
	})
}

func TestChronicConditionService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) error {
				assert.Equal(t, "C02", fields["condition_code"])
				return nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		code := "C02"
		input := UpdateChronicConditionInput{
			ConditionCode: &code,
		}
		res, err := svc.Update(ctx, 1, 100, 10, input)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("condition not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PetChronicCondition, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("update error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		_, err := svc.Update(ctx, 1, 100, 10, UpdateChronicConditionInput{})
		assert.Error(t, err)
	})
}

func TestChronicConditionService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID, OwnerID: 500}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		}
		tagSync := &mockLstepTagSyncService{}
		svc := NewChronicConditionService(repo, petRepo, tagSync)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.NoError(t, err)
	})

	t.Run("pet not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, errors.New("pet not found")
			},
		}
		svc := NewChronicConditionService(nil, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})

	t.Run("condition not found", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PetChronicCondition, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		petRepo := &mockPetRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
				return &model.Pet{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockPetChronicConditionRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PetChronicCondition, error) {
				return &model.PetChronicCondition{ID: id, ClinicID: clinicID}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("delete error")
			},
		}
		svc := NewChronicConditionService(repo, petRepo, nil)
		err := svc.Delete(ctx, 1, 100, 10)
		assert.Error(t, err)
	})
}
