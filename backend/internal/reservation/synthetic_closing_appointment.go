package reservation

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// UnscopedDeleteSyntheticClosingReservations は合成 clinic の予約系ルート行を
// 物理削除する。billing の synthetic closing teardown から呼ばれるが、
// appointments の write owner はこの package（TestAppointmentWriteOwnerLint）
// のためここに置く。appointments は reservation_types / staffs を RESTRICT 参照
// する子なので、appointments → reservation_types の順で staffs より先に消す。
// testdb の AutoMigrate スキーマには両テーブルが無い場合があるため存在確認する。
func UnscopedDeleteSyntheticClosingReservations(ctx context.Context, db *gorm.DB, clinicID uint64) error {
	if clinicID == 0 {
		return apperrors.WrapInvalidInput("clinic id is required")
	}
	if clinicID == 1 || clinicID == 2 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("clinic_id %d is reserved", clinicID))
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}
	for _, table := range []string{"appointments", "reservation_types"} {
		var exists bool
		if err := persistence.DBOrTx(ctx, db).Raw(
			fmt.Sprintf("SELECT to_regclass('%s') IS NOT NULL", table),
		).Scan(&exists).Error; err != nil {
			return apperrors.Wrap(err, fmt.Sprintf("check %s existence", table))
		}
		if !exists {
			continue
		}
		if err := persistence.DBOrTx(ctx, db).Exec(
			fmt.Sprintf("DELETE FROM %s WHERE clinic_id = ?", table), clinicID,
		).Error; err != nil {
			return apperrors.FromGORM(err, table, fmt.Sprintf("clinic_id=%d", clinicID))
		}
	}
	return nil
}
