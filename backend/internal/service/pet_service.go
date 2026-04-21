package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	colPetNameKana        = "name_kana"
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
	repo              repository.PetRepository
	ownerRepo         repository.OwnerRepository
	insuranceRepo     repository.InsuranceRepository
	medicalRecordRepo repository.MedicalRecordRepository
}

func NewPetService(
	repo repository.PetRepository,
	ownerRepo repository.OwnerRepository,
	insuranceRepo repository.InsuranceRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
) PetService {
	return &petService{
		repo:              repo,
		ownerRepo:         ownerRepo,
		insuranceRepo:     insuranceRepo,
		medicalRecordRepo: medicalRecordRepo,
	}
}

func (s *petService) List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	pets, total, err := s.repo.FindAll(ctx, clinicID, ownerID, page, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list pets")
	}
	return pets, total, nil
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get pet")
	}
	return pet, nil
}

func (s *petService) Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error) {
	// 名前バリデーション（スペースのみ・NULL バイト・制御文字チェック）
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	input.Name = strings.TrimSpace(input.Name)

	// ビジネスルールバリデーション
	if input.Weight != nil && *input.Weight < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgWeightZeroOrMore)
	}
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
		if _, err := s.insuranceRepo.FindByID(ctx, clinicID, *input.InsuranceID); err != nil {
			return nil, apperrors.WrapInvalidInput("insurance not found in this clinic")
		}
	}

	// ペット番号を自動採番: 「飼主ID-連番」形式
	count, err := s.repo.CountByOwner(ctx, clinicID, input.OwnerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count pets by owner")
	}
	petNumber := fmt.Sprintf("%d-%d", input.OwnerID, count+1)

	// DTO → Model 変換
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       petNumber,
		Name:            input.Name,
		NameKana:        input.PetNameKana,
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
		return nil, apperrors.Wrap(err, "failed to create pet")
	}

	slog.InfoContext(ctx, "pet created",
		slog.Uint64("pet_id", pet.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("owner_id", input.OwnerID))

	return pet, nil
}

func (s *petService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet")
	}
	// 名前バリデーション（スペースのみ・NULL バイト・制御文字チェック）
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}

	// ビジネスルールバリデーション
	if input.Weight != nil && *input.Weight < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgWeightZeroOrMore)
	}
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
		if _, err := s.insuranceRepo.FindByID(ctx, clinicID, **input.InsuranceID); err != nil {
			return nil, apperrors.WrapInvalidInput("insurance not found in this clinic")
		}
	}

	fields := buildPetUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update pet")
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated pet")
	}
	return pet, nil
}

// buildPetUpdate はポインタが非 nil のフィールドのみ map に追加する
func buildPetUpdate(input *UpdatePetInput) map[string]any {
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
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find pet")
	}
	// FK依存チェック: カルテが紐付いている場合は削除を拒否
	recordCount, err := s.medicalRecordRepo.CountByPetID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check medical record dependencies")
	}
	if recordCount > 0 {
		return apperrors.WrapConflict("カルテが紐付いているため削除できません。先にカルテを削除またはこのペットを変更してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete pet")
	}
	slog.InfoContext(ctx, "pet deleted",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
