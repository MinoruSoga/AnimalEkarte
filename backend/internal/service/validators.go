package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// MasterNameMaxLength はマスタ系エンティティの name カラムに許容する
// 最大文字数 (UTF-8 rune count)。BUG-379 対応。
const MasterNameMaxLength = 255

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
	ErrMsgAtLeastOneField   = "少なくとも1つのフィールドを指定してください"
	ErrMsgIDsNotEmpty       = "並び順のIDリストが空です"
	ErrMsgInputNotNil       = "更新内容が指定されていません"
	ErrMsgPriceZeroOrMore   = "金額は0以上を入力してください"
	ErrMsgQuantityPositive  = "数量は0より大きい値を入力してください"
	ErrMsgWeightZeroOrMore  = "体重は0以上の値を入力してください"
	ErrMsgResourceNameEmpty = "リソース名が空です"
)

// validateOwnedMasterFK は request 由来の clinic-scoped master FK の所有権を検証する
// (R24: マスタFKガードの複数イディオムを共有ヘルパーへ統一)。
// find は対象マスタの FindByID を包んだアダプタ。nil の FK はスキップ（optional FK 規約）。
func validateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	if id == nil {
		return nil
	}
	if err := find(ctx, clinicID, *id); err != nil {
		slog.ErrorContext(ctx, "failed to verify "+entity+" ownership",
			"error", err, "id", *id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify "+entity+" ownership")
	}
	return nil
}

// validateOwnedMasterFKs は validateOwnedMasterFK の slice 版（campaign の
// merchandise_item_ids 等、request 由来の clinic-scoped master FK 一覧の所有権検証）。
func validateOwnedMasterFKs(ctx context.Context, entity string, clinicID uint64, ids []uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	for i := range ids {
		if err := validateOwnedMasterFK(ctx, entity, clinicID, &ids[i], find); err != nil {
			return err
		}
	}
	return nil
}
