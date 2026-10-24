package pet

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: pet / owner の名前・動物種・生死などの基本情報は pets / owners の列ではなく現行 basic revision にある。
// 一覧の検索・並び順・生存の絞り込みは、次の別名で結んだ revision の列を参照する。

const (
	// petBasicAlias は pet の現行 basic revision の JOIN 別名。
	petBasicAlias = "pet_basic"
	// ownerBasicAlias は一覧の並び順に使う owner の現行 basic revision の JOIN 別名。
	ownerBasicAlias = "owner_basic"
	// searchedOwnerBasicAlias は飼主検索の副問い合わせで結ぶ owner の現行 basic revision の JOIN 別名。
	searchedOwnerBasicAlias = "searched_owner_basic"
)

// petLivingPredicate は revisionAlias の現行 basic revision で生存と確認できる pet だけを通す述語。
// プレースホルダ 1 つに model.PetStatusAlive を渡す。生存は status='alive' かつ deceased_at IS NULL（D6）で、
// status が NULL の pet・LEFT JOIN で revision が無い行は生存として扱わない。
// revisionAlias はコンパイル時定数だけを渡すこと（SQL 文字列へ直接埋め込む）。
func petLivingPredicate(revisionAlias string) string {
	return revisionAlias + ".status = ? AND " + revisionAlias + ".deceased_at IS NULL"
}

// hydratePetBasics は pets と、preload 済みの Owner へ現行 basic revision を載せる。動物種は HydratePets が
// revision の animal_species_id から引く。1 件でも解決できなければ全体が unresolved（F1）で、空値や廃止列への
// フォールバックはしない。Owner が nil の pet（未 preload・clinic 不一致で除去済み）は owner を hydrate しない
// （owner の不在は基本情報の未解決ではない）。clinicIDs は pets / owners を読んだ認可集合と同じものを渡す。
func hydratePetBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, pets []*model.Pet) error {
	owners := make([]*model.Owner, 0, len(pets))
	for _, p := range pets {
		if p != nil && p.Owner != nil {
			owners = append(owners, p.Owner)
		}
	}
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, owners); err != nil {
		return err
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, pets)
}
