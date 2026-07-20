package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑤ Batch A: dailyrecord subpackage を internal/medicalrecord へ roll-up。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type DailyRecordRepository = medicalrecord.DailyRecordRepository

// NewDailyRecordRepository constructs the daily record repository.
func NewDailyRecordRepository(db *gorm.DB) DailyRecordRepository {
	return medicalrecord.NewDailyRecordRepository(db)
}
