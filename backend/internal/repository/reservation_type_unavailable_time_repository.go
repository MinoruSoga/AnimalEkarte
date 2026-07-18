package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/reservationtypeunavailabletime"
)

// ReservationTypeUnavailableTimeRepository is a stable facade alias for the
// reservationtypeunavailabletime domain package (BE8-4). Service/handler imports keep
// using repository.* so the split does not churn all importers.
type ReservationTypeUnavailableTimeRepository = reservationtypeunavailabletime.Repository

// NewReservationTypeUnavailableTimeRepository constructs the reservation type unavailable time repository.
func NewReservationTypeUnavailableTimeRepository(db *gorm.DB) ReservationTypeUnavailableTimeRepository {
	return reservationtypeunavailabletime.New(db)
}
