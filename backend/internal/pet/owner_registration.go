// Package pet owns pet lifecycle use cases and persistence capabilities.
package pet

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const dangerReasonMaxRunes = 500

// CreatePetDraft is the pet-owned immutable input for every pet create path.
// Identity, tenant, owner and pet number are deliberately absent: the pet write
// owner derives them inside its transaction.
type CreatePetDraft struct {
	AnimalSpeciesID uint64
	Name            string
	NameKana        string
	Gender          model.PetGender
	Status          model.PetStatus
	BirthDate       *time.Time
	Breed           string
	Color           string
	BloodType       *string
	MicrochipNumber *string
	Weight          *float64
	NeuteredDate    *time.Time
	NeuteredStatus  model.PetNeuteredStatus
	AcquisitionType *model.AcquisitionType
	DangerLevel     model.DangerLevel
	DangerReason    *string
	Food            string
	Environment     string
	Phone           string
	InsuranceID     *uint64
	Remarks         string
	// EMR-174: 名前の由来 / 出逢いのストーリーは任意記録（nil = NULL）。
	NameOrigin   *string
	MeetingStory *string
}

// CreateIntent describes one direct pet create. Number allocation and all
// persistence validation remain inside the pet-owned transaction.
type CreateIntent struct {
	ClinicID uint64
	OwnerID  uint64
	Pet      CreatePetDraft
}

// OwnerRegistrationIntent describes the only cross-domain pet write needed by
// owner registration. The caller owns the surrounding owner+pet transaction.
type OwnerRegistrationIntent struct {
	ClinicID uint64
	OwnerID  uint64
	Pets     []CreatePetDraft
}

// OwnerRegistrationWriter creates nested pets in an existing ambient
// transaction. It does not expose a generic pet insert operation.
type OwnerRegistrationWriter interface {
	CreateForOwnerRegistration(ctx context.Context, intent OwnerRegistrationIntent) ([]model.Pet, error)
}

// Creator owns direct pet creation, including owner locking, number allocation,
// master validation, persistence and response reload.
type Creator interface {
	Create(ctx context.Context, intent CreateIntent) (*model.Pet, error)
}

// Writer is the complete pet-create write owner used by repository facades.
type Writer interface {
	Creator
	OwnerRegistrationWriter
}

type writer struct {
	db *gorm.DB
}

// NewWriter constructs the complete pet-create write owner.
func NewWriter(db *gorm.DB) Writer {
	return &writer{db: db}
}

// NewOwnerRegistrationWriter returns the pet-owned writer used by owner
// registration orchestration.
func NewOwnerRegistrationWriter() OwnerRegistrationWriter {
	return &writer{}
}

// Create は既存飼主へのペット直接登録の writer が未接続の間、書き込みの前に固定 422（F6）で拒否する。
// 基本情報は pets の列ではなく共通 basic revision に書くため、subject・binding・revision・operation receipt を
// 伴う writer を接続するまで pets 行も作らない（孤児 pet は以降の読み取りをすべて unresolved にする）。
func (*writer) Create(_ context.Context, _ CreateIntent) (*model.Pet, error) {
	return nil, errPetRegistrationPending()
}

// CreateForOwnerRegistration は飼主とペットの同時登録の pet 側 writer。owner 側の登録 writer が未接続で
// owner.CreateWithPets が先に固定 422（F3）を返すため通常は到達しないが、到達しても書き込まず F6 で拒否する。
func (*writer) CreateForOwnerRegistration(_ context.Context, _ OwnerRegistrationIntent) ([]model.Pet, error) {
	return nil, errPetRegistrationPending()
}

func normalizeCreatePetDrafts(pets []CreatePetDraft) ([]CreatePetDraft, error) {
	normalized := make([]CreatePetDraft, len(pets))
	for i := range pets {
		normalized[i] = pets[i]
		reason, err := normalizeDangerReason(pets[i].DangerLevel, pets[i].DangerReason)
		if err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		normalized[i].DangerReason = reason
		nameOrigin, err := normalizeOptionalNarrative("名前の由来", pets[i].NameOrigin, nameOriginMaxRunes)
		if err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		normalized[i].NameOrigin = nameOrigin
		meetingStory, err := normalizeOptionalNarrative("出逢いのストーリー", pets[i].MeetingStory, meetingStoryMaxRunes)
		if err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		normalized[i].MeetingStory = meetingStory
	}
	return normalized, nil
}

func normalizeDangerReason(level model.DangerLevel, reason *string) (*string, error) {
	// EMR-231: 400 の error 本文は FE の toast にそのまま出るため、警戒表現は含めない。
	if reason == nil {
		if level == model.DangerLevelHigh {
			return nil, apperrors.WrapInvalidInput("特記レベルがhighの場合は理由を入力してください")
		}
		return nil, nil
	}

	trimmed := strings.TrimSpace(*reason)
	if utf8.RuneCountInString(trimmed) > dangerReasonMaxRunes {
		return nil, apperrors.WrapInvalidInput("理由は500文字以内で入力してください")
	}
	if trimmed == "" {
		if level == model.DangerLevelHigh {
			return nil, apperrors.WrapInvalidInput("特記レベルがhighの場合は理由を入力してください")
		}
		return nil, nil
	}
	return &trimmed, nil
}

func loadPetGraph(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, petID uint64,
) (*model.Pet, error) {
	var pet model.Pet
	err := tx.WithContext(ctx).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Insurance", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	// 現行 basic revision（pet・動物種・nested owner）を載せる。未解決なら F1（422）で、空値は返さない。
	if err := hydratePetBasics(ctx, tx, []uint64{clinicID}, []*model.Pet{&pet}); err != nil {
		return nil, err
	}
	return &pet, nil
}

func lockOwnerForRegistration(db *gorm.DB, clinicID, ownerID uint64) error {
	var owner model.Owner
	err := db.
		Select("id", "clinic_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", ownerID, clinicID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&owner).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.WrapInvalidInput("owner not found in this clinic")
	}
	if err != nil {
		return apperrors.FromGORM(err, "owner", "")
	}
	return nil
}

func lockOwnerRegistrationMasters(
	db *gorm.DB,
	clinicID uint64,
	pets []CreatePetDraft,
) error {
	speciesIDs := make([]uint64, 0, len(pets))
	insuranceIDs := make([]uint64, 0, len(pets))
	speciesSeen := make(map[uint64]struct{}, len(pets))
	insuranceSeen := make(map[uint64]struct{}, len(pets))

	for i := range pets {
		if _, ok := speciesSeen[pets[i].AnimalSpeciesID]; !ok {
			speciesSeen[pets[i].AnimalSpeciesID] = struct{}{}
			speciesIDs = append(speciesIDs, pets[i].AnimalSpeciesID)
		}
		if pets[i].InsuranceID == nil {
			continue
		}
		if _, ok := insuranceSeen[*pets[i].InsuranceID]; !ok {
			insuranceSeen[*pets[i].InsuranceID] = struct{}{}
			insuranceIDs = append(insuranceIDs, *pets[i].InsuranceID)
		}
	}
	sort.Slice(speciesIDs, func(i, j int) bool { return speciesIDs[i] < speciesIDs[j] })
	sort.Slice(insuranceIDs, func(i, j int) bool { return insuranceIDs[i] < insuranceIDs[j] })

	var species []model.AnimalSpecies
	if err := db.
		Select("id").
		Where("id IN ?", speciesIDs).
		Order("id").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Find(&species).Error; err != nil {
		return apperrors.FromGORM(err, "animal_species", "")
	}
	if len(species) != len(speciesIDs) {
		return apperrors.WrapInvalidInput("animal species not found")
	}

	if len(insuranceIDs) == 0 {
		return nil
	}
	var insurances []model.Insurance
	if err := db.
		Select("id").
		Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, insuranceIDs).
		Order("id").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Find(&insurances).Error; err != nil {
		return apperrors.FromGORM(err, "insurance", "")
	}
	if len(insurances) != len(insuranceIDs) {
		return apperrors.WrapInvalidInput("insurance not found in this clinic")
	}
	return nil
}
