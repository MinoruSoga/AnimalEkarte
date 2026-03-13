package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- Pet DB column constants ---

const (
	colPetOwnerID         = "owner_id"
	colPetAnimalSpeciesID = "animal_species_id"
	colPetNumber          = "pet_number"
	colPetName            = "name"
	colPetNameKana        = "pet_name_kana"
	colPetGender          = "gender"
	colPetBirthDate       = "birth_date"
	colPetBreed           = "breed"
	colPetWeight          = "weight"
	colPetEnvironment     = "environment"
	colPetStatus          = "status"
	colPetInsuranceID     = "insurance_id"
	colPetRemarks         = "remarks"
)

// --- Input DTOs（Service層の公開I/F） ---

// CreatePetInput はペット作成の入力DTO
// ビジネスルール（Gender/Status 列挙値）は Service 層で検証する。
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

// UpdatePetInput はペット更新の入力DTO（全フィールドポインタ型: nil = 未指定, 非nil = 更新対象）
// ビジネスルールは Service 層で検証する。
type UpdatePetInput struct {
	OwnerID         *uint64    `json:"owner_id"`
	AnimalSpeciesID *uint64    `json:"animal_species_id"`
	PetNumber       *string    `json:"pet_number"`
	Name            *string    `json:"name"`
	PetNameKana     *string    `json:"pet_name_kana"`
	Gender          *string    `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           *string    `json:"breed"`
	Weight          *float64   `json:"weight"`
	Environment     *string    `json:"environment"`
	Status          *string    `json:"status"`
	InsuranceID     *uint64    `json:"insurance_id"`
	Remarks         *string    `json:"remarks"`
}

// --- Interface ---

type PetService interface {
	List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error)
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

func (s *petService) Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error) {
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

func (s *petService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	// ビジネスルールバリデーション
	if input.Gender != nil {
		if err := validatePetGender(*input.Gender); err != nil {
			return nil, err
		}
	}
	if input.Status != nil {
		if err := validatePetStatus(*input.Status); err != nil {
			return nil, err
		}
	}

	fields := buildPetUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

	return s.repo.FindByID(ctx, clinicID, id)
}

// buildPetUpdateFields はポインタが非 nil のフィールドのみ map に追加する
func buildPetUpdateFields(input *UpdatePetInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerID != nil {
		fields[colPetOwnerID] = *input.OwnerID
	}
	if input.AnimalSpeciesID != nil {
		fields[colPetAnimalSpeciesID] = *input.AnimalSpeciesID
	}
	if input.PetNumber != nil {
		fields[colPetNumber] = *input.PetNumber
	}
	if input.Name != nil {
		fields[colPetName] = *input.Name
	}
	if input.PetNameKana != nil {
		fields[colPetNameKana] = *input.PetNameKana
	}
	if input.Gender != nil {
		fields[colPetGender] = *input.Gender
	}
	if input.BirthDate != nil {
		fields[colPetBirthDate] = *input.BirthDate
	}
	if input.Breed != nil {
		fields[colPetBreed] = *input.Breed
	}
	if input.Weight != nil {
		fields[colPetWeight] = *input.Weight
	}
	if input.Environment != nil {
		fields[colPetEnvironment] = *input.Environment
	}
	if input.Status != nil {
		fields[colPetStatus] = *input.Status
	}
	if input.InsuranceID != nil {
		fields[colPetInsuranceID] = *input.InsuranceID
	}
	if input.Remarks != nil {
		fields[colPetRemarks] = *input.Remarks
	}
	return fields
}

func (s *petService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
