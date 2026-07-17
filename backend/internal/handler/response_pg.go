package handler

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

// classifyPgError は PostgreSQL エラーコードに基づいてユーザー向けメッセージを返す
func classifyPgError(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return "参照先が存在しません"
		case "23505": // unique_violation
			return "既に登録されています"
		case "22003": // numeric_value_out_of_range
			return "数値が範囲外です"
		case "22P02": // invalid_text_representation
			return "入力値の形式が正しくありません"
		case "23514": // check_violation
			if msg, ok := checkConstraintMessages[pgErr.ConstraintName]; ok {
				return msg
			}
			return "入力値が制約条件を満たしていません"
		}
	}
	return "入力値が正しくありません"
}
