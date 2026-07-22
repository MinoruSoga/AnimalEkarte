package lstep

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type lifecycleOwnerRepository interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
	Update(ctx context.Context, clinicID, ownerID uint64, fields map[string]any) error
}

type lifecyclePetRepository interface {
	FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
	Update(ctx context.Context, clinicID, petID uint64, fields map[string]any) error
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
// internal/service adapts it to the shared audit service input at the composition boundary.
type LifecycleAuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
}

type lifecycleAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *LifecycleAuditEntry) error
}
