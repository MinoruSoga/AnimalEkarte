package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationTypeLiffRepository は internal/reservation への移行facade（BE9-2C R①・BE9-2F削除予定）。
type ReservationTypeLiffRepository = reservation.ReservationTypeLiffRepository

// NewReservationTypeLiffRepository は internal/reservation の実装を返す（BE9-2C R① facade）。
func NewReservationTypeLiffRepository(db *gorm.DB) ReservationTypeLiffRepository {
	return reservation.NewReservationTypeLiffRepository(db)
}
