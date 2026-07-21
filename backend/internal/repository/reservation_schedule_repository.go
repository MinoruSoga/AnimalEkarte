package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/shiftentry"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// ReservationScheduleRepository は internal/reservation への移行facade（BE9-2C R②・BE9-2F削除予定）。
type ReservationScheduleRepository = reservation.ReservationScheduleRepository

// NewReservationScheduleRepository は internal/reservation の実装を返す（shift_entries write は
// staff domain の shiftentry を ADR-006 論点#1 案A の書き込み者として注入する）。
func NewReservationScheduleRepository(db *gorm.DB) ReservationScheduleRepository {
	return reservation.NewReservationScheduleRepository(db, shiftentry.New(db))
}
