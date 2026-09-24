package reservation

// reservation_conflict_test.go — EMR-76 / BUG-RES-OVERLAP-500 + BUG-TRIM-EXCL-TIMERANGE-500:
// PostgreSQL exclusion violation 23P01 on excl_appointments_doctor_timerange must become a
// typed reservation-domain conflict (ErrConflict + reservation_time_conflict + 既に予約が存在します)
// instead of leaking as HTTP 500. Fail-closed for every other SQLSTATE/constraint.

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

func exclusionViolation() *pgconn.PgError {
	return &pgconn.PgError{
		Code:           "23P01",
		ConstraintName: "excl_appointments_doctor_timerange",
		Message:        `conflicting key value violates exclusion constraint "excl_appointments_doctor_timerange"`,
	}
}

func TestAsReservationTimeConflict_ClassifiesExclusionViolation(t *testing.T) {
	pgErr := exclusionViolation()

	got := asReservationTimeConflict(pgErr)

	require.NotNil(t, got)
	assert.True(t, errors.Is(got, apperrors.ErrConflict),
		"classified error must satisfy errors.Is(err, apperrors.ErrConflict)")
	var appErr *apperrors.AppError
	require.True(t, errors.As(got, &appErr))
	assert.Equal(t, "reservation_time_conflict", appErr.Code)
	assert.Equal(t, "既に予約が存在します", appErr.Message)
	var chained *pgconn.PgError
	require.True(t, errors.As(got, &chained), "original PgError must stay in the chain")
	assert.Same(t, pgErr, chained)
}

func TestAsReservationTimeConflict_ClassifiesWrappedForms(t *testing.T) {
	cases := map[string]error{
		"raw":              exclusionViolation(),
		"fmt_wrapped":      fmt.Errorf("create reservation: %w", exclusionViolation()),
		"service_wrapped":  apperrors.Wrap(exclusionViolation(), "failed to update reservation with conflict check"),
		"fromgorm_wrapped": apperrors.FromGORM(exclusionViolation(), "reservation", "1"),
		// retried form: a second overlapping attempt replays the same classified error shape
		"retried": apperrors.Wrap(apperrors.FromGORM(exclusionViolation(), "reservation", ""), "retry after concurrent booking"),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			got := asReservationTimeConflict(err)
			require.NotNil(t, got, "wrapped form %q must still classify", name)
			assert.True(t, errors.Is(got, apperrors.ErrConflict))
			var appErr *apperrors.AppError
			require.True(t, errors.As(got, &appErr))
			assert.Equal(t, "reservation_time_conflict", appErr.Code)
			assert.Equal(t, "既に予約が存在します", appErr.Message)
			var chained *pgconn.PgError
			require.True(t, errors.As(got, &chained))
		})
	}
}

func TestAsReservationTimeConflict_FailClosed(t *testing.T) {
	cases := map[string]error{
		"exclusion_other_constraint": &pgconn.PgError{Code: "23P01", ConstraintName: "excl_other_table_timerange"},
		"unique_violation":           &pgconn.PgError{Code: "23505", ConstraintName: "excl_appointments_doctor_timerange"},
		"unique_violation_other":     &pgconn.PgError{Code: "23505", ConstraintName: "uk_appointments_staff_time"},
		"fk_violation":               &pgconn.PgError{Code: "23503", ConstraintName: "fk_appointments_pet"},
		"check_violation":            &pgconn.PgError{Code: "23514", ConstraintName: "chk_time_range"},
		"pg_error_no_constraint":     &pgconn.PgError{Code: "23P01"},
		"non_pg_error":               errors.New("connection refused"),
		"conflict_sentinel_only":     apperrors.WrapConflict("予約が同時に更新されました。再読み込みしてください"),
		"nil_chain":                  fmt.Errorf("outer: %w", errors.New("plain")),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Nil(t, asReservationTimeConflict(err),
				"non-matching error %q must not be elevated to a reservation conflict", name)
		})
	}
	assert.Nil(t, asReservationTimeConflict(nil))
}

func TestReservationTimeConflict_HTTPMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	classified := asReservationTimeConflict(exclusionViolation())
	require.NotNil(t, classified)

	t.Run("ResolveErrorResponse raw", func(t *testing.T) {
		status, message, code := httpapi.ResolveErrorResponse(classified)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "既に予約が存在します", message)
		assert.Equal(t, "reservation_time_conflict", code)
	})

	wrappedForms := map[string]error{
		"raw":     classified,
		"wrapped": apperrors.Wrap(classified, "failed to update reservation with conflict check"),
		"retried": apperrors.Wrap(asReservationTimeConflict(exclusionViolation()), "second overlapping attempt"),
	}
	for name, err := range wrappedForms {
		t.Run("ResolveErrorResponse_"+name, func(t *testing.T) {
			status, message, code := httpapi.ResolveErrorResponse(err)
			assert.Equal(t, http.StatusConflict, status)
			assert.Equal(t, "既に予約が存在します", message)
			assert.Equal(t, "reservation_time_conflict", code)
		})
		t.Run("respondError_"+name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/reservations", http.NoBody)
			respondError(c, err)
			assert.Equal(t, http.StatusConflict, w.Code)
			assert.Contains(t, w.Body.String(), "既に予約が存在します")
			assert.Contains(t, w.Body.String(), "reservation_time_conflict")
		})
		t.Run("generic_RespondError_"+name, func(t *testing.T) {
			// trimming handler path: generic responder, error field only.
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/trimmings", http.NoBody)
			httpapi.RespondError(c, err)
			assert.Equal(t, http.StatusConflict, w.Code)
			assert.Contains(t, w.Body.String(), "既に予約が存在します")
		})
	}
}

func TestReservationTimeConflict_NonMatchingKeeps500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// fail-closed end-to-end: a 23P01 on a different constraint stays a 500.
	other := apperrors.FromGORM(
		&pgconn.PgError{Code: "23P01", ConstraintName: "excl_other"}, "reservation", "1")
	assert.Nil(t, asReservationTimeConflict(other))
	status, _, _ := httpapi.ResolveErrorResponse(other)
	assert.Equal(t, http.StatusInternalServerError, status)
}
