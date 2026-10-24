package billing

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: 見積・会計の入れ子要約（toOwnerSummary / toPetSummary）は hydrate 済みの現行 basic revision を読む。
// Owner / Pet を preload する読み取り経路は、返す前にこの helper で basic revision を載せる。1 件でも解決できなければ
// 全体が unresolved（422）で、空値や廃止列へのフォールバックはしない。引数の要素は値のまま渡してよい
// （Owner / Pet はポインタなので、要素のコピーに載せた結果も元の値から見える）。

// ownerPetTargets は hydrate 対象の owner / pet を集める。nil（preload 対象外・sanitize で除去済み）は集めない。
type ownerPetTargets struct {
	owners []*model.Owner
	pets   []*model.Pet
}

func (t *ownerPetTargets) add(owner *model.Owner, pet *model.Pet) {
	if owner != nil {
		t.owners = append(t.owners, owner)
	}
	if pet != nil {
		t.pets = append(t.pets, pet)
	}
}

func (t *ownerPetTargets) hydrate(ctx context.Context, db *gorm.DB, clinicIDs []uint64) error {
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, t.owners); err != nil {
		return err
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, t.pets)
}

// hydrateEstimateBasics は見積の Owner / Pet に現行 basic revision を載せる。
func hydrateEstimateBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, estimates []model.Estimate) error {
	var targets ownerPetTargets
	for i := range estimates {
		targets.add(estimates[i].Owner, estimates[i].Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}

// hydrateBillingBasics は会計の Owner / Pet に現行 basic revision を載せる（sanitizeBillingRelations の後に呼ぶ）。
func hydrateBillingBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, billings []model.Billing) error {
	var targets ownerPetTargets
	for i := range billings {
		targets.add(billings[i].Owner, billings[i].Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}
