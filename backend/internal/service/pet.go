package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/validation"
)

func (s *Service) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	return s.repo.GetAllPets(ctx)
}

func (s *Service) GetPetByID(ctx context.Context, id string) (*model.Pet, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}
	return s.repo.GetPetByID(ctx, uid)
}

func (s *Service) CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error) {
	// バリデーション
	if err := validation.ValidateCreatePet(req); err != nil {
		return nil, err
	}

	// OwnerIDをUUIDに変換
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	pet := &model.Pet{
		OwnerID:   ownerID,
		PetNumber: req.PetNumber,
		Name:      req.Name,
		Species:   req.Species,
		Breed:     req.Breed,
		Gender:    req.Gender,
	}

	if req.Weight > 0 {
		pet.Weight = &req.Weight
	}

	if req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid birth date format, expected YYYY-MM-DD")
		}
		pet.BirthDate = &t
	}

	if err := s.repo.CreatePet(ctx, pet); err != nil {
		return nil, err
	}
	return pet, nil
}

func (s *Service) UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error) {
	// バリデーション
	if err := validation.ValidateUpdatePet(req); err != nil {
		return nil, err
	}

	pet, err := s.GetPetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		pet.Name = req.Name
	}
	if req.Species != "" {
		pet.Species = req.Species
	}
	if req.Breed != "" {
		pet.Breed = req.Breed
	}
	if req.Gender != "" {
		pet.Gender = req.Gender
	}
	if req.Weight > 0 {
		pet.Weight = &req.Weight
	}
	if req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid birth date format, expected YYYY-MM-DD")
		}
		pet.BirthDate = &t
	}

	if err := s.repo.UpdatePet(ctx, pet); err != nil {
		return nil, err
	}
	return pet, nil
}

func (s *Service) DeletePet(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid pet ID format")
	}
	return s.repo.DeletePet(ctx, uid)
}
