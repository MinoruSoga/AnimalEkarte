package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationStaffRepository は internal/reservation への移行facade（BE9-2C R②・BE9-2F削除予定）。
type ReservationStaffRepository = reservation.ReservationStaffRepository

// NewReservationStaffRepository は internal/reservation の実装を返す（staffs write は
// staff domain の StaffRepository を ADR-006 論点#1 案A の書き込み者として注入する）。
func NewReservationStaffRepository(db *gorm.DB) ReservationStaffRepository {
	return reservation.NewReservationStaffRepository(db, NewStaffRepository(db))
}
