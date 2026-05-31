package service

// MasterNameMaxLength はマスタ系エンティティの name カラムに許容する
// 最大文字数 (UTF-8 rune count)。BUG-379 対応。
const MasterNameMaxLength = 255

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
