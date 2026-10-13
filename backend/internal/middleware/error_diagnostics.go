package middleware

// error_diagnostics.go — EMR-336: safe error diagnostics for the request log.
//
// POST /api/v1/reservations が 5xx で落ちたときだけ、既存の request log 1 行へ
// error_type / sqlstate / constraint_name を載せる。生の error 文字列は出さない。
// SQLSTATE と constraint name は bounded な形式を持つ値だけ受理し、malformed な値は
// ログへ出さない (任意テキスト注入を防ぐ)。エラーチェーンは変更しない。

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// reservationCreateRoute は POST /api/v1/reservations の登録ルートパターン。
// batch / admin / LIFF の兄弟経路は対象外 (EMR-336 scope)。
const reservationCreateRoute = "/api/v1/reservations"

// isReservationCreateServerError はこの request が安全診断属性の対象か判定する。
// method と登録ルートと status がすべて揃ったときだけ true。
func isReservationCreateServerError(c *gin.Context) bool {
	return c.Request.Method == http.MethodPost &&
		c.FullPath() == reservationCreateRoute &&
		c.Writer.Status() >= http.StatusInternalServerError
}

// reservationCreateErrorDiagnostics は error chain から bounded な診断属性だけを作る。
// *pgconn.PgError が見つかればその型名と検証済みの SQLSTATE/constraint name、
// そうでなければ最外側 error の型名だけを返す。Message/Detail/Hint/Where や
// 生の error 文字列は含めない。
func reservationCreateErrorDiagnostics(err error) []slog.Attr {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr != nil {
		attrs := []slog.Attr{slog.String("error_type", fmt.Sprintf("%T", pgErr))}
		if isSQLState(pgErr.Code) {
			attrs = append(attrs, slog.String("sqlstate", pgErr.Code))
		}
		if isDBIdentifier(pgErr.ConstraintName) {
			attrs = append(attrs, slog.String("constraint_name", pgErr.ConstraintName))
		}
		return attrs
	}
	if err == nil {
		return nil
	}
	return []slog.Attr{slog.String("error_type", fmt.Sprintf("%T", err))}
}

// isSQLState は PostgreSQL SQLSTATE の形 (ASCII 大文字英数字ちょうど 5 文字)
// だけを受理する。それ以外の値はログへ出さない。
func isSQLState(s string) bool {
	if len(s) != 5 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// isDBIdentifier は単純な DB 識別子 (ASCII 英数字と underscore、63 bytes 以下 —
// PostgreSQL NAMEDATALEN-1) だけを受理する。それ以外の値はログへ出さない。
func isDBIdentifier(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}
