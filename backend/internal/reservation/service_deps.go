package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// consumer-side narrow views（ADR-006/BE9-2C R①）: reservation が必要とする他domain・
// 未移行repoの最小メソッド集合。実装は cmd/api/main.go（または service 集約）が注入する。

// occupationFinder は職種マスタ（staff domain）の所有権確認に使う最小view。
type occupationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
}

// reservationTypeUsageChecker は予約区分の使用中判定に使う最小view（未移行のreservation query repo）。
type reservationTypeUsageChecker interface {
	ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
}
