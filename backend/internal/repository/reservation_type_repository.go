package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/reservationtype"
)

// ReservationTypeRepository is a stable facade alias for reservationtype.
type ReservationTypeRepository = reservationtype.Repository

// NewReservationTypeRepository constructs the reservation type master repository.
func NewReservationTypeRepository(db *gorm.DB) ReservationTypeRepository {
	return reservationtype.New(db)
}
