package reservation

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: 予約の入れ子要約（toOwnerSummary / toPetSummary）は hydrate 済みの現行 basic revision を読む。
// Owner / Pet（と Pet.Owner）を preload する読み取り経路は、返す前にこの helper で basic revision を載せる。
// 1 件でも解決できなければ全体が unresolved（422）で、空値や廃止列へのフォールバックはしない。
// clinicIDs は preload と同じ認可集合を渡す。

// hydrateReservationBasics は予約の Owner / Pet（と preload 済みの Pet.Owner）に現行 basic revision を載せる。
func hydrateReservationBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, reservations []*model.Reservation) error {
	var owners []*model.Owner
	var pets []*model.Pet
	for _, reservation := range reservations {
		if reservation == nil {
			continue
		}
		if reservation.Owner != nil {
			owners = append(owners, reservation.Owner)
		}
		if reservation.Pet == nil {
			continue
		}
		pets = append(pets, reservation.Pet)
		if reservation.Pet.Owner != nil {
			owners = append(owners, reservation.Pet.Owner)
		}
	}
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, owners); err != nil {
		return err
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, pets)
}
