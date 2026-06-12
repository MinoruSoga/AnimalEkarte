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

// --- Interface ---

type PetService interface {
	List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	// GetByIDForClinics は複数医院スコープでペットを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
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
	tagSyncSvc        LstepTagSyncService
}

func NewPetService(
	repo repository.PetRepository,
	ownerRepo repository.OwnerRepository,
	insuranceRepo repository.InsuranceRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
	tagSyncSvc LstepTagSyncService,
) PetService {
	return &petService{
		repo:              repo,
		ownerRepo:         ownerRepo,
		insuranceRepo:     insuranceRepo,
		medicalRecordRepo: medicalRecordRepo,
		tagSyncSvc:        tagSyncSvc,
	}
}

func (s *petService) List(ctx context.Context, clinicID uint64, ownerID *uint64, page, limit int, search string) ([]model.Pet, int64, error) {
	pets, total, err := s.repo.FindAll(ctx, clinicID, ownerID, page, limit, search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list pets", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list pets")
	}
	return pets, total, nil
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to get pet")
	}
	return pet, nil
}

func (s *petService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get pet for clinics", "error", err)
		return nil, apperrors.Wrap(err, "failed to get pet for clinics")
	}
	return pet, nil
}

func (s *petService) Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error) {
	if err := validateCreatePetInput(input); err != nil {
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
		slog.ErrorContext(ctx, "failed to count pets by owner", "error", err)
		return nil, apperrors.Wrap(err, "failed to count pets by owner")
	}
	petNumber := fmt.Sprintf("%d-%d", input.OwnerID, count+1)

	pet := buildPetModel(clinicID, petNumber, input)

	if err := s.repo.Create(ctx, pet); err != nil {
		slog.ErrorContext(ctx, "failed to create pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to create pet")
	}

	slog.InfoContext(ctx, "pet created",
		slog.Uint64("pet_id", pet.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("owner_id", input.OwnerID))
	s.syncLstepTags(ctx, clinicID, pet.OwnerID)

	return pet, nil
}

func (s *petService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to find pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to find pet")
	}

	if err := validateUpdatePetInput(input); err != nil {
		return nil, err
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
		slog.ErrorContext(ctx, "failed to update pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to update pet")
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to get updated pet")
	}
	s.syncLstepTags(ctx, clinicID, pet.OwnerID)
	return pet, nil
}

func (s *petService) syncLstepTags(ctx context.Context, clinicID, ownerID uint64) {
	if s.tagSyncSvc == nil {
		return
	}
	if err := s.tagSyncSvc.SyncOwnerAnimalClassificationTags(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync animal classification tags after pet change", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
	}
	if err := s.tagSyncSvc.SyncPetBasicInfoTags(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync pet basic info tags after pet change", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
	}
}

func (s *petService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find pet")
	}
	// FK依存チェック: カルテが紐付いている場合は削除を拒否
	recordCount, err := s.medicalRecordRepo.CountByPetID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check medical record dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check medical record dependencies")
	}
	if recordCount > 0 {
		return apperrors.WrapConflict("カルテが紐付いているため削除できません。先にカルテを削除またはこのペットを変更してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete pet", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete pet")
	}
	slog.InfoContext(ctx, "pet deleted",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
