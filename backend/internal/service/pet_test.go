package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MockPetRepository is a mock implementation of PetRepository
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

func TestGetAllPets_Success(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	expectedPets := []model.Pet{
		{ID: uuid.New(), Name: "ポチ", Species: "犬"},
		{ID: uuid.New(), Name: "タマ", Species: "猫"},
	}

	mockRepo.On("GetAllPets", ctx).Return(expectedPets, nil)

	pets, err := svc.GetAllPets(ctx)

	assert.NoError(t, err)
	assert.Len(t, pets, 2)
	assert.Equal(t, "ポチ", pets[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestGetPetByID_Success(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	petID := uuid.New()
	expectedPet := &model.Pet{ID: petID, Name: "ポチ", Species: "犬"}

	mockRepo.On("GetPetByID", ctx, petID).Return(expectedPet, nil)

	pet, err := svc.GetPetByID(ctx, petID.String())

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.Equal(t, "ポチ", pet.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetPetByID_InvalidUUID(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	pet, err := svc.GetPetByID(ctx, "invalid-uuid")

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestGetPetByID_NotFound(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	petID := uuid.New()
	mockRepo.On("GetPetByID", ctx, petID).Return(nil, apperrors.WrapNotFound("pet", petID.String()))

	pet, err := svc.GetPetByID(ctx, petID.String())

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
}

func TestCreatePet_Success(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	req := &model.CreatePetRequest{
		Name:      "ポチ",
		Species:   "犬",
		Breed:     "柴犬",
		Gender:    "オス",
		Weight:    10.5,
		BirthDate: "2020-01-15",
	}

	mockRepo.On("CreatePet", ctx, mock.AnythingOfType("*model.Pet")).Return(nil)

	pet, err := svc.CreatePet(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.Equal(t, "ポチ", pet.Name)
	assert.Equal(t, "犬", pet.Species)
	assert.Equal(t, "柴犬", pet.Breed)
	assert.NotNil(t, pet.BirthDate)
	mockRepo.AssertExpectations(t)
}

func TestCreatePet_EmptyName(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	req := &model.CreatePetRequest{
		Name:    "",
		Species: "犬",
	}

	pet, err := svc.CreatePet(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestCreatePet_EmptySpecies(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	req := &model.CreatePetRequest{
		Name:    "ポチ",
		Species: "",
	}

	pet, err := svc.CreatePet(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestCreatePet_InvalidBirthDate(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	req := &model.CreatePetRequest{
		Name:      "ポチ",
		Species:   "犬",
		BirthDate: "invalid-date",
	}

	pet, err := svc.CreatePet(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestUpdatePet_Success(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	petID := uuid.New()
	birthDate := time.Now().AddDate(-2, 0, 0)
	weight := 10.5
	existingPet := &model.Pet{
		ID:        petID,
		Name:      "ポチ",
		Species:   "犬",
		Breed:     "柴犬",
		BirthDate: &birthDate,
		Gender:    "オス",
		Weight:    &weight,
	}

	req := &model.UpdatePetRequest{
		Name:   "ポチ太郎",
		Weight: 12.0,
	}

	mockRepo.On("GetPetByID", ctx, petID).Return(existingPet, nil)
	mockRepo.On("UpdatePet", ctx, mock.AnythingOfType("*model.Pet")).Return(nil)

	pet, err := svc.UpdatePet(ctx, petID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.Equal(t, "ポチ太郎", pet.Name)
	assert.Equal(t, 12.0, *pet.Weight)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePet_NotFound(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	petID := uuid.New()
	req := &model.UpdatePetRequest{Name: "ポチ太郎"}

	mockRepo.On("GetPetByID", ctx, petID).Return(nil, apperrors.WrapNotFound("pet", petID.String()))

	pet, err := svc.UpdatePet(ctx, petID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.True(t, apperrors.IsNotFound(err))
	mockRepo.AssertExpectations(t)
}

func TestDeletePet_Success(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	petID := uuid.New()
	mockRepo.On("DeletePet", ctx, petID).Return(nil)

	err := svc.DeletePet(ctx, petID.String())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePet_InvalidUUID(t *testing.T) {
	mockRepo := new(MockPetRepository)
	svc := New(mockRepo)
	ctx := context.Background()

	err := svc.DeletePet(ctx, "invalid-uuid")

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}
