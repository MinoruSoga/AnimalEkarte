package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationAdminRepository は internal/reservation への移行facade（BE9-2C R④・BE9-2F削除予定）。
type ReservationAdminRepository = reservation.ReservationAdminRepository

// NewReservationAdminRepository は internal/reservation の実装を返す（BE9-2C R④ facade）。
func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return reservation.NewReservationAdminRepository(db)
}
