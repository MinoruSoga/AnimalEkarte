package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// MasterNameMaxLength はマスタ系エンティティの name カラムに許容する
// 最大文字数 (UTF-8 rune count)。BUG-379 対応。
// MasterNameMaxLength の正本は sharedkernel（共有カーネル昇格batch）。
const MasterNameMaxLength = sharedkernel.MasterNameMaxLength

// DefaultTaxRate は消費税標準税率。nil 税率時のデフォルト。
const DefaultTaxRate = 0.10

// normalizePagination はページネーションパラメータを正規化する（C-7）。
// page<=0 は1に、perPage<=0 は defaultPerPage に、perPage>maxPerPage は
// maxPerPage に丸める。戻り値は (page, perPage, offset)。
func normalizePagination(page, perPage, defaultPerPage, maxPerPage int) (outPage, outPerPage, outOffset int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	offset := (page - 1) * perPage
	return page, perPage, offset
}

// 日本語エラーメッセージ定数 (BUG-385)
const (
	// 両 package 共有 subset の正本は sharedkernel（値の単一ソース化・共有カーネル昇格batch）。
	ErrMsgAtLeastOneField   = sharedkernel.ErrMsgAtLeastOneField
	ErrMsgIDsNotEmpty       = sharedkernel.ErrMsgIDsNotEmpty
	ErrMsgInputNotNil       = sharedkernel.ErrMsgInputNotNil
	ErrMsgPriceZeroOrMore   = sharedkernel.ErrMsgPriceZeroOrMore
	ErrMsgQuantityPositive  = sharedkernel.ErrMsgQuantityPositive
	ErrMsgWeightZeroOrMore  = "体重は0以上の値を入力してください"
	ErrMsgResourceNameEmpty = "リソース名が空です"
)

// validateOwnedMasterFK は request 由来の clinic-scoped master FK の所有権を検証する
// (R24: マスタFKガードの複数イディオムを共有ヘルパーへ統一)。
// find は対象マスタの FindByID を包んだアダプタ。nil の FK はスキップ（optional FK 規約）。
// validateOwnedMasterFK / validateOwnedMasterFKs は sharedkernel への既存呼び出し面互換 delegate
// （実装正本は sharedkernel・共有カーネル昇格batch。呼び出し側の直参照切替=各 domain 移行時）。
func validateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	return sharedkernel.ValidateOwnedMasterFK(ctx, entity, clinicID, id, find)
}

// validateOwnedMasterFKs は validateOwnedMasterFK の slice 版（campaign の
// merchandise_item_ids 等、request 由来の clinic-scoped master FK 一覧の所有権検証）。
func validateOwnedMasterFKs(ctx context.Context, entity string, clinicID uint64, ids []uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	return sharedkernel.ValidateOwnedMasterFKs(ctx, entity, clinicID, ids, find)
}
