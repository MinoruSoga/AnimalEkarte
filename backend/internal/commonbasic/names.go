package commonbasic

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// ReadOwnerNames は clinicID 配下の owner の現行 basic revision の氏名を返す（表示専用の呼び出し向け）。
// R-1: 1 件でも解決できなければ全体を unresolved（422）で失敗させる。ownerIDs の重複は 1 件に畳む。
// ローカルの owners 行は引かない（binding が残っていれば soft-delete 済みでも氏名を返す）。
func ReadOwnerNames(ctx context.Context, db *gorm.DB, clinicID uint64, ownerIDs []uint64) (map[uint64]string, error) {
	names := make(map[uint64]string, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return names, nil
	}
	owners := make([]*model.Owner, 0, len(ownerIDs))
	seen := make(map[uint64]struct{}, len(ownerIDs))
	for _, id := range ownerIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		owners = append(owners, &model.Owner{ID: id, ClinicID: clinicID})
	}
	if err := HydrateOwners(ctx, db, []uint64{clinicID}, owners); err != nil {
		return nil, err
	}
	for _, owner := range owners {
		names[owner.ID] = owner.Basic.Name
	}
	return names, nil
}
