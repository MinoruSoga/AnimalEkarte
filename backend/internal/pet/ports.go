package pet

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

// OwnerFinder is the minimum owner capability needed to validate pet writes.
type OwnerFinder interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

// InsuranceFinder is the minimum insurance capability needed to validate an optional
// clinic-owned insurance foreign key.
type InsuranceFinder interface {
	FindByID(ctx context.Context, clinicID, insuranceID uint64) (*model.Insurance, error)
}

// MedicalRecordReader supplies derived first-visit data and delete dependency checks.
type MedicalRecordReader interface {
	FindFirstVisitDateByPetID(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
}

// PetTagSynchronizer is the best-effort LSTEP tag capability used after pet changes.
type PetTagSynchronizer interface {
	SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error
}

// ChronicConditionTagSynchronizer is the best-effort LSTEP capability used after
// chronic-condition changes.
type ChronicConditionTagSynchronizer interface {
	SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error
}

// PetFinder is the minimum pet read capability needed by chronic-condition use cases.
type PetFinder interface {
	FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
}

// PetOwnerReader is the minimum secondary-owner capability needed to prevent
// a secondary owner from being promoted without an explicit unlink.
type PetOwnerReader interface {
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetOwner, error)
}

// PetOwnerTransactor owns the transaction shared by replacement and audit.
type PetOwnerTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// PetOwnerAuditLogger writes a fail-closed audit entry through the ambient tx.
type PetOwnerAuditLogger interface {
	LogEntryTx(ctx context.Context, input *audit.Entry) error
}

// AnimalSpeciesUsageCounter protects deletion of a species that is still referenced.
type AnimalSpeciesUsageCounter interface {
	CountUsageByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error)
}

// AnimalSpeciesAuditLogger writes fail-closed audit entries for global animal_species
// mutations (POC-07 / U-X03-PET-SPECIES-AUDIT). Same shape as PetOwnerAuditLogger.
type AnimalSpeciesAuditLogger interface {
	LogEntryTx(ctx context.Context, input *audit.Entry) error
}

// AnimalSpeciesTransactor owns the transaction shared by species write and audit.
type AnimalSpeciesTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}
