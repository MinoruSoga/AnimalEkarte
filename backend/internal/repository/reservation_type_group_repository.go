package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/reservationtypegroup"
)

// ReservationTypeGroupRepository is a stable facade alias for the reservationtypegroup
// domain package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type ReservationTypeGroupRepository = reservationtypegroup.Repository

// NewReservationTypeGroupRepository constructs the reservation type group repository.
func NewReservationTypeGroupRepository(db *gorm.DB) ReservationTypeGroupRepository {
	return reservationtypegroup.New(db)
}
