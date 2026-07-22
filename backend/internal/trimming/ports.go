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
