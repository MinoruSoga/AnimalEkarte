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
		}
	}
	return "入力値が正しくありません"
}
