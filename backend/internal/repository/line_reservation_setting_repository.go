package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/reservation"
)

// LineReservationSettingRepository は internal/reservation への移行facade（BE9-2C R⑥・BE9-2F削除予定）。
type LineReservationSettingRepository = reservation.LineReservationSettingRepository

// NewLineReservationSettingRepository は internal/reservation の実装を返す（BE9-2C R⑥ facade）。
func NewLineReservationSettingRepository(db *gorm.DB) LineReservationSettingRepository {
	return reservation.NewLineReservationSettingRepository(db)
}
