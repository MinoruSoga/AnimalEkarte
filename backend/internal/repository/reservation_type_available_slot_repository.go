package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationTypeAvailableSlotRepository は internal/reservation への移行facade（BE9-2C R①・BE9-2F削除予定）。
type ReservationTypeAvailableSlotRepository = reservation.ReservationTypeAvailableSlotRepository

// NewReservationTypeAvailableSlotRepository は internal/reservation の実装を返す（BE9-2C R① facade）。
func NewReservationTypeAvailableSlotRepository(db *gorm.DB) ReservationTypeAvailableSlotRepository {
	return reservation.NewReservationTypeAvailableSlotRepository(db)
}
