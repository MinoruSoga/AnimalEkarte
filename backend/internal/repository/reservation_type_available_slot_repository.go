package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/reservationtypeavailableslot"
)

// ReservationTypeAvailableSlotRepository is a stable facade alias for the
// reservationtypeavailableslot domain package (BE8-4). Service/handler imports keep
// using repository.* so the split does not churn all importers.
type ReservationTypeAvailableSlotRepository = reservationtypeavailableslot.Repository

// NewReservationTypeAvailableSlotRepository constructs the reservation type available slot repository.
func NewReservationTypeAvailableSlotRepository(db *gorm.DB) ReservationTypeAvailableSlotRepository {
	return reservationtypeavailableslot.New(db)
}
