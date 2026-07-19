package handler

import (
	"errors"
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/service"
)

// DEPRECATED facade (BE9-2B): the generic classification logic moved to
// internal/httpapi.RespondError/RespondErrorWithExtras/ResolveErrorResponse (BE9-2A
// target:httpapi, ADR-006) so domain packages like internal/manualarticle can use it without
// importing internal/handler. internal/httpapi must never import internal/service (it is the
// topologically-first node in ADR-006's permitted dependency graph), so the one remaining
// domain-specific fallback this package's ~269 call sites still rely on —
// *service.ReservationLimitError (liff/reservation only, see liff_handler.go) — is kept here
// and checked before delegating. Delete this residual once the reservation domain migrates
// (BE9-2C/2D) and ReservationLimitError moves with it.

// reservationLimitErrorResponse handles the *service.ReservationLimitError fallback that
// predates httpapi's generic classifier. ok=false means "not a ReservationLimitError;
// delegate to httpapi.RespondError/RespondErrorWithExtras".
func reservationLimitErrorResponse(err error) (status int, message, code string, ok bool) {
	var rle *service.ReservationLimitError
	if !errors.As(err, &rle) {
		return 0, "", "", false
	}
	message = rle.Message
	if message == "" {
		message = err.Error()
	}
	return http.StatusConflict, message, rle.Code, true
}

// RespondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
// 内部エラー(5xx)は details を露出しない。
func RespondError(c *gin.Context, err error) {
	if status, message, _, ok := reservationLimitErrorResponse(err); ok {
		c.JSON(status, gin.H{"error": message})
		return
	}
	httpapi.RespondError(c, err)
}

// RespondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
func RespondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	if status, message, code, ok := reservationLimitErrorResponse(err); ok {
		response := gin.H{"error": message, "code": code}
		maps.Copy(response, extras)
		c.JSON(status, response)
		return
	}
	httpapi.RespondErrorWithExtras(c, err, extras)
}
