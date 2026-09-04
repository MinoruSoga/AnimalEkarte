package pet

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	colPetBloodType       = "blood_type"
	colPetMicrochipNumber = "microchip_number"
	colPetWeight          = "weight"
	colPetEnvironment     = "environment"
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
	BloodType       string
	MicrochipNumber string
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType string
	DangerLevel     string
	DangerReason    *string
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
	// Status は意図的に持たない(BUG-415)。generic update から status を書けなくし、
	// 唯一の書込元を Create と HandlePetDeath/HandlePetRevival(監査+deceased_at 同一tx・
	// fail-closed)に一本化する。
	BirthDate       *time.Time
	Breed           *string
	Color           *string
	BloodType       *string
	MicrochipNumber *string
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType *string
	DangerLevel     *string
	// DangerReason は nil=未指定 / &nil=NULLクリア / &&value=更新対象。
	DangerReason **string
	Food         *string
	Environment  *string
	Phone        *string
	LastVisit    *time.Time
	InsuranceID  **uint64
	Remarks      *string
}

// PetUpdate is the typed atomic update command owned by the pet package.
// The repository merges its danger fields over the locked row before writing.
type PetUpdate struct {
	fields       map[string]any
	dangerLevel  *model.DangerLevel
	dangerReason **string
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
		fields[colPetNameKana] = normalizeNameKana(*input.PetNameKana)
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
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	if input.BloodType != nil {
		fields[colPetBloodType] = *input.BloodType
	}
	if input.MicrochipNumber != nil {
		fields[colPetMicrochipNumber] = *input.MicrochipNumber
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
	if input.DangerReason != nil {
		fields["danger_reason"] = *input.DangerReason
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

type Service interface {
	// List は指定した複数医院 (#86 拠点横断) のペット一覧を返す。clinicIDs はハンドラ層で所属検証済みであること。
	List(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error)
	// ListOwnerReportPets は認可済み医院内の対象飼主について、Owner Report 用のペット一覧を返す。
	ListOwnerReportPets(ctx context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	// GetByIDForClinics は複数医院スコープでペットを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
	Create(ctx context.Context, clinicID uint64, input *CreatePetInput) (*model.Pet, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// GetFirstVisitDate は指定ペットの初診日（最古の有効カルテ date）を返す（#158 飼主レポート）。
	// カルテが無い場合は nil, nil を返す。clinic 隔離・論理削除除外は repository が担保する。
	GetFirstVisitDate(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
}

// --- Implementation ---

type petService struct {
	repo              ServiceRepository
	ownerRepo         OwnerFinder
	insuranceRepo     InsuranceFinder
	medicalRecordRepo MedicalRecordReader
	tagSyncSvc        PetTagSynchronizer
	petOwnerRepo      PetOwnerReader
	petOwnerTx        PetOwnerTransactor
}

// NewService constructs the pet use-case service.
func NewService(
	repo ServiceRepository,
	ownerRepo OwnerFinder,
	insuranceRepo InsuranceFinder,
	medicalRecordRepo MedicalRecordReader,
	tagSyncSvc PetTagSynchronizer,
) Service {
	return newService(repo, ownerRepo, insuranceRepo, medicalRecordRepo, tagSyncSvc, nil, nil)
}

// NewServiceWithPetOwnerReader constructs the pet service with the secondary-owner
// guard required by production owner changes.
func NewServiceWithPetOwnerReader(
	repo ServiceRepository,
	ownerRepo OwnerFinder,
	insuranceRepo InsuranceFinder,
	medicalRecordRepo MedicalRecordReader,
	tagSyncSvc PetTagSynchronizer,
	petOwnerRepo PetOwnerReader,
	petOwnerTx PetOwnerTransactor,
) Service {
	return newService(repo, ownerRepo, insuranceRepo, medicalRecordRepo, tagSyncSvc, petOwnerRepo, petOwnerTx)
}

func newService(
	repo ServiceRepository,
	ownerRepo OwnerFinder,
	insuranceRepo InsuranceFinder,
	medicalRecordRepo MedicalRecordReader,
	tagSyncSvc PetTagSynchronizer,
	petOwnerRepo PetOwnerReader,
	petOwnerTx PetOwnerTransactor,
) Service {
	return &petService{
		repo:              repo,
		ownerRepo:         ownerRepo,
		insuranceRepo:     insuranceRepo,
		medicalRecordRepo: medicalRecordRepo,
		tagSyncSvc:        tagSyncSvc,
		petOwnerRepo:      petOwnerRepo,
		petOwnerTx:        petOwnerTx,
	}
}

func (s *petService) List(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error) {
	pets, total, err := s.repo.FindAll(ctx, clinicIDs, filters, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list pets", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list pets")
	}
	return pets, total, nil
}

func (s *petService) ListOwnerReportPets(ctx context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error) {
	pets, err := s.repo.FindOwnerReportPets(ctx, clinicIDs, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list owner report pets", "error", err)
		return nil, apperrors.Wrap(err, "failed to list owner report pets")
	}
	return pets, nil
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to get pet")
	}
	return pet, nil
}

func (s *petService) GetFirstVisitDate(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	// 当該医院にペットが存在することを先に確認する。存在しない／他医院のペット ID には
	// 404 を返し、初診日 null との区別（ID 存在の推測）を与えない（IDOR 対策）。
	if _, err := s.repo.FindByID(ctx, clinicID, petID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet")
	}
	date, err := s.medicalRecordRepo.FindFirstVisitDateByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get pet first visit date", "error", err)
		return nil, apperrors.Wrap(err, "failed to get pet first visit date")
	}
	return date, nil
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
	normalizedReason, err := normalizeDangerReason(
		model.DangerLevel(input.DangerLevel),
		input.DangerReason,
	)
	if err != nil {
		return nil, err
	}
	input.DangerReason = normalizedReason

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

	// 採番は pet-owned repository capability が owner row lock と同一
	// transaction 内で行う。service で count-before-create しない。
	pet := buildPetModel(clinicID, "", input)

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
	currentPet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to find pet")
	}

	if err := validateUpdatePetInput(input); err != nil {
		return nil, err
	}

	// owner_id 変更時の clinic 所属確認
	ownerUpdateRequested := false
	if input.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
			return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
		}
		if currentPet == nil {
			return nil, apperrors.WrapInternalServerError("pet lookup returned no result")
		}
		if s.petOwnerRepo == nil || s.petOwnerTx == nil {
			return nil, apperrors.WrapInternalServerError("pet owner transaction dependencies are unavailable")
		}
		ownerUpdateRequested = true
		links, err := s.petOwnerRepo.FindByPetID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to find pet owners")
		}
		if containsPetOwner(links, *input.OwnerID) {
			return nil, apperrors.WrapConflict("副飼主を主飼主へ変更する前に副飼主の紐付けを解除してください")
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

	var dangerLevel *model.DangerLevel
	if input.DangerLevel != nil {
		level := model.DangerLevel(*input.DangerLevel)
		dangerLevel = &level
	}
	update := PetUpdate{
		fields:       fields,
		dangerLevel:  dangerLevel,
		dangerReason: input.DangerReason,
	}
	var pet *model.Pet
	updateAndVerify := func(updateCtx context.Context) error {
		var updateErr error
		pet, updateErr = s.repo.UpdateAndFind(updateCtx, clinicID, id, update)
		if updateErr != nil {
			return updateErr
		}
		if !ownerUpdateRequested {
			return nil
		}
		links, readErr := s.petOwnerRepo.FindByPetID(updateCtx, clinicID, id)
		if readErr != nil {
			return apperrors.Wrap(readErr, "failed to verify pet owners")
		}
		if containsPetOwner(links, *input.OwnerID) {
			return apperrors.WrapConflict("副飼主を主飼主へ変更する前に副飼主の紐付けを解除してください")
		}
		return nil
	}
	if ownerUpdateRequested {
		err = s.petOwnerTx.WithTx(ctx, updateAndVerify)
	} else {
		err = updateAndVerify(ctx)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to update pet", "error", err)
		return nil, apperrors.Wrap(err, "failed to update pet")
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

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
