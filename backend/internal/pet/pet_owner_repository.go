package pet

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// PetOwnerRepository はペットと副飼主の追加紐付けを永続化する。
type PetOwnerRepository interface {
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetOwner, error)
	ReplaceForPet(ctx context.Context, clinicID, petID uint64, links []model.PetOwner) error
	CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
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

func (r *petOwnerRepository) ReplaceForPet(
	ctx context.Context,
	clinicID, petID uint64,
	links []model.PetOwner,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	return db.Transaction(func(tx *gorm.DB) error {
		var pet model.Pet
		if err := tx.Unscoped().
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			Where("clinic_id = ? AND id = ?", clinicID, petID).
			First(&pet).Error; err != nil {
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
		}

		ownerIDs := uniqueSortedPetOwnerIDs(links)
		if len(ownerIDs) > 0 {
			var owners []model.Owner
			if err := tx.Unscoped().
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
		if len(links) == 0 {
			return nil
		}

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
