package reservation

// reservation_conflict.go — EMR-76 / BUG-RES-OVERLAP-500 + BUG-TRIM-EXCL-TIMERANGE-500.
//
// Overlapping appointment writes can pass every application-level precheck
// (CheckSlotConflict is intentionally skipped on record_shortcut/reception/exam_room
// routes and checked-in-or-later statuses, and is racy even when enabled) and hit the
// PostgreSQL exclusion constraint excl_appointments_doctor_timerange (SQLSTATE 23P01).
// apperrors.FromGORM has no 23P01 branch, so the violation used to surface as HTTP 500.
// This classifier elevates that single DB signature to a typed reservation-domain
// conflict at the repository write boundary, keeping every other error unchanged
// (fail-closed). The original *pgconn.PgError stays in the chain for logging.

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const (
	// reservationTimeConflictConstraint is the appointments doctor×timerange exclusion
	// constraint that guards against double-booking at the database level.
	reservationTimeConflictConstraint = "excl_appointments_doctor_timerange"
	// ReservationTimeConflictCode is the stable client-facing code for the 409 body.
	ReservationTimeConflictCode    = "reservation_time_conflict"
	reservationTimeConflictMessage = "既に予約が存在します"
	pgExclusionViolationCode       = "23P01"
)

// asReservationTimeConflict converts err into a typed conflict only when the chain
// contains a PostgreSQL exclusion violation on the doctor×timerange constraint.
// Returns nil for any other error (fail-closed: no unrelated failure is relabeled).
func asReservationTimeConflict(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code != pgExclusionViolationCode || pgErr.ConstraintName != reservationTimeConflictConstraint {
		return nil
	}
	return &apperrors.AppError{
		Code:    ReservationTimeConflictCode,
		Message: reservationTimeConflictMessage,
		// Wrap both the conflict sentinel (HTTP 409 mapping) and the original error
		// (which carries the PgError for logging/errors.As).
		Err: fmt.Errorf("%w: %w", apperrors.ErrConflict, err),
	}
}

// isReservationTimeConflict reports whether err is the typed reservation-time conflict
// produced by asReservationTimeConflict.
func isReservationTimeConflict(err error) bool {
	var appErr *apperrors.AppError
	return errors.As(err, &appErr) && appErr.Code == ReservationTimeConflictCode
}
