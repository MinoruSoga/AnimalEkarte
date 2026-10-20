package medicalrecord

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

func respondHospitalizationCageWriteError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) && errors.Is(err, apperrors.ErrConflict) {
		switch appErr.Code {
		case codeCageAssignmentChanged, codeCageOccupied, codeOccupancyUnknown:
			httpapi.RespondErrorWithExtras(c, err, nil)
			return
		}
	}
	httpapi.RespondError(c, err)
}
