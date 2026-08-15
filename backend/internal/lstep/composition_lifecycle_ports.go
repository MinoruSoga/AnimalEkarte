package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type ownerLifecyclePorts struct {
	reader lifecycleOwnerRepository
	writer OwnerLifecycleWriter
}

func newOwnerLifecyclePorts(reader lifecycleOwnerRepository, writer OwnerLifecycleWriter) lifecycleOwnerDependency {
	return &ownerLifecyclePorts{reader: reader, writer: writer}
}

func (p *ownerLifecyclePorts) FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	return p.reader.FindByID(ctx, clinicID, ownerID)
}

func (p *ownerLifecyclePorts) RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error {
	return p.writer.RecordLstepOptOut(ctx, clinicID, ownerID, at, reason)
}

func (p *ownerLifecyclePorts) ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error {
	return p.writer.ClearLstepOptOut(ctx, clinicID, ownerID)
}

type petLifecyclePorts struct {
	reader lifecyclePetRepository
	writer PetLifecycleWriter
}

func newPetLifecyclePorts(reader lifecyclePetRepository, writer PetLifecycleWriter) lifecyclePetDependency {
	return &petLifecyclePorts{reader: reader, writer: writer}
}

func (p *petLifecyclePorts) FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	return p.reader.FindByID(ctx, clinicID, petID)
}

func (p *petLifecyclePorts) FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
	return p.reader.FindLivingByOwner(ctx, clinicID, ownerID)
}

func (p *petLifecyclePorts) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	return p.writer.RecordDeath(ctx, clinicID, petID, deceasedAt, reason)
}

func (p *petLifecyclePorts) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	return p.writer.ClearDeath(ctx, clinicID, petID)
}
