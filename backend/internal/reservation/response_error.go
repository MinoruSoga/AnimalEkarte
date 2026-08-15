package reservation

// response_error.go — BE9-2C R⑤: internal/handler/response.go に残置されていた
// *ReservationLimitError の 409 フォールバック（liff/reservation 専用・「reservation domain
// 移行時に移設して削除」と明記されていた residual）を予定通り本 package へ移設。
// 本 package の handler は httpapi.RespondError 直呼びではなく必ずこれを経由する。

import (
	"errors"
	"maps"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

func reservationLimitErrorResponse(err error) (status int, message, code string, ok bool) {
	var rle *ReservationLimitError
	if !errors.As(err, &rle) {
		return 0, "", "", false
	}
	message = rle.Message
	if message == "" {
		message = err.Error()
	}
	return http.StatusConflict, message, rle.Code, true
}

// respondError はエラーを適切なHTTPステータスコードとメッセージにマッピングして返す。
func respondError(c *gin.Context, err error) {
	if status, message, _, ok := reservationLimitErrorResponse(err); ok {
		c.JSON(status, gin.H{"error": message})
		return
	}
	httpapi.RespondError(c, err)
}

// respondErrorWithExtras は custom extra fields を含むエラーレスポンスを返す。
func respondErrorWithExtras(c *gin.Context, err error, extras map[string]any) {
	if status, message, code, ok := reservationLimitErrorResponse(err); ok {
		response := gin.H{"error": message, "code": code}
		maps.Copy(response, extras)
		c.JSON(status, response)
		return
	}
	httpapi.RespondErrorWithExtras(c, err, extras)
}
