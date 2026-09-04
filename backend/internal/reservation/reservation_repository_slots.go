package reservation

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (r *reservationRepository) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", reservationTypeID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) FindClinicIDsByStaffID(
	ctx context.Context,
	clinicIDs []uint64,
	staffID uint64,
) ([]uint64, error) {
	result := make([]uint64, 0)
	if len(clinicIDs) == 0 {
		return result, nil
	}

	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Distinct().
		Order("clinic_id ASC").
		Pluck("clinic_id", &result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	return result, nil
}

// CountMedicalRecordsByReservationID は予約を参照している有効カルテの件数を返す（BUG-201 / SEC-SWEEP-02-RES-B1）。
// 親 appointments の clinic 相関で cross-tenant 親を除外する一方、medical_records.clinic_id は
// フィルタしない（参照が存在する限り削除・identity 変更ガードを fail-closed に保つ — BILL-B1b と同型）。
// 親 appointments.deleted_at は入れない（MR-B1 / TRIM-B1 と同じく clinic 相関のみ）。
// Delete / UpdateForTrimming / DeleteForTrimming の依存チェックと同じ ambient transaction へ参加する。
func (r *reservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, clinicID, reservationID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.MedicalRecord{}).
		Joins("JOIN appointments ON appointments.id = medical_records.appointment_id AND appointments.clinic_id = ?", clinicID).
		Where("medical_records.appointment_id = ? AND medical_records.deleted_at IS NULL", reservationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// AcquireBookingLock は clinic 単位の pg_advisory_xact_lock を取得する（BE-refactor.md X-9）。
// hashtextextended で "appointments:{clinicID}" をハッシュ化した bigint をロックキーに使う。
// pg_advisory_xact_lock はトランザクションスコープのため、呼び出し元の WithTx がコミット/
// ロールバックした時点で自動解放される（明示的な unlock 不要）。dbOrTx でトランザクション
// 内の ambient tx に参加する。
func (r *reservationRepository) AcquireBookingLock(ctx context.Context, clinicID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("booking lock requires an ambient transaction")
	}
	lockKey := fmt.Sprintf("appointments:%d", clinicID)
	if err := persistence.DBOrTx(ctx, r.db).Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		lockKey,
	).Error; err != nil {
		return apperrors.Wrap(err, "failed to acquire booking lock")
	}
	return nil
}

// LockAndFindByID は FOR UPDATE で予約を行ロック取得する。
// updateWithConflictCheck のトランザクション内で使用する。
func (r *reservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return nil, err
	}
	var appt model.Reservation
	err := persistence.DBOrTx(ctx, r.db).Raw(
		`SELECT * FROM appointments WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL FOR UPDATE`,
		clinicID, id,
	).Scan(&appt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	if appt.ID == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

// HasDoctorConflict は指定医師の時間枠重複を SELECT FOR UPDATE でチェックする。
func (r *reservationRepository) HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return false, err
	}
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled', 'no_show')
		  AND start_time < ?
		  AND end_time > ?
		  AND doctor_id = ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, doctorID, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return false, apperrors.Wrap(err, "lock reservations for doctor conflict check")
	}
	return len(existing) > 0, nil
}

// CountOnDutyDoctors は当日の出勤医師数を返す。
func (r *reservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT COUNT(DISTINCT se.staff_id)
		FROM shift_entries se
		JOIN staffs s ON s.id = se.staff_id
		WHERE se.clinic_id = ?
		  AND se.date = DATE(? AT TIME ZONE 'Asia/Tokyo')
		  AND se.shift_type NOT IN ('off', 'paid_leave')
		  AND s.staff_type = 'doctor'
		  AND s.is_active = true
		  AND s.deleted_at IS NULL`,
		clinicID, date,
	).Scan(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count on-duty doctors")
	}
	return count, nil
}

// CountConflicts は時間枠の競合予約数を SELECT FOR UPDATE で返す。
func (r *reservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return 0, err
	}
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled')
		  AND start_time < ?
		  AND end_time > ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "lock reservations for capacity check")
	}
	return int64(len(existing)), nil
}
