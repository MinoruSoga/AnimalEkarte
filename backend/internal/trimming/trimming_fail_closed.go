package trimming

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// respondTrimmingError は fail-closed の安定コード付き 422（trimming_booking_pending /
// trimming_master_usage_unverified）を error + code で返し、それ以外は httpapi.RespondError と同じ応答にする。
// httpapi.RespondError は code を出さないため、FE が機械判定できる安定コードをここで通す。
func respondTrimmingError(c *gin.Context, err error) {
	if apperrors.HasCode(err, apperrors.CodeTrimmingBookingPending) ||
		apperrors.HasCode(err, apperrors.CodeTrimmingMasterUsageUnverified) {
		httpapi.RespondErrorWithExtras(c, err, nil)
		return
	}
	httpapi.RespondError(c, err)
}
