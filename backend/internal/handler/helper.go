package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// handleError handles errors from the service layer and returns appropriate HTTP responses.
// nolint:unparam // resource is kept for future extensibility with other resource types
func (h *Handler) handleError(c *gin.Context, err error, resource, id string) {
	ctx := c.Request.Context()

	switch {
	case apperrors.IsNotFound(err):
		slog.WarnContext(ctx, "resource not found",
			slog.String("resource", resource),
			slog.String("id", id),
		)
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})

	case apperrors.IsInvalidInput(err):
		slog.WarnContext(ctx, "invalid input",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	default:
		slog.ErrorContext(ctx, "internal error",
			slog.String("error", err.Error()),
			slog.String("resource", resource),
			slog.String("id", id),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
