package handler

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Pet), args.Error(1)
}

func (m *MockService) GetPetByID(ctx context.Context, id string) (*model.Pet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) DeletePet(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Owner Mock Methods
func (m *MockService) GetAllOwners(ctx context.Context) ([]model.Owner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Owner), args.Error(1)
}

func (m *MockService) GetOwnerByID(ctx context.Context, id string) (*model.Owner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) CreateOwner(ctx context.Context, req *model.CreateOwnerRequest) (*model.Owner, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) UpdateOwner(ctx context.Context, id string, req *model.UpdateOwnerRequest) (*model.Owner, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) DeleteOwner(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetDB Mock Method
func (m *MockService) GetDB() (interface{ DB() *gorm.DB }, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interface{ DB() *gorm.DB }), args.Error(1)
}
