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

// --- Pet DB column constants ---

const (
	colPetOwnerID         = "owner_id"
	colPetAnimalSpeciesID = "animal_species_id"
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
	OwnerID         uint64
	AnimalSpeciesID uint64
	Name            string
	PetNameKana     string
	Gender          string
	Status          string
	BirthDate       *time.Time
	Breed           string
	Color           string
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType string
	DangerLevel     string
	Food            string
	Environment     string
	Phone           string
	InsuranceID     *uint64
	Remarks         string
}

// UpdatePetInput はペット更新の入力DTO（全フィールドポインタ型: nil = 未指定, 非nil = 更新対象）
// InsuranceID は **uint64: nil = フィールド未送信, &nil = NULL クリア, &&value = 値セット
// ビジネスルールは Service 層で検証する。
type UpdatePetInput struct {
	OwnerID         *uint64
	AnimalSpeciesID *uint64
	PetNumber       *string // 自動採番後も手動変更可
	Name            *string
	PetNameKana     *string
	Gender          *string
	Status          *string
	BirthDate       *time.Time
	Breed           *string
	Color           *string
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType *string
	DangerLevel     *string
	Food            *string
	Environment     *string
	Phone           *string
	LastVisit       *time.Time
	InsuranceID     **uint64
	Remarks         *string
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
	repo          repository.PetRepository
	ownerRepo     repository.OwnerRepository
	insuranceRepo repository.InsuranceRepository
}

func NewPetService(
	repo repository.PetRepository,
	ownerRepo repository.OwnerRepository,
	insuranceRepo repository.InsuranceRepository,
) PetService {
	return &petService{
		repo:          repo,
		ownerRepo:     ownerRepo,
		insuranceRepo: insuranceRepo,
	}
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
	if err := validatePetAcquisitionType(input.AcquisitionType); err != nil {
		return nil, err
	}
	if err := validatePetDangerLevel(input.DangerLevel); err != nil {
		return nil, err
	}

	// owner_id の clinic 所属確認
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, input.OwnerID); err != nil {
		return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
	}

	// insurance_id の clinic 所属確認（指定された場合のみ）
	if input.InsuranceID != nil {
		insurance, err := s.insuranceRepo.FindByID(ctx, *input.InsuranceID)
		if err != nil || insurance.ClinicID != clinicID {
			return nil, apperrors.WrapInvalidInput("insurance not found in this clinic")
		}
	}

	// ペット番号を自動採番: 「飼主ID-連番」形式
	count, err := s.repo.CountByOwner(ctx, clinicID, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pets by owner: %w", err)
	}
	petNumber := fmt.Sprintf("%d-%d", input.OwnerID, count+1)

	// DTO → Model 変換
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       petNumber,
		Name:            input.Name,
		PetNameKana:     input.PetNameKana,
		BirthDate:       input.BirthDate,
		Breed:           input.Breed,
		Color:           input.Color,
		Weight:          input.Weight,
		NeuteredDate:    input.NeuteredDate,
		Food:            input.Food,
		Environment:     input.Environment,
		Phone:           input.Phone,
		InsuranceID:     input.InsuranceID,
		Remarks:         input.Remarks,
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}
	if input.AcquisitionType != "" {
		at := model.AcquisitionType(input.AcquisitionType)
		pet.AcquisitionType = &at
	}
	if input.DangerLevel != "" {
		pet.DangerLevel = model.DangerLevel(input.DangerLevel)
	}

	if err := s.repo.Create(ctx, pet); err != nil {
		return nil, fmt.Errorf("failed to create pet: %w", err)
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
	if input.AcquisitionType != nil {
		if err := validatePetAcquisitionType(*input.AcquisitionType); err != nil {
			return nil, err
		}
	}
	if input.DangerLevel != nil {
		if err := validatePetDangerLevel(*input.DangerLevel); err != nil {
			return nil, err
		}
	}

	// owner_id 変更時の clinic 所属確認
	if input.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
			return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// insurance_id 変更時の clinic 所属確認（非 nil かつ値が指定された場合のみ）
	if input.InsuranceID != nil && *input.InsuranceID != nil {
		insurance, err := s.insuranceRepo.FindByID(ctx, **input.InsuranceID)
		if err != nil || insurance.ClinicID != clinicID {
			return nil, apperrors.WrapInvalidInput("insurance not found in this clinic")
		}
	}

	fields := buildPetUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, fmt.Errorf("failed to update pet: %w", err)
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
		fields["pet_number"] = *input.PetNumber
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
	if input.Status != nil {
		fields[colPetStatus] = *input.Status
	}
	if input.BirthDate != nil {
		fields[colPetBirthDate] = *input.BirthDate
	}
	if input.Breed != nil {
		fields[colPetBreed] = *input.Breed
	}
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	if input.Weight != nil {
		fields[colPetWeight] = *input.Weight
	}
	if input.NeuteredDate != nil {
		fields["neutered_date"] = *input.NeuteredDate
	}
	if input.AcquisitionType != nil {
		fields["acquisition_type"] = *input.AcquisitionType
	}
	if input.DangerLevel != nil {
		fields["danger_level"] = *input.DangerLevel
	}
	if input.Food != nil {
		fields["food"] = *input.Food
	}
	if input.Environment != nil {
		fields[colPetEnvironment] = *input.Environment
	}
	if input.Phone != nil {
		fields["phone"] = *input.Phone
	}
	if input.LastVisit != nil {
		fields["last_visit"] = *input.LastVisit
	}
	if input.InsuranceID != nil {
		// *input.InsuranceID は *uint64: nil = NULL クリア、非nil = 値セット
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
