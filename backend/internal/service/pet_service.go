package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- Input DTOs（Service層の公開I/F） ---

// CreatePetInput はペット作成の入力DTO
type CreatePetInput struct {
	OwnerID         uint64     `json:"owner_id" binding:"required"`
	AnimalSpeciesID uint64     `json:"animal_species_id" binding:"required"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name" binding:"required"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           string     `json:"breed"`
	Weight          *float64   `json:"weight"`
	Environment     string     `json:"environment"`
	Status          string     `json:"status"`
	InsuranceID     *uint64    `json:"insurance_id"`
	Remarks         string     `json:"remarks"`
}

// UpdatePetInput はペット更新の入力DTO
type UpdatePetInput struct {
	OwnerID         *uint64    `json:"owner_id"`
	AnimalSpeciesID *uint64    `json:"animal_species_id"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           string     `json:"breed"`
	Weight          *float64   `json:"weight"`
	Environment     string     `json:"environment"`
	Status          string     `json:"status"`
	InsuranceID     *uint64    `json:"insurance_id"`
	Remarks         string     `json:"remarks"`
}

// --- Interface ---

type PetService interface {
	List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	Create(ctx context.Context, clinicID uint64, input CreatePetInput) (*model.Pet, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdatePetInput) (*model.Pet, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// --- Implementation ---

type petService struct {
	repo repository.PetRepository
}

func NewPetService(repo repository.PetRepository) PetService {
	return &petService{repo: repo}
}

func (s *petService) List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	return s.repo.FindAll(ctx, clinicID, ownerID, page, limit, search)
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *petService) Create(ctx context.Context, clinicID uint64, input CreatePetInput) (*model.Pet, error) {
	// ビジネスルールバリデーション
	if err := validatePetGender(input.Gender); err != nil {
		return nil, err
	}
	if err := validatePetStatus(input.Status); err != nil {
		return nil, err
	}

	// DTO → Model 変換
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       input.PetNumber,
		Name:            input.Name,
		PetNameKana:     input.PetNameKana,
		BirthDate:       input.BirthDate,
		Breed:           input.Breed,
		Weight:          input.Weight,
		Environment:     input.Environment,
		InsuranceID:     input.InsuranceID,
		Remarks:         input.Remarks,
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}

	if err := s.repo.Create(ctx, pet); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "pet created",
		slog.Uint64("pet_id", pet.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("owner_id", input.OwnerID))

	return pet, nil
}

func (s *petService) Update(ctx context.Context, clinicID, id uint64, input UpdatePetInput) (*model.Pet, error) {
	// ビジネスルールバリデーション
	if err := validatePetGender(input.Gender); err != nil {
		return nil, err
	}
	if err := validatePetStatus(input.Status); err != nil {
		return nil, err
	}

	// DTO → Model 変換
	pet := &model.Pet{
		ID:          id,
		ClinicID:    clinicID,
		PetNumber:   input.PetNumber,
		Name:        input.Name,
		PetNameKana: input.PetNameKana,
		BirthDate:   input.BirthDate,
		Breed:       input.Breed,
		Weight:      input.Weight,
		Environment: input.Environment,
		InsuranceID: input.InsuranceID,
		Remarks:     input.Remarks,
	}
	if input.OwnerID != nil {
		pet.OwnerID = *input.OwnerID
	}
	if input.AnimalSpeciesID != nil {
		pet.AnimalSpeciesID = *input.AnimalSpeciesID
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}

	if err := s.repo.Update(ctx, pet); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

	return pet, nil
}

func (s *petService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}

// --- バリデーションヘルパー ---

func validatePetGender(gender string) error {
	if gender == "" {
		return nil
	}
	switch model.PetGender(gender) {
	case model.PetGenderMale, model.PetGenderFemale, model.PetGenderUnknown:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid gender: %s", gender))
	}
}

func validatePetStatus(status string) error {
	if status == "" {
		return nil
	}
	switch model.PetStatus(status) {
	case model.PetStatusAlive, model.PetStatusDeceased:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid status: %s", status))
	}
}
