package service

import "github.com/animal-ekarte/backend/internal/sharedkernel"

// MasterNameMaxLength はマスタ系エンティティの name カラムに許容する
// 最大文字数 (UTF-8 rune count)。BUG-379 対応。
// MasterNameMaxLength の正本は sharedkernel（共有カーネル昇格batch）。
const MasterNameMaxLength = sharedkernel.MasterNameMaxLength

// DefaultTaxRate は消費税標準税率。nil 税率時のデフォルト。
// DefaultTaxRate の正本は sharedkernel（⑤で medicalrecord と共有化）。
const DefaultTaxRate = sharedkernel.DefaultTaxRate

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
