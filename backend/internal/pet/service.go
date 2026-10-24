package pet

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	colPetNameOrigin      = "name_origin"
	colPetMeetingStory    = "meeting_story"
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
	NeuteredStatus  string
	AcquisitionType string
	DangerLevel     string
	DangerReason    *string
	Food            string
	Environment     string
	Phone           string
	InsuranceID     *uint64
	Remarks         string
	// EMR-174: 名前の由来 / 出逢いのストーリーは任意記録。nil = 未記録(NULL)。
	NameOrigin   *string
	MeetingStory *string
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
	NeuteredStatus  *string
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
	// EMR-174: DangerReason と同型 tri-state（nil=未指定 / &nil=NULLクリア / &&value=更新対象）。
	NameOrigin   **string
	MeetingStory **string
	ActorID      *uint64
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
	if input.NeuteredStatus != nil {
		fields["neutered_status"] = *input.NeuteredStatus
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
	if input.NameOrigin != nil {
		// *input.NameOrigin は *string: nil = NULL クリア、非nil = 値セット
		fields[colPetNameOrigin] = *input.NameOrigin
	}
	if input.MeetingStory != nil {
		fields[colPetMeetingStory] = *input.MeetingStory
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
	repo                ServiceRepository
	ownerRepo           OwnerFinder
	insuranceRepo       InsuranceFinder
	medicalRecordRepo   MedicalRecordReader
	tagSyncSvc          PetTagSynchronizer
	petOwnerRepo        PetOwnerReader
	petOwnerTx          PetOwnerTransactor
	reservationTransfer UpcomingReservationOwnerTransfer
	transferAudit       PetOwnerAuditLogger
}

// NewService constructs the pet use-case service without transaction-dependent
// owner changes or deletion. Production uses NewServiceWithPetOwnerReader.
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
// guard and transaction required by production owner changes and deletion.
func NewServiceWithPetOwnerReader(
	repo ServiceRepository,
	ownerRepo OwnerFinder,
	insuranceRepo InsuranceFinder,
	medicalRecordRepo MedicalRecordReader,
	tagSyncSvc PetTagSynchronizer,
	petOwnerRepo PetOwnerReader,
	petOwnerTx PetOwnerTransactor,
	dependencies ...PetOwnerTransferDependencies,
) Service {
	svc := newService(repo, ownerRepo, insuranceRepo, medicalRecordRepo, tagSyncSvc, petOwnerRepo, petOwnerTx)
	if len(dependencies) > 0 {
		svc.reservationTransfer = dependencies[0].Reservations
		svc.transferAudit = dependencies[0].Audit
	}
	return svc
}

func newService(
	repo ServiceRepository,
	ownerRepo OwnerFinder,
	insuranceRepo InsuranceFinder,
	medicalRecordRepo MedicalRecordReader,
	tagSyncSvc PetTagSynchronizer,
	petOwnerRepo PetOwnerReader,
	petOwnerTx PetOwnerTransactor,
) *petService {
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
		return nil, 0, apperrors.Wrap(err, "failed to list pets")
	}
	return pets, total, nil
}

func (s *petService) ListOwnerReportPets(ctx context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error) {
	pets, err := s.repo.FindOwnerReportPets(ctx, clinicIDs, ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list owner report pets")
	}
	return pets, nil
}

func (s *petService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
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
		return nil, apperrors.Wrap(err, "failed to get pet first visit date")
	}
	return date, nil
}

func (s *petService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error) {
	pet, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get pet for clinics")
	}
	return pet, nil
}

// Create は既存飼主へのペット直接登録の writer が未接続の間、検証・owner 参照・書き込みの前に固定 422（F6）で拒否する。
// ペットの基本情報は pets の列ではなく共通 basic revision に書くため、subject・binding・revision・operation receipt を
// 伴う writer を接続するまで pets 行も作らない（孤児 pet は以降の読み取りをすべて unresolved にする）。
func (s *petService) Create(_ context.Context, _ uint64, _ *CreatePetInput) (*model.Pet, error) {
	return nil, errPetRegistrationPending()
}

func (s *petService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePetInput) (*model.Pet, error) {
	// 基本情報の指定は pets から廃止された列への書き込みになる。revision の writer が接続されるまで、
	// 行の読み取りや検証より前に固定 422（F4）で拒否する（HTTP 境界の key 判定に対する二重防御）。
	if updateInputTouchesBasic(input) {
		return nil, errPetBasicRevisionPending()
	}
	currentPet, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
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
	var previousOwnerID uint64
	updateAndVerify := func(updateCtx context.Context) error {
		if ownerUpdateRequested {
			locker, ok := s.repo.(RowLockingRepository)
			if !ok {
				return apperrors.WrapInternalServerError("pet owner change requires row locking")
			}
			previous, lockErr := locker.LockByIDForUpdate(updateCtx, clinicID, id)
			if lockErr != nil {
				return lockErr
			}
			if previous == nil {
				return apperrors.WrapInternalServerError("pet lock returned no result")
			}
			previousOwnerID = previous.OwnerID
		}
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
		if previousOwnerID == *input.OwnerID {
			return nil
		}
		return s.transferUpcomingReservations(updateCtx, clinicID, id, previousOwnerID, *input.OwnerID, input.ActorID)
	}
	if ownerUpdateRequested {
		err = s.petOwnerTx.WithTx(ctx, updateAndVerify)
	} else {
		err = updateAndVerify(ctx)
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update pet")
	}

	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))

	s.syncLstepTags(ctx, clinicID, pet.OwnerID, previousOwnerID)
	return pet, nil
}

func (s *petService) transferUpcomingReservations(
	ctx context.Context,
	clinicID, petID, previousOwnerID, newOwnerID uint64,
	actorID *uint64,
) error {
	if s.reservationTransfer == nil || s.transferAudit == nil {
		return apperrors.WrapInternalServerError("pet owner transfer dependencies are unavailable")
	}
	now := time.Now().In(config.JST)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
	transferredIDs, err := s.reservationTransfer.TransferUpcomingPetReservations(
		ctx, clinicID, petID, previousOwnerID, newOwnerID, from)
	if err != nil {
		return apperrors.Wrap(err, "failed to transfer upcoming reservations")
	}
	entry := &audit.Entry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     model.AuditActionPetOwnerTransfer,
		Resource:   "pet",
		ResourceID: &petID,
		OldValue:   map[string]any{"owner_id": previousOwnerID},
		NewValue:   map[string]any{"owner_id": newOwnerID},
		Metadata: map[string]any{
			"appointment_ids": transferredIDs,
			"from_date":       from.Format("2006-01-02"),
		},
	}
	if err := s.transferAudit.LogEntryTx(ctx, entry); err != nil {
		return apperrors.Wrap(err, "failed to audit pet owner transfer")
	}
	return nil
}

const petTagSyncTimeout = 30 * time.Second

func (s *petService) syncLstepTags(ctx context.Context, clinicID uint64, ownerIDs ...uint64) {
	if s.tagSyncSvc == nil {
		return
	}
	// Keep trace values, but not request cancellation or a completed transaction.
	// Every tag kind and both transfer owners share one post-commit budget.
	ctx, cancel := context.WithTimeout(persistence.DetachTx(ctx), petTagSyncTimeout)
	defer cancel()
	seen := make(map[uint64]struct{}, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		if ownerID == 0 {
			continue
		}
		if _, duplicate := seen[ownerID]; duplicate {
			continue
		}
		seen[ownerID] = struct{}{}
		if err := s.tagSyncSvc.SyncOwnerAnimalClassificationTags(ctx, clinicID, ownerID); err != nil {
			slog.ErrorContext(ctx, "failed to sync animal classification tags after pet change", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
		}
		if err := s.tagSyncSvc.SyncPetBasicInfoTags(ctx, clinicID, ownerID); err != nil {
			slog.ErrorContext(ctx, "failed to sync pet basic info tags after pet change", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
		}
		if err := s.tagSyncSvc.SyncNextVisitTag(ctx, clinicID, ownerID); err != nil {
			slog.ErrorContext(ctx, "failed to sync next visit tag after pet change", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
		}
	}
}

func (s *petService) Delete(ctx context.Context, clinicID, id uint64) error {
	locker, ok := s.repo.(RowLockingRepository)
	if !ok || s.petOwnerTx == nil {
		return apperrors.WrapInternalServerError("pet deletion requires row locking and a transaction")
	}
	var ownerID uint64
	if err := s.petOwnerTx.WithTx(ctx, func(txCtx context.Context) error {
		pet, err := locker.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock pet")
		}
		if pet == nil {
			return apperrors.WrapInternalServerError("pet lock returned no result")
		}
		ownerID = pet.OwnerID
		// FK依存チェック: カルテが紐付いている場合は削除を拒否
		recordCount, err := s.medicalRecordRepo.CountByPetID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check medical record dependencies")
		}
		if recordCount > 0 {
			return apperrors.WrapConflict("カルテが紐付いているため削除できません。先にカルテを削除またはこのペットを変更してください")
		}
		return s.repo.Delete(txCtx, clinicID, id)
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete pet")
	}
	// Deleted pets no longer contribute tentative next visits. Sync only after
	// commit, using the owner fixed by the same row lock as owner transfers.
	if s.tagSyncSvc != nil {
		syncCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), petTagSyncTimeout)
		err := s.tagSyncSvc.SyncNextVisitTag(syncCtx, clinicID, ownerID)
		cancel()
		if err != nil {
			slog.ErrorContext(ctx, "failed to sync next visit tag after pet deletion", "error", err, "owner_id", ownerID, "clinic_id", clinicID)
		}
	}
	slog.InfoContext(ctx, "pet deleted",
		slog.Uint64("pet_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
