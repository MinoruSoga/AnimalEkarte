package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type lifecycleOwnerRepository interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

// OwnerLifecycleWriter exposes only the two owner lifecycle intents consumed by LSTEP.
type OwnerLifecycleWriter interface {
	RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error
	ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error
}

type lifecycleOwnerDependency interface {
	lifecycleOwnerRepository
	OwnerLifecycleWriter
}

type lifecyclePetRepository interface {
	FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
}

// PetLifecycleWriter exposes only death/revival intents (CMD-04 / CODING_RULES:35).
// Composition must inject a typed pet lifecycle writer (RecordDeath/ClearDeath), not a
// generic map[string]any field-update API.
type PetLifecycleWriter interface {
	RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error
	ClearDeath(ctx context.Context, clinicID, petID uint64) error
}

type lifecyclePetDependency interface {
	lifecyclePetRepository
	PetLifecycleWriter
}

type lifecycleTagCacheRepository interface {
	DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
}

type lifecycleTagSyncer interface {
	SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error
}

type lifecycleOperationAuditor interface {
	LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
}

type lifecycleTagConfigRepository interface {
	FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
}

type lifecycleTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// LifecycleAuditEntry is the transaction-local audit contract owned by this domain.
// cmd/api adapts it to the shared audit service input at the composition boundary.
type LifecycleAuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
}

type LifecycleAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *LifecycleAuditEntry) error
}
