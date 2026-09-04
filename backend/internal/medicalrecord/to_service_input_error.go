package medicalrecord

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

const invalidDateTimeFormatMessage = "日時の形式が正しくありません"

func respondToServiceInputError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		httpapi.RespondError(c, err)
		return
	}
	httpapi.RespondError(c, apperrors.WrapInvalidInput(invalidDateTimeFormatMessage))
}
