package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// consumer-side narrow views（ADR-006/BE9-2C B①）: billing が必要とする未移行 domain repo の
// 最小メソッド集合。具象は service 集約または cmd/api/main.go が注入する。

// merchandiseItemFinder は物販マスタ（inventory domain）の所有権確認に使う最小view。
type merchandiseItemFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
}
