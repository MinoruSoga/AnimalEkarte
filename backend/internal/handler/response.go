package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DEPRECATED facade (BE9-2B): the generic classification logic moved to
// internal/httpapi.RespondError/RespondErrorWithExtras/ResolveErrorResponse (BE9-2A
// target:httpapi, ADR-006) so domain packages like internal/manualarticle can use it without
// importing internal/handler. internal/httpapi must never import internal/service (it is the
// topologically-first node in ADR-006's permitted dependency graph), so the one remaining
// domain-specific fallback this package's ~269 call sites still rely on —
// *reservation.ReservationLimitError（R⑤で internal/reservation/response_error.go へ移設済み） (liff/reservation only, see liff_handler.go) — is kept here
// and checked before delegating. Delete this residual once the reservation domain migrates
// (BE9-2C/2D) and ReservationLimitError moves with it.

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	httpapi.RespondError(c, err)
}

// RespondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
func RespondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	httpapi.RespondErrorWithExtras(c, err, extras)
}
