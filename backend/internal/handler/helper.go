package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// parseIDParam parses a uint64 path parameter.
// Returns 0 and writes error response on failure.
func parseIDParam(c *gin.Context, paramName string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(paramName), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid %s", paramName)))
		return 0, false
	}
	return id, true
}
