package pet

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// PetOwnerRepository はペットと副飼主の追加紐付けを永続化する。
type PetOwnerRepository interface {
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetOwner, error)
	FindSharedPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]SharedPet, error)
	ReplaceForPet(ctx context.Context, clinicID, petID uint64, links []model.PetOwner, expectedVersion *int) error
	CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

// SharedPet is the display projection for a pet linked through pet_owners.
type SharedPet struct {
	ID                uint64
	PetNumber         string
	Name              string
	Status            *model.PetStatus
	AnimalSpeciesName string
	Gender            *model.PetGender
	BirthDate         *time.Time
	DeceasedAt        *time.Time
	Color             *string
	Weight            *model.ExactNumeric
	Environment       *string
	Remarks           *string
	Relationship      string
}

type petOwnerRepository struct {
	db *gorm.DB
}

// NewPetOwnerRepository は PetOwnerRepository を初期化して返す。
func NewPetOwnerRepository(db *gorm.DB) PetOwnerRepository {
	return &petOwnerRepository{db: db}
}

func (r *petOwnerRepository) FindByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.PetOwner, error) {
	var links []model.PetOwner
	err := persistence.DBOrTx(ctx, r.db).
		Where("pet_owners.clinic_id = ? AND pet_owners.pet_id = ?", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_owners.pet_id AND p.clinic_id = pet_owners.clinic_id)").
		Where("EXISTS (SELECT 1 FROM owners o WHERE o.id = pet_owners.owner_id AND o.clinic_id = pet_owners.clinic_id)").
		Order("pet_owners.id ASC").
		Find(&links).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_owner", "")
	}
	return links, nil
}

// sharedPetLink は pet_owners 経由で共有された pet のローカル行と続柄。基本情報は含めず、現行 basic revision から読む。
type sharedPetLink struct {
	ID           uint64 `gorm:"column:id"`
	ClinicID     uint64 `gorm:"column:clinic_id"`
	PetNumber    string `gorm:"column:pet_number"`
	Relationship string `gorm:"column:relationship"`
}

func (r *petOwnerRepository) FindSharedPetsByOwnerID(
	ctx context.Context,
	clinicID, ownerID uint64,
) ([]SharedPet, error) {
	var links []sharedPetLink
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Pet{}).
		Select("pets.id AS id, pets.clinic_id AS clinic_id, pets.pet_number AS pet_number, pet_owners.relationship AS relationship").
		Joins("JOIN pet_owners ON pet_owners.pet_id = pets.id AND pet_owners.clinic_id = pets.clinic_id").
		Where("pet_owners.clinic_id = ? AND pet_owners.owner_id = ?", clinicID, ownerID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_owners.pet_id AND p.clinic_id = pet_owners.clinic_id)").
		Where("EXISTS (SELECT 1 FROM owners o WHERE o.id = pet_owners.owner_id AND o.clinic_id = pet_owners.clinic_id)").
		Where("pets.owner_id <> pet_owners.owner_id").
		Order("pet_owners.id ASC").
		Find(&links).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet_owner", "")
	}

	// 名前・動物種・生死などは pets の列ではなく現行 basic revision にある。1 件でも未解決なら F1（422）で失敗させ、
	// 行を黙って落とさない。
	pets := make([]model.Pet, len(links))
	for i, link := range links {
		pets[i] = model.Pet{ID: link.ID, ClinicID: link.ClinicID, PetNumber: link.PetNumber}
	}
	if err := commonbasic.HydratePets(ctx, r.db, []uint64{clinicID}, commonbasic.Ptrs(pets)); err != nil {
		return nil, err
	}
	sharedPets := make([]SharedPet, 0, len(links))
	for i := range pets {
		sharedPet, buildErr := newSharedPet(&pets[i], links[i].Relationship)
		if buildErr != nil {
			return nil, buildErr
		}
		sharedPets = append(sharedPets, sharedPet)
	}
	return sharedPets, nil
}

// newSharedPet は hydrate 済みの pet を共有 pet の表示用射影へ写す。基本情報や動物種が未解決なら F1（422）。
func newSharedPet(pet *model.Pet, relationship string) (SharedPet, error) {
	basic, err := commonbasic.PetOf(pet)
	if err != nil {
		return SharedPet{}, err
	}
	if pet.AnimalSpecies == nil {
		return SharedPet{}, apperrors.WrapCommonBasicUnresolved(
			fmt.Sprintf("pet %d/%d: animal species is not hydrated", pet.ClinicID, pet.ID))
	}
	return SharedPet{
		ID:                pet.ID,
		PetNumber:         pet.PetNumber,
		Name:              basic.Name,
		Status:            basic.Status,
		AnimalSpeciesName: pet.AnimalSpecies.Name,
		Gender:            basic.Gender,
		BirthDate:         basic.BirthDate,
		DeceasedAt:        basic.DeceasedAt,
		Color:             basic.Color,
		Weight:            copyExactNumeric(basic.RegistrationWeight),
		Environment:       basic.Environment,
		Remarks:           basic.Remarks,
		Relationship:      relationship,
	}, nil
}

func (r *petOwnerRepository) ReplaceForPet(
	ctx context.Context,
	clinicID, petID uint64,
	links []model.PetOwner,
	expectedVersion *int,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	return db.Transaction(func(tx *gorm.DB) error {
		var pet model.Pet
		if err := tx.Unscoped().
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "version").
			Where("clinic_id = ? AND id = ?", clinicID, petID).
			First(&pet).Error; err != nil {
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
		}
		if expectedVersion != nil && pet.Version != *expectedVersion {
			return apperrors.WrapConflict("他のユーザーがこのペットの副飼主を変更しました。再読み込みしてください")
		}

		ownerIDs := uniqueSortedPetOwnerIDs(links)
		if len(ownerIDs) > 0 {
			var owners []model.Owner
			if err := tx.
				Clauses(clause.Locking{Strength: "SHARE"}).
				Select("id").
				Where("clinic_id = ? AND id IN ?", clinicID, ownerIDs).
				Order("id ASC").
				Find(&owners).Error; err != nil {
				return apperrors.FromGORM(err, "owner", "")
			}
			if len(owners) != len(ownerIDs) {
				return apperrors.WrapNotFound("owner", "")
			}
		}

		result := tx.
			Where("clinic_id = ? AND pet_id = ?", clinicID, petID).
			Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_owners.pet_id AND p.clinic_id = pet_owners.clinic_id)").
			Delete(&model.PetOwner{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "pet_owner", "")
		}

		if len(links) > 0 {
			replacements := make([]model.PetOwner, len(links))
			for i, link := range links {
				replacements[i] = model.PetOwner{
					ClinicID:     clinicID,
					PetID:        petID,
					OwnerID:      link.OwnerID,
					Relationship: link.Relationship,
				}
			}
			if err := tx.Create(&replacements).Error; err != nil {
				return apperrors.FromGORM(err, "pet_owner", "")
			}
		}

		versionResult := tx.Unscoped().
			Model(&model.Pet{}).
			Where("clinic_id = ? AND id = ? AND version = ?", clinicID, petID, pet.Version).
			UpdateColumn("version", pet.Version+1)
		if versionResult.Error != nil {
			return apperrors.FromGORM(versionResult.Error, "pet", fmt.Sprintf("%d", petID))
		}
		if versionResult.RowsAffected != 1 {
			return apperrors.WrapConflict("他のユーザーがこのペットの副飼主を変更しました。再読み込みしてください")
		}
		return nil
	})
}

func (r *petOwnerRepository) CountByOwnerID(
	ctx context.Context,
	clinicID, ownerID uint64,
) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.PetOwner{}).
		Where("pet_owners.clinic_id = ? AND pet_owners.owner_id = ?", clinicID, ownerID).
		Where("EXISTS (SELECT 1 FROM owners o WHERE o.id = pet_owners.owner_id AND o.clinic_id = pet_owners.clinic_id)").
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = pet_owners.pet_id AND p.clinic_id = pet_owners.clinic_id)").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet_owner", "")
	}
	return count, nil
}

func uniqueSortedPetOwnerIDs(links []model.PetOwner) []uint64 {
	seen := make(map[uint64]struct{}, len(links))
	ownerIDs := make([]uint64, 0, len(links))
	for _, link := range links {
		if _, exists := seen[link.OwnerID]; exists {
			continue
		}
		seen[link.OwnerID] = struct{}{}
		ownerIDs = append(ownerIDs, link.OwnerID)
	}
	sort.Slice(ownerIDs, func(i, j int) bool {
		return ownerIDs[i] < ownerIDs[j]
	})
	return ownerIDs
}
