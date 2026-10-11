package reservation

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClearNextVisitIntentExpectation identifies the exact revision shown in a confirmation.
// Keep UpdatedAt as RFC3339Nano at the HTTP boundary; a JavaScript Date loses precision.
type ClearNextVisitIntentExpectation struct {
	ReservationID uint64
	UpdatedAt     time.Time
}

type guardedNextVisitIntentClearer interface {
	ClearNextVisitIntentIfMatches(context.Context, uint64, uint64, ClearNextVisitIntentExpectation) error
}

func parseClearNextVisitIntentExpectation(c *gin.Context) (*ClearNextVisitIntentExpectation, error) {
	idText, hasID := c.Request.URL.Query()["expected_reservation_id"]
	timeText, hasTime := c.Request.URL.Query()["expected_updated_at"]
	if !hasID && !hasTime {
		return nil, nil
	}
	if !hasID || !hasTime || len(idText) != 1 || len(timeText) != 1 {
		return nil, apperrors.WrapInvalidInput("expected_reservation_id and expected_updated_at must be provided together")
	}
	id, err := strconv.ParseUint(idText[0], 10, 64)
	if err != nil || id == 0 {
		return nil, apperrors.WrapInvalidInput("expected_reservation_id must be a positive ID")
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, timeText[0])
	if err != nil || updatedAt.IsZero() {
		return nil, apperrors.WrapInvalidInput("expected_updated_at must be an RFC3339 timestamp")
	}
	return &ClearNextVisitIntentExpectation{ReservationID: id, UpdatedAt: updatedAt}, nil
}

func checkClearNextVisitIntentExpectation(existing *model.Reservation, expected *ClearNextVisitIntentExpectation) error {
	if expected == nil {
		return nil
	}
	if expected.ReservationID == 0 || expected.UpdatedAt.IsZero() {
		return apperrors.WrapInvalidInput("intent identity and revision are required")
	}
	if existing == nil || existing.ID != expected.ReservationID || !existing.UpdatedAt.Equal(expected.UpdatedAt) {
		return apperrors.WrapConflict("次回予定が更新されています。再読み込みしてください")
	}
	return nil
}
