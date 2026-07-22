package repository

// BE9-2C L④ compatibility facade. The implementation and owned contracts live
// in internal/lstep; remaining L⑤ analytics consumers migrate in the next batch.

import (
	"gorm.io/gorm"

	lstepdomain "github.com/animal-ekarte/backend/internal/lstep"
)

type DeliveryTriggerLogRow = lstepdomain.DeliveryTriggerLogRow
type DeliveryStatsRow = lstepdomain.DeliveryStatsRow
type VisitConversionRow = lstepdomain.VisitConversionRow
type LstepDeliveryTriggerLogRepository = lstepdomain.LstepDeliveryTriggerLogRepository

func NewLstepDeliveryTriggerLogRepository(db *gorm.DB) LstepDeliveryTriggerLogRepository {
	return lstepdomain.NewLstepDeliveryTriggerLogRepository(db)
}
