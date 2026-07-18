package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/dailyrecord"
)

// DailyRecordRepository is a stable facade alias for the dailyrecord domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type DailyRecordRepository = dailyrecord.Repository

// NewDailyRecordRepository constructs the daily record repository.
func NewDailyRecordRepository(db *gorm.DB) DailyRecordRepository {
	return dailyrecord.New(db)
}
