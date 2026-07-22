package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// These aliases preserve the exact reservation validation contracts while keeping the trimming
// implementation independent of the legacy repository aggregator.
type ReservationTypeRepository = reservation.ReservationTypeRepository
type ReservationStaffRepository = reservation.ReservationStaffRepository
type ReservationTypeUnavailableTimeRepository = reservation.ReservationTypeUnavailableTimeRepository

// Transactor is the consumer-side ambient transaction port needed by trimming writes.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
