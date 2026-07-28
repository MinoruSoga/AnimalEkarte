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
	"github.com/animal-ekarte/backend/internal/persistence"
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
	AcquisitionType *model.AcquisitionType
	DangerLevel     model.DangerLevel
	DangerReason    *string
	Food            string
	Environment     string
	Phone           string
	InsuranceID     *uint64
	Remarks         string
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

// CreatePetDraftFromModel copies the caller-owned model into a write draft and
// intentionally drops identity, tenant, numbering and GORM associations.
func CreatePetDraftFromModel(pet model.Pet) CreatePetDraft {
	return CreatePetDraft{
		AnimalSpeciesID: pet.AnimalSpeciesID,
		Name:            pet.Name,
		NameKana:        pet.NameKana,
		Gender:          pet.Gender,
		Status:          pet.Status,
		BirthDate:       pet.BirthDate,
		Breed:           pet.Breed,
		Color:           pet.Color,
		BloodType:       pet.BloodType,
		MicrochipNumber: pet.MicrochipNumber,
		Weight:          pet.Weight,
		NeuteredDate:    pet.NeuteredDate,
		AcquisitionType: pet.AcquisitionType,
		DangerLevel:     pet.DangerLevel,
		DangerReason:    pet.DangerReason,
		Food:            pet.Food,
		Environment:     pet.Environment,
		Phone:           pet.Phone,
		InsuranceID:     pet.InsuranceID,
		Remarks:         pet.Remarks,
	}
}

func (w *writer) Create(ctx context.Context, intent CreateIntent) (*model.Pet, error) {
	if w.db == nil {
		return nil, apperrors.WrapInternalServerError("pet writer database is not configured")
	}

	var loaded *model.Pet
	err := persistence.DBOrTx(ctx, w.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		created, err := createPetsInTransaction(txCtx, tx, OwnerRegistrationIntent{
			ClinicID: intent.ClinicID,
			OwnerID:  intent.OwnerID,
			Pets:     []CreatePetDraft{intent.Pet},
		})
		if err != nil {
			return err
		}
		loaded, err = loadPetGraph(txCtx, tx, intent.ClinicID, created[0].ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (w *writer) CreateForOwnerRegistration(
	ctx context.Context,
	intent OwnerRegistrationIntent,
) ([]model.Pet, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("owner registration pet write requires an ambient transaction")
	}
	return createPetsInTransaction(ctx, tx, intent)
}

func createPetsInTransaction(
	ctx context.Context,
	tx *gorm.DB,
	intent OwnerRegistrationIntent,
) ([]model.Pet, error) {
	if intent.ClinicID == 0 || intent.OwnerID == 0 {
		return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
	}
	if len(intent.Pets) == 0 {
		return []model.Pet{}, nil
	}

	normalizedPets, err := normalizeCreatePetDraftDangerReasons(intent.Pets)
	if err != nil {
		return nil, err
	}
	intent.Pets = normalizedPets

	db := tx.WithContext(ctx)
	if err := lockOwnerForRegistration(db, intent.ClinicID, intent.OwnerID); err != nil {
		return nil, err
	}
	if err := lockOwnerRegistrationMasters(db, intent.ClinicID, intent.Pets); err != nil {
		return nil, err
	}

	var existingCount int64
	if err := db.Model(&model.Pet{}).
		Where("clinic_id = ? AND owner_id = ? AND deleted_at IS NULL", intent.ClinicID, intent.OwnerID).
		Count(&existingCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", "")
	}

	created := make([]model.Pet, 0, len(intent.Pets))
	for i := range intent.Pets {
		petNumber := fmt.Sprintf("%d-%d", intent.OwnerID, existingCount+int64(i)+1)
		pet := intent.Pets[i].model(intent.ClinicID, intent.OwnerID, petNumber)
		if err := db.Omit(clause.Associations).Create(&pet).Error; err != nil {
			return nil, apperrors.FromGORM(err, "pet", "")
		}
		created = append(created, pet)
	}
	return created, nil
}

func normalizeCreatePetDraftDangerReasons(pets []CreatePetDraft) ([]CreatePetDraft, error) {
	normalized := make([]CreatePetDraft, len(pets))
	for i := range pets {
		normalized[i] = pets[i]
		reason, err := normalizeDangerReason(pets[i].DangerLevel, pets[i].DangerReason)
		if err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		normalized[i].DangerReason = reason
	}
	return normalized, nil
}

func normalizeDangerReason(level model.DangerLevel, reason *string) (*string, error) {
	if reason == nil {
		if level == model.DangerLevelHigh {
			return nil, apperrors.WrapInvalidInput("危険度がhighの場合は危険理由を入力してください")
		}
		return nil, nil
	}

	trimmed := strings.TrimSpace(*reason)
	if utf8.RuneCountInString(trimmed) > dangerReasonMaxRunes {
		return nil, apperrors.WrapInvalidInput("危険理由は500文字以内で入力してください")
	}
	if trimmed == "" {
		if level == model.DangerLevelHigh {
			return nil, apperrors.WrapInvalidInput("危険度がhighの場合は危険理由を入力してください")
		}
		return nil, nil
	}
	return &trimmed, nil
}

func (p CreatePetDraft) model(clinicID, ownerID uint64, petNumber string) model.Pet {
	return model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       petNumber,
		Name:            p.Name,
		NameKana:        p.NameKana,
		Gender:          p.Gender,
		Status:          p.Status,
		BirthDate:       p.BirthDate,
		Breed:           p.Breed,
		Color:           p.Color,
		BloodType:       p.BloodType,
		MicrochipNumber: p.MicrochipNumber,
		Weight:          p.Weight,
		NeuteredDate:    p.NeuteredDate,
		AcquisitionType: p.AcquisitionType,
		DangerLevel:     p.DangerLevel,
		DangerReason:    p.DangerReason,
		Food:            p.Food,
		Environment:     p.Environment,
		Phone:           p.Phone,
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
	}
}

func loadPetGraph(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, petID uint64,
) (*model.Pet, error) {
	var pet model.Pet
	err := tx.WithContext(ctx).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("AnimalSpecies").
		Preload("Insurance", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
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
