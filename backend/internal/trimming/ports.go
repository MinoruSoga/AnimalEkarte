package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationTypeRepository is the reservation-type view consumed by trimming validation.
type ReservationTypeRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

// ReservationStaffRepository is the least-capability staff view consumed by trimming validation.
type ReservationStaffRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

// ReservationTypeUnavailableTimeRepository is the unavailable-time view consumed by trimming.
type ReservationTypeUnavailableTimeRepository interface {
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
}

// Transactor is the consumer-side ambient transaction port needed by trimming writes.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AuditEntry is trimming's consumer-side view of the shared durable audit input.
// The composition layer maps this field-for-field to service.AuditLogInput.
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
}

// AuditTxLogger persists an audit entry through the ambient transaction in ctx.
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}
