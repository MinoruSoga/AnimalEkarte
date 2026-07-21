package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationTypeOccupationRepository は internal/reservation への移行facade（BE9-2C R①・BE9-2F削除予定）。
type ReservationTypeOccupationRepository = reservation.ReservationTypeOccupationRepository

// NewReservationTypeOccupationRepository は internal/reservation の実装を返す（BE9-2C R① facade）。
func NewReservationTypeOccupationRepository(db *gorm.DB) ReservationTypeOccupationRepository {
	return reservation.NewReservationTypeOccupationRepository(db)
}
