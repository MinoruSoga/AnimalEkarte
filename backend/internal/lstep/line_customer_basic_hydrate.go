package lstep

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: LINE 顧客に紐づく Owner（と preload 済みの Owner.Pets）の氏名・名前・動物種・生死は、owners / pets の列ではなく
// 現行 basic revision にある。LINE 顧客の応答と LIFF の変換関数は hydrate 済みの Basic を読むため、Owner を載せる経路は
// 返す前にここで basic revision を載せる。1 件でも解決できなければ全体が unresolved（422）で、空値にはしない。

// hydrateLineCustomerBasics は LINE 顧客の Owner と Owner.Pets に現行 basic revision を載せる。
func hydrateLineCustomerBasics(ctx context.Context, db *gorm.DB, clinicID uint64, customers []*model.LineCustomer) error {
	var owners []*model.Owner
	var pets []*model.Pet
	for _, customer := range customers {
		if customer == nil || customer.Owner == nil {
			continue
		}
		owners = append(owners, customer.Owner)
		for i := range customer.Owner.Pets {
			pets = append(pets, &customer.Owner.Pets[i])
		}
	}
	clinicIDs := []uint64{clinicID}
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, owners); err != nil {
		return err
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, pets)
}
