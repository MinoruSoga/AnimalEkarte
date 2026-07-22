package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// BE9-2C R③ 移行facade（internal/reservation が実装・BE9-2F削除予定）。
type (
	ReservationCRUDRepository  = reservation.ReservationCRUDRepository
	ReservationSlotRepository  = reservation.ReservationSlotRepository
	ReservationQueryRepository = reservation.ReservationQueryRepository
	ReservationRepository      = reservation.ReservationStore
)

// NewReservationRepository は internal/reservation の実装を返す（BE9-2C R③ facade）。
func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return reservation.NewReservationRepository(db)
}
