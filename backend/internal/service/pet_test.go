package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type MockPetRepository struct {
	mock.Mock
}

func (m *MockPetRepository) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Pet), args.Error(1)
}

func (m *MockPetRepository) GetPetByID(ctx context.Context, id uuid.UUID) (*model.Pet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockPetRepository) CreatePet(ctx context.Context, pet *model.Pet) error {
	args := m.Called(ctx, pet)
	return args.Error(0)
}

func (m *MockPetRepository) UpdatePet(ctx context.Context, pet *model.Pet) error {
	args := m.Called(ctx, pet)
	return args.Error(0)
}

func (m *MockPetRepository) DeletePet(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockOwnerRepository struct {
	mock.Mock
}

func (m *MockOwnerRepository) GetAllOwners(ctx context.Context) ([]model.Owner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Owner), args.Error(1)
}

func (m *MockOwnerRepository) GetOwnerByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockOwnerRepository) CreateOwner(ctx context.Context, owner *model.Owner) error {
	args := m.Called(ctx, owner)
	return args.Error(0)
}

func (m *MockOwnerRepository) UpdateOwner(ctx context.Context, owner *model.Owner) error {
	args := m.Called(ctx, owner)
	return args.Error(0)
}

func (m *MockOwnerRepository) DeleteOwner(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetAllPets(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	expectedPets := []model.Pet{
		{ID: uuid.New(), Name: "Pochi"},
	}

	mockRepo.On("GetAllPets", ctx).Return(expectedPets, nil)

	pets, err := svc.GetAllPets(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedPets, pets)
	mockRepo.AssertExpectations(t)
}

func TestGetPetByID(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	id := uuid.New()
	expectedPet := &model.Pet{ID: id, Name: "Pochi"}

	mockRepo.On("GetPetByID", ctx, id).Return(expectedPet, nil)

	pet, err := svc.GetPetByID(ctx, id.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedPet, pet)
	mockRepo.AssertExpectations(t)
}

func TestCreatePet(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	req := &model.CreatePetRequest{
		Name:    "Pochi",
		Species: "Dog",
		OwnerID: uuid.New().String(),
	}

	mockRepo.On("CreatePet", ctx, mock.MatchedBy(func(p *model.Pet) bool {
		return p.Name == req.Name && p.Species == req.Species
	})).Return(nil)

	pet, err := svc.CreatePet(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.Equal(t, req.Name, pet.Name)
	mockRepo.AssertExpectations(t)
}

func TestCreatePet_Validation(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *model.CreatePetRequest
	}{
		{
			name: "missing name",
			req:  &model.CreatePetRequest{Species: "Dog"},
		},
		{
			name: "missing species",
			req:  &model.CreatePetRequest{Name: "Pochi"},
		},
		{
			name: "invalid birth date",
			req:  &model.CreatePetRequest{Name: "Pochi", Species: "Dog", BirthDate: "invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pet, err := svc.CreatePet(ctx, tt.req)
			assert.Error(t, err)
			assert.Nil(t, pet)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestUpdatePet(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	id := uuid.New()
	existingPet := &model.Pet{ID: id, Name: "Pochi", Species: "Dog"}
	req := &model.UpdatePetRequest{Name: "Tama"}

	mockRepo.On("GetPetByID", ctx, id).Return(existingPet, nil)
	mockRepo.On("UpdatePet", ctx, mock.MatchedBy(func(p *model.Pet) bool {
		return p.ID == id && p.Name == "Tama" && p.Species == "Dog"
	})).Return(nil)

	pet, err := svc.UpdatePet(ctx, id.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Tama", pet.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePet_NotFound(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	id := uuid.New()
	mockRepo.On("GetPetByID", ctx, id).Return(nil, errors.New("not found"))

	pet, err := svc.UpdatePet(ctx, id.String(), &model.UpdatePetRequest{Name: "Tama"})

	assert.Error(t, err)
	assert.Nil(t, pet)
}

func TestDeletePet(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	id := uuid.New()
	mockRepo.On("DeletePet", ctx, id).Return(nil)

	err := svc.DeletePet(ctx, id.String())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePet_InvalidID(t *testing.T) {
	mockRepo := new(MockPetRepository)
	mockOwnerRepo := new(MockOwnerRepository)
	svc := New(mockRepo, mockOwnerRepo, nil, nil)
	ctx := context.Background()

	err := svc.DeletePet(ctx, "invalid-id")

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}
