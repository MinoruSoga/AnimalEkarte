package httpapi

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isPgError はエラーチェーンに pgconn.PgError が含まれるか判定する
func isPgError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

// checkConstraintMessages は check_violation (23514) の制約名から
// ユーザー向けの具体的なメッセージへのマッピング。制約追加時はここに追記する。
var checkConstraintMessages = map[string]string{
	// chk_care_plan_item_ref (001_init.sql): 投薬=medicine_id・処置検査=procedure_id・
	// 持ち物=hospitalization_plan_id が必須。BUG-403: FE が該当マスタを未選択のまま
	// 送信すると発生する。
	"chk_care_plan_item_ref": "選択した種別に応じたマスタ(投薬=薬剤・処置検査=処置・持ち物=入院プラン)を選択してください",
}

// classifyPgError は PostgreSQL エラーコードに基づいてユーザー向けメッセージを返す。
//
// 第2戻り値 known は「そのエラーがクライアント入力に起因すると断定できる既知コードか」
// を表す。known=false の場合、呼び出し側はメッセージを使ってはならず、500 として扱う。
//
// BUG-2026-07-27-01: 以前は未知コードもすべて "入力値が正しくありません" を返し、
// ResolveErrorResponse が一律 400 にしていた。そのため 42703 (undefined_column) —
// GORM model と稼働 DB スキーマの乖離という純然たるサーバ側欠陥 — が
// 「利用者の入力ミス」として表示され、ペット一覧全滅の原因特定が遅れた。
// 既知コードだけを 400 とし、それ以外はサーバ欠陥として 500 に落とす。
func classifyPgError(err error) (message string, known bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return "", false
	}
	switch pgErr.Code {
	case "23503": // foreign_key_violation
		return "参照先が存在しません", true
	case "23505": // unique_violation
		return "既に登録されています", true
	case "22003": // numeric_value_out_of_range
		return "数値が範囲外です", true
	case "22P02": // invalid_text_representation
		return "入力値の形式が正しくありません", true
	case "23514": // check_violation
		if msg, ok := checkConstraintMessages[pgErr.ConstraintName]; ok {
			return msg, true
		}
		return "入力値が制約条件を満たしていません", true
	}
	return "", false
}
