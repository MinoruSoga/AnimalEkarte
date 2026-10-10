package reservation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreateTrimmingReservationInput は trimming orchestration が appointment を作成する際の
// typed command。reservation が appointments の唯一の persistence owner となる。
type CreateTrimmingReservationInput struct {
	ReservationTypeID uint64
	StartTime         time.Time
	EndTime           time.Time
	PetID             *uint64
	DoctorID          *uint64
	Status            model.ReservationStatus
	ReservationRoute  *string
}

// UpdateTrimmingReservationInput は trimming が変更を許される appointment 列だけを公開する。
type UpdateTrimmingReservationInput struct {
	StartTime *time.Time
	EndTime   *time.Time
	PetID     *uint64
	DoctorID  *uint64
	Status    *model.ReservationStatus
}

// CompleteForAccounting は会計完了に伴う appointment の完了化を行う。
// caller が開始した billing transaction を必須とし、DBOrTx で参加してbilling writeと
// 同時にcommit/rollbackされる。ambient transaction欠落は部分commit防止のためfail-closedにする。
func (r *reservationRepository) CompleteForAccounting(
	ctx context.Context,
	clinicID uint64,
	medicalRecordID, ownerID, petID *uint64,
	scheduledDate time.Time,
) (int64, error) {
	if persistence.TxFromContext(ctx) == nil {
		return 0, apperrors.WrapInternalServerError("accounting appointment completion requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	var totalAffected int64

	// (1) 同日同一ペットの会計待ち予約を完了化する。
	if ownerID != nil && petID != nil && !scheduledDate.IsZero() {
		result := db.Model(&model.Reservation{}).
			Where("clinic_id = ? AND owner_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
				clinicID, *ownerID, *petID, model.ReservationStatusAccounting).
			Where("DATE(start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", scheduledDate).
			Update("status", model.ReservationStatusCompleted)
		if result.Error != nil {
			return totalAffected, apperrors.FromGORM(
				result.Error,
				"reservation",
				fmt.Sprintf("clinic=%d owner=%d pet=%d scheduled_date=%s", clinicID, *ownerID, *petID, scheduledDate.Format(time.DateOnly)),
			)
		}
		totalAffected += result.RowsAffected
	}

	// (2) billing.medical_record_id が指す診察 appointment を完了化する。
	// terminal status は上書きせず、medical_record と appointment の双方を clinic scope する。
	if medicalRecordID != nil {
		result := db.Model(&model.Reservation{}).
			Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
			Where("status NOT IN ?", []model.ReservationStatus{
				model.ReservationStatusCompleted,
				model.ReservationStatusCancelled,
				model.ReservationStatusNoShow,
			}).
			Where("id IN (?)",
				db.Model(&model.MedicalRecord{}).
					Select("appointment_id").
					Where("id = ? AND clinic_id = ? AND appointment_id IS NOT NULL AND deleted_at IS NULL", *medicalRecordID, clinicID)).
			Update("status", model.ReservationStatusCompleted)
		if result.Error != nil {
			return totalAffected, apperrors.FromGORM(
				result.Error,
				"reservation",
				fmt.Sprintf("clinic=%d medical_record=%d", clinicID, *medicalRecordID),
			)
		}
		totalAffected += result.RowsAffected
	}

	return totalAffected, nil
}

// BackfillForMedicalRecord は medical record 作成時に appointment の欠損 context だけを補完する。
// caller の transaction 内で行ロックを取得して再検証するため、事前read後の並行更新を任意値で
// 上書きしない。owner/pet の clinic ownership と相互整合も write 直前に検証する。
func (r *reservationRepository) BackfillForMedicalRecord(
	ctx context.Context,
	clinicID, id uint64,
	ownerID, petID, doctorID *uint64,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("medical record appointment preparation requires an ambient transaction")
	}
	current, err := r.LockAndFindByID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if current.Status == model.ReservationStatusTentative {
		// tentative 行は general placeholder 区分を持つため category チェックを素通り
		// するが、未確定の日付意図行にカルテは紐付けない（先に実予約へ昇格させる）。
		return apperrors.WrapConflict("次回予定（未確定）をカルテの対象予約にはできません。先に予約を確定してください")
	}
	if err := r.assertGeneralMedicalRecordReservationType(ctx, clinicID, current.ReservationTypeID); err != nil {
		return err
	}

	fields := make(map[string]any, 3)
	finalOwnerID, err := reservationBackfillValue("owner_id", current.OwnerID, ownerID, fields)
	if err != nil {
		return err
	}
	finalPetID, err := reservationBackfillValue("pet_id", current.PetID, petID, fields)
	if err != nil {
		return err
	}
	if err := sharedkernel.ValidateReservationOwnerPetLinks(ctx, r, clinicID, finalOwnerID, finalPetID); err != nil {
		return err
	}
	finalDoctorID, err := reservationBackfillValue("doctor_id", current.DoctorID, doctorID, fields)
	if err != nil {
		return err
	}
	if finalDoctorID != nil {
		if err := r.assertStaffAssignedToClinic(ctx, clinicID, *finalDoctorID); err != nil {
			return err
		}
	}
	if len(fields) == 0 {
		return nil
	}
	if _, err := r.update(ctx, clinicID, id, fields); err != nil {
		return err
	}
	return nil
}

func (r *reservationRepository) assertGeneralMedicalRecordReservationType(ctx context.Context, clinicID, id uint64) error {
	var reservationType model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("reservation_types.id").
		Where("id = ? AND clinic_id = ? AND category = ? AND deleted_at IS NULL", id, clinicID, model.ReservationTypeCategoryGeneral).
		Take(&reservationType).Error
	if err == nil {
		return nil
	}
	mapped := apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	if apperrors.IsNotFound(mapped) {
		return apperrors.WrapInvalidInput("appointment must use a general reservation type for a medical record")
	}
	return mapped
}

// assertStaffAssignedToClinic prevents a request-derived staff ID from linking an appointment
// to a staff member who is not currently assigned to the appointment's clinic. Staff can belong
// to multiple clinics, so staff_clinic_assignments is authoritative rather than staffs.clinic_id.
func (r *reservationRepository) assertStaffAssignedToClinic(ctx context.Context, clinicID, staffID uint64) error {
	var assignment model.StaffClinicAssignment
	db := persistence.DBOrTx(ctx, r.db).Model(&model.StaffClinicAssignment{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("staff_clinic_assignments.id").
		Joins("JOIN staffs ON staffs.id = staff_clinic_assignments.staff_id AND staffs.deleted_at IS NULL AND staffs.is_active = TRUE").
		Where("staff_clinic_assignments.staff_id = ? AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", staffID, clinicID).
		Take(&assignment).Error
	if err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	return nil
}

// AssertMedicalRecordDoctorInClinic validates a request-derived medical-record doctor against
// active staff-clinic assignment. It is intentionally intent-named so consumers do not receive a
// broad staff repository capability through the appointment owner.
func (r *reservationRepository) AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error {
	var assignment model.StaffClinicAssignment
	db := persistence.DBOrTx(ctx, r.db).Model(&model.StaffClinicAssignment{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("staff_clinic_assignments.id").
		Joins("JOIN staffs ON staffs.id = staff_clinic_assignments.staff_id AND staffs.deleted_at IS NULL AND staffs.is_active = TRUE AND staffs.staff_type = 'doctor'").
		Where("staff_clinic_assignments.staff_id = ? AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", doctorID, clinicID).
		Take(&assignment).Error
	if err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", doctorID))
	}
	return nil
}

func reservationBackfillValue(
	field string,
	current, requested *uint64,
	fields map[string]any,
) (*uint64, error) {
	if current != nil {
		if requested != nil && *requested != *current {
			return nil, apperrors.WrapInvalidInput(field + " does not match appointment")
		}
		return current, nil
	}
	if requested != nil {
		fields[field] = *requested
	}
	return requested, nil
}

// acquireAppointmentLifecycleLock serializes no-show classification with medical-record
// finalization for one clinic-scoped appointment. Both callers must already own an ambient
// transaction so the advisory lock remains held through their state change.
func (r *reservationRepository) acquireAppointmentLifecycleLock(ctx context.Context, clinicID, id uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("appointment lifecycle operation requires an ambient transaction")
	}
	lockKey := fmt.Sprintf("appointment-lifecycle:%d:%d", clinicID, id)
	if err := persistence.DBOrTx(ctx, r.db).
		Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).
		Error; err != nil {
		return apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return nil
}

// PrepareForMedicalRecordFinalization obtains the shared lifecycle lock and then locks the
// appointment row. A no-show or cancelled appointment must be corrected through the reservation
// workflow before its medical record can be finalized.
func (r *reservationRepository) PrepareForMedicalRecordFinalization(ctx context.Context, clinicID, id uint64) error {
	if err := r.acquireAppointmentLifecycleLock(ctx, clinicID, id); err != nil {
		return err
	}
	appointment, err := r.LockAndFindByID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	switch appointment.Status {
	case model.ReservationStatusNoShow, model.ReservationStatusCancelled:
		return apperrors.WrapConflict("no-showまたはキャンセル済みの予約に紐づくカルテは確定できません")
	case model.ReservationStatusTentative:
		// tentative（次回予定）は general placeholder 区分を持つため category チェックを
		// 素通りし、未確定の sentinel 時刻に対してカルテ確定できてしまう。実予約へ
		// 昇格（confirm / absorb / check-in）してからでなければカルテを紐付けない。
		return apperrors.WrapConflict("次回予定（未確定）に紐づくカルテは確定できません。先に予約を確定してください")
	default:
		return nil
	}
}

// MarkNoShow performs an atomic, idempotent no-show transition and returns the exact prior
// status for audit. A stale candidate cannot overwrite a concurrent lifecycle transition, and
// an appointment with a finalized medical record can never be classified as a no-show.
func (r *reservationRepository) MarkNoShow(ctx context.Context, clinicID, id uint64) (NoShowTransition, error) {
	return r.MarkNoShowAt(ctx, clinicID, id, time.Now().UTC())
}

// MarkNoShowAt performs the same atomic lifecycle transition while rechecking
// eligibility against the exact timestamp used by candidate selection.
func (r *reservationRepository) MarkNoShowAt(
	ctx context.Context,
	clinicID uint64,
	id uint64,
	evaluatedAt time.Time,
) (NoShowTransition, error) {
	if err := r.acquireAppointmentLifecycleLock(ctx, clinicID, id); err != nil {
		return NoShowTransition{}, err
	}
	const query = `
		WITH candidate AS (
			SELECT a.id, a.status, a.owner_id
			FROM appointments a
			WHERE a.clinic_id = ?
			  AND a.id = ?
			  AND a.deleted_at IS NULL
			  AND a.status IN (?, ?)
			  AND a.end_time <= CAST(? AS timestamptz) - interval '4 hours'
			  AND NOT EXISTS (
				SELECT 1 FROM medical_records mr
				WHERE mr.clinic_id = a.clinic_id
				  AND mr.appointment_id = a.id
				  AND mr.status = ?
				  AND mr.deleted_at IS NULL
			  )
			FOR UPDATE
		)
		UPDATE appointments AS a
		SET status = ?, updated_at = NOW()
		FROM candidate
		WHERE a.id = candidate.id
		RETURNING candidate.status, candidate.owner_id
	`

	var previousStatus model.ReservationStatus
	// Codex #25: candidate 抽出〜行ロックの間に飼い主譲渡が割り込み得るため、
	// CAS でロック確定した行の owner_id を同時に返す（NULL は 0 に潰す）。
	// sql.Null[uint64] で直接スキャンして int64→uint64 変換を排除する（G115）—
	// 負値は driver の変換境界で scan error として fail-closed に表面化する。
	var ownerID sql.Null[uint64]
	err := persistence.DBOrTx(ctx, r.db).
		Raw(
			query,
			clinicID,
			id,
			model.ReservationStatusPending,
			model.ReservationStatusConfirmed,
			evaluatedAt,
			model.MedicalRecordStatusFinalized,
			model.ReservationStatusNoShow,
		).
		Row().
		Scan(&previousStatus, &ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		return NoShowTransition{}, nil
	}
	if err != nil {
		return NoShowTransition{}, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return NoShowTransition{Changed: true, PreviousStatus: previousStatus, OwnerID: ownerID.V}, nil
}

func (r *reservationRepository) CreateForTrimming(
	ctx context.Context,
	clinicID uint64,
	input CreateTrimmingReservationInput,
) (*model.Reservation, *uint64, error) {
	if err := requireTrimmingAmbientTransaction(ctx); err != nil {
		return nil, nil, err
	}
	if err := validateTrimmingCreateStatus(input.Status); err != nil {
		return nil, nil, err
	}
	if err := r.assertActiveTrimmingReservationType(ctx, clinicID, input.ReservationTypeID); err != nil {
		return nil, nil, err
	}
	ownerID, err := r.resolveTrimmingOwnerForPet(ctx, clinicID, nil, input.PetID)
	if err != nil {
		return nil, nil, err
	}
	if input.DoctorID != nil {
		if err := r.assertStaffAssignedToClinic(ctx, clinicID, *input.DoctorID); err != nil {
			return nil, nil, err
		}
	}
	appointment := &model.Reservation{
		ClinicID:          clinicID,
		ReservationTypeID: input.ReservationTypeID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		OwnerID:           ownerID,
		PetID:             input.PetID,
		DoctorID:          input.DoctorID,
		Status:            input.Status,
		Source:            model.ReservationSourceManual,
		ReservationRoute:  input.ReservationRoute,
	}
	// EMR-308/Codex: trimming 作成も他経路と同じく tentative absorb を通す。
	// ペットに active tentative があれば新規行ではなく同一行を要求区分で昇格し、
	// なければ通常 INSERT にフォールバックする。死亡ペットは tx 内で拒否する。
	if err := ValidateReservationPetNotDeceased(ctx, r, clinicID, input.PetID); err != nil {
		return nil, nil, err
	}
	return createOrPromoteTentative(ctx, r, appointment)
}

func (r *reservationRepository) resolveTrimmingOwnerForPet(
	ctx context.Context,
	clinicID uint64,
	currentOwnerID, petID *uint64,
) (*uint64, error) {
	if petID == nil {
		return currentOwnerID, nil
	}
	petOwnerID, err := r.FindPetOwnerInClinic(ctx, clinicID, *petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to verify trimming pet ownership")
	}
	if currentOwnerID != nil && *currentOwnerID != petOwnerID {
		return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", *petID))
	}
	return &petOwnerID, nil
}

func (r *reservationRepository) assertTrimmingReservationType(ctx context.Context, clinicID, id uint64) error {
	var reservationType model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("reservation_types.id").
		Where("id = ? AND clinic_id = ? AND category = ? AND deleted_at IS NULL", id, clinicID, model.ReservationTypeCategoryTrimming).
		Take(&reservationType).Error
	if err != nil {
		return apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationRepository) assertActiveTrimmingReservationType(ctx context.Context, clinicID, id uint64) error {
	var reservationType model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("reservation_types.id").
		Where("id = ? AND clinic_id = ? AND category = ? AND is_active = true AND deleted_at IS NULL", id, clinicID, model.ReservationTypeCategoryTrimming).
		Take(&reservationType).Error
	if err != nil {
		return apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// FindTrimmingByID returns only appointments backed by a non-deleted trimming reservation type in
// the same clinic. Trimming RBAC must not authorize reads of general appointments by ID.
func (r *reservationRepository) FindTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	appointment, err := r.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if appointment.ReservationType == nil || appointment.ReservationType.Category != model.ReservationTypeCategoryTrimming {
		return nil, apperrors.WrapNotFound("trimming appointment", fmt.Sprintf("%d", id))
	}
	return appointment, nil
}

func (r *reservationRepository) lockTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if err := requireTrimmingAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	appointment, err := r.LockAndFindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := r.assertTrimmingReservationType(ctx, clinicID, appointment.ReservationTypeID); err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.WrapNotFound("trimming appointment", fmt.Sprintf("%d", id))
		}
		return nil, err
	}
	return appointment, nil
}

// LockTrimmingByID locks the appointment and its same-clinic trimming reservation type for the
// caller's transaction. The type SHARE lock keeps category/deletion stable through detail writes.
func (r *reservationRepository) LockTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return r.lockTrimmingByID(ctx, clinicID, id)
}

func requireTrimmingAmbientTransaction(ctx context.Context) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("trimming appointment write requires an ambient transaction")
	}
	return nil
}

// ValidateTrimmingAppointmentMutable rejects post-terminal mutation of appointment or trimming
// clinical detail. Any future correction workflow must be separately specified and audited.
func ValidateTrimmingAppointmentMutable(current model.ReservationStatus) error {
	switch current {
	case model.ReservationStatusCompleted, model.ReservationStatusCancelled, model.ReservationStatusNoShow:
		return apperrors.WrapConflict("完了または終了済みのトリミング記録は変更できません")
	default:
		return nil
	}
}

func validateTrimmingCreateStatus(status model.ReservationStatus) error {
	switch status {
	case model.ReservationStatusCompleted:
		return apperrors.WrapInvalidInput("completed is owned by the accounting completion transition")
	case model.ReservationStatusNoShow:
		return apperrors.WrapInvalidInput("no_show is owned by the no-show transition")
	default:
		return nil
	}
}

func validateTrimmingStatusTransition(current model.ReservationStatus, requested *model.ReservationStatus) error {
	if err := ValidateTrimmingAppointmentMutable(current); err != nil {
		return err
	}
	if requested == nil || *requested == current {
		return nil
	}
	if *requested == model.ReservationStatusNoShow {
		return apperrors.WrapInvalidInput("no_show is owned by the no-show transition")
	}
	if *requested == model.ReservationStatusCompleted {
		return apperrors.WrapInvalidInput("completed is owned by the accounting completion transition")
	}
	return nil
}

func trimmingIdentityChanged(current *model.Reservation, input UpdateTrimmingReservationInput) bool {
	if input.StartTime != nil && !input.StartTime.Equal(current.StartTime) {
		return true
	}
	if input.EndTime != nil && !input.EndTime.Equal(current.EndTime) {
		return true
	}
	if input.PetID != nil && (current.PetID == nil || *input.PetID != *current.PetID) {
		return true
	}
	if input.DoctorID != nil && (current.DoctorID == nil || *input.DoctorID != *current.DoctorID) {
		return true
	}
	return false
}

func (r *reservationRepository) UpdateForTrimming(
	ctx context.Context,
	clinicID, id uint64,
	input UpdateTrimmingReservationInput,
) (*model.Reservation, error) {
	if err := requireTrimmingAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	current, err := r.lockTrimmingByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if err := validateTrimmingStatusTransition(current.Status, input.Status); err != nil {
		return nil, err
	}
	if trimmingIdentityChanged(current, input) {
		count, err := r.CountMedicalRecordsByReservationID(ctx, clinicID, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to check trimming appointment dependencies")
		}
		if count > 0 {
			return nil, apperrors.WrapConflict("この予約にはカルテが紐付いているため患者・担当者・日時を変更できません")
		}
	}
	ownerID := current.OwnerID
	petIDForOwnerResolution := input.PetID
	if petIDForOwnerResolution == nil && ownerID == nil {
		petIDForOwnerResolution = current.PetID
	}
	if petIDForOwnerResolution != nil {
		ownerID, err = r.resolveTrimmingOwnerForPet(ctx, clinicID, ownerID, petIDForOwnerResolution)
		if err != nil {
			return nil, err
		}
	}
	if input.DoctorID != nil {
		if err := r.assertStaffAssignedToClinic(ctx, clinicID, *input.DoctorID); err != nil {
			return nil, err
		}
	}
	fields := make(map[string]any, 6)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if current.OwnerID == nil && ownerID != nil {
		fields["owner_id"] = *ownerID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if len(fields) == 0 {
		return r.FindTrimmingByID(ctx, clinicID, id)
	}
	query := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Where("appointments.clinic_id = ? AND appointments.id = ? AND appointments.deleted_at IS NULL", clinicID, id).
		Where("appointments.status = ?", current.Status).
		Where(`EXISTS (
			SELECT 1 FROM reservation_types rt
			WHERE rt.id = appointments.reservation_type_id
			  AND rt.clinic_id = ?
			  AND rt.category = ?
			  AND rt.deleted_at IS NULL
		)`, clinicID, model.ReservationTypeCategoryTrimming)
	result := query.Updates(fields)
	if result.Error != nil {
		// EMR-76: reschedule/update can also hit excl_appointments_doctor_timerange.
		if conflict := asReservationTimeConflict(result.Error); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindTrimmingByID(ctx, clinicID, id); err != nil {
			return nil, err
		}
		return nil, apperrors.WrapConflict("予約が同時に更新されました。再読み込みしてください")
	}
	return r.FindTrimmingByID(ctx, clinicID, id)
}

func (r *reservationRepository) DeleteForTrimming(ctx context.Context, clinicID, id uint64) error {
	if err := requireTrimmingAmbientTransaction(ctx); err != nil {
		return err
	}
	current, err := r.lockTrimmingByID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if err := ValidateTrimmingAppointmentMutable(current.Status); err != nil {
		return err
	}
	count, err := r.CountMedicalRecordsByReservationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check trimming appointment dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この予約にはカルテが紐付いているため削除できません")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Where("appointments.clinic_id = ? AND appointments.id = ? AND appointments.deleted_at IS NULL", clinicID, id).
		Where(`EXISTS (
			SELECT 1 FROM reservation_types rt
			WHERE rt.id = appointments.reservation_type_id
			  AND rt.clinic_id = ?
			  AND rt.category = ?
			  AND rt.deleted_at IS NULL
		)`, clinicID, model.ReservationTypeCategoryTrimming).
		Delete(&model.Reservation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming appointment", fmt.Sprintf("%d", id))
	}
	return nil
}

// ---------------------------------------------------------------------------
// EMR-308: next-visit intent（仮予約）
//
// カルテ上の「次回予定」は appointments の tentative 行で保持する。日付のみが
// 確定するため intended_date が意味上の予定日を担い、start_time/end_time は
// JST 00:00–00:15 の sentinel を使う。reservation_route は record_intent、
// doctor_id は NULL、reservation_type_id は医院の general 区分 placeholder。
// source_medical_record_id は登録元カルテへの provenance を保持する。
// tentative は予約枠容量・時間枠競合・通知の対象外（booking pipeline に乗らない）。
// ---------------------------------------------------------------------------

const nextVisitIntentRoute = "record_intent"

// nextVisitIntentDurationMinutes は tentative 行の sentinel end を定義する幅。
const nextVisitIntentDurationMinutes = 15

// SetNextVisitIntentInput は 1医院・1ペットの active tentative を書き込む typed command。
// IntendedDate は暦日のみ意味を持ち、service 層で日付窓（過去不可・約2年上限）を検証済み。
// SourceMedicalRecordID は非 nil のとき登録元カルテの同一 clinic・同一 pet を検証する。
type SetNextVisitIntentInput struct {
	IntendedDate          time.Time
	SourceMedicalRecordID *uint64
	CreatedBy             *uint64
	// CanUpdateExisting は reservations:edit 権限を handler が解決して渡す。
	// active tentative が既にある場合の upsert は既存行の更新であり、create のみの
	// ロールが他人の intent を上書きできないよう既存行の上書き時に 403 へ倒す。
	CanUpdateExisting bool
}

// ConfirmTentativeUpdate は tentative 行が本予約へ遷移する際に確定した値を渡す。
// clinic/pet の ownership と slot/capacity 検証は service 層で完了済みであること。
type ConfirmTentativeUpdate struct {
	StartTime         time.Time
	EndTime           time.Time
	ReservationTypeID uint64
	DoctorID          *uint64 // nil = 現在値維持、0 = 未指定（NULL）
	IsDesignated      *bool
	Notes             *string
	Status            model.ReservationStatus // pending または confirmed のみ許可
}

// requireNextVisitIntentAmbientTransaction は tentative 書き込みに ambient tx を必須化する。
// 欠落時は部分 commit 防止のため fail-closed にする。
func requireNextVisitIntentAmbientTransaction(ctx context.Context) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("next visit intent write requires an ambient transaction")
	}
	return nil
}

// acquireNextVisitIntentLock は (clinic, pet) 単位の advisory lock を ambient tx に
// 紐づけて取得し、set/clear/confirm の create→update 判定を直列化する。
func (r *reservationRepository) acquireNextVisitIntentLock(ctx context.Context, clinicID, petID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("next visit intent lock requires an ambient transaction")
	}
	lockKey := fmt.Sprintf("next-visit-intent:%d:%d", clinicID, petID)
	if err := persistence.DBOrTx(ctx, r.db).
		Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).
		Error; err != nil {
		return apperrors.FromGORM(err, "reservation", lockKey)
	}
	return nil
}

// AcquireNextVisitIntentLock は (clinic,pet) intent advisory lock を ambient tx 上で
// 取得する公開口（ReservationSlotRepository 経由）。Codex #4: tentative 関連遷移の
// lock 順序を「intent lock → 行 lock」に統一するため service 層から先取りできる。
// pg advisory lock は同一 tx 内で再入可能なため、repo 側の再取得は no-op となる。
func (r *reservationRepository) AcquireNextVisitIntentLock(ctx context.Context, clinicID, petID uint64) error {
	return r.acquireNextVisitIntentLock(ctx, clinicID, petID)
}

// FindPetIDForIntentLock は intent advisory lock の引数となる pet_id を非ロック read で
// 解決する。reservationRelationsMatchParentClinic は意図的に適用しない — Codex #16 の
// ペット譲渡直後のように「行の owner_id ≠ pet.owner_id」の一時不整合行を、ロック取得と
// その後の owner 追従の対象から外すと confirm/遷移が not found で失敗し不整合を
// 自己修復できない。clinic_id スコープは維持するため、別 clinic の行には到達しない。
func (r *reservationRepository) FindPetIDForIntentLock(ctx context.Context, clinicID, id uint64) (*uint64, error) {
	var row struct {
		ID    uint64
		PetID *uint64
	}
	if err := persistence.DBOrTx(ctx, r.db).Raw(
		`SELECT id, pet_id FROM appointments WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL`,
		clinicID, id,
	).Scan(&row).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	if row.ID == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return row.PetID, nil
}

// WithTentativeReminderSerialization は tentative mutation（set/clear/confirm/
// absorb/checkin）と同じ (clinic,pet) intent advisory lock を保持する tx 内で
// fn を実行する（EMR-308/Codex: LSTEP リマインダーの再検証〜外部送信 race 対策）。
// petIDs は送信根拠となり得る全 candidate ペット —— Codex #22: 同一 owner の
// 同日 tentative が複数ペットに跨る場合、一部だけをロックすると「ロックした
// intent ≠ 送信根拠の intent」になり得るため、全対象を昇順に取得して直列化する
// （昇順は複数ロックのデッドロック回避のための固定順序）。ロック取得時点で
// コミット済みの mutation は fn 内の再読み込みに反映され、ロック保持中に
// 到着した mutation は tx 完了まで待機するため、stale な tentative に基づく
// 外部送信を防げる。fn に渡す ctx が tx スコープの再検証に使うもので、
// 外部送信や別リポジトリへの書込は呼び出し側が元の ctx で行ってよい。
func (r *reservationRepository) WithTentativeReminderSerialization(
	ctx context.Context,
	clinicID uint64,
	petIDs []uint64,
	fn func(txCtx context.Context) error,
) error {
	// 重複除去 + 昇順固定: 複数 (clinic,pet) ロックの取得順を一意に定める。
	sorted := make([]uint64, 0, len(petIDs))
	seen := make(map[uint64]struct{}, len(petIDs))
	for _, petID := range petIDs {
		if petID == 0 {
			continue
		}
		if _, dup := seen[petID]; dup {
			continue
		}
		seen[petID] = struct{}{}
		sorted = append(sorted, petID)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		for _, petID := range sorted {
			if err := r.acquireNextVisitIntentLock(txCtx, clinicID, petID); err != nil {
				return err
			}
		}
		// Hold pet state/ownership through revalidation and external delivery.
		// NOWAIT avoids a lock cycle with existing writers that already hold a
		// pet SHARE lock before waiting for the intent lock. A conflicting death,
		// deletion or transfer is a known-unsent error that the caller can retry.
		var pets []model.Pet
		if len(sorted) > 0 {
			if err := tx.Model(&model.Pet{}).
				Select("id").
				Where("clinic_id = ? AND id IN ?", clinicID, sorted).
				Order("id ASC").
				Clauses(clause.Locking{Strength: "SHARE", Options: "NOWAIT"}).
				Find(&pets).Error; err != nil {
				return apperrors.Wrap(err, "failed to lock reminder pet state")
			}
		}
		return fn(txCtx)
	})
}

// findActiveTentativeForUpdate はペットの active tentative を行ロックして返す。未存在は nil を返す。
func (r *reservationRepository) findActiveTentativeForUpdate(ctx context.Context, clinicID, petID uint64) (*model.Reservation, error) {
	var appointment model.Reservation
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, petID, model.ReservationStatusTentative).
		Order("id DESC").
		Take(&appointment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d pet=%d tentative", clinicID, petID))
	}
	return &appointment, nil
}

// findGeneralReservationTypeID は tentative placeholder 用に医院の general 診療区分を解決する。
// Codex #15: 無効化区分は新しい intent の placeholder に使えない — is_active を
// ORDER BY の優先度ではなく必須述語にし、active な general 区分の未存在は
// InvalidInput で fail-closed にする（列は NOT NULL）。
func (r *reservationRepository) findGeneralReservationTypeID(ctx context.Context, clinicID uint64) (uint64, error) {
	var reservationType model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("reservation_types.id").
		Where("clinic_id = ? AND category = ? AND is_active = TRUE AND deleted_at IS NULL", clinicID, model.ReservationTypeCategoryGeneral).
		Order("id ASC").
		Take(&reservationType).Error
	if err != nil {
		mapped := apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("clinic=%d general", clinicID))
		if apperrors.IsNotFound(mapped) {
			return 0, apperrors.WrapInvalidInput("医院に一般の予約区分がありません")
		}
		return 0, mapped
	}
	return reservationType.ID, nil
}

// assertSourceMedicalRecordForPet は provenance の整合を守る。参照カルテは同一 clinic で
// 同一 pet に属しなければならない。
func (r *reservationRepository) assertSourceMedicalRecordForPet(ctx context.Context, clinicID, medicalRecordID, petID uint64) error {
	var record model.MedicalRecord
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("medical_records.id", "medical_records.pet_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", medicalRecordID, clinicID).
		Take(&record).Error
	if err != nil {
		return apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", medicalRecordID))
	}
	if record.PetID == nil || *record.PetID != petID {
		return apperrors.WrapInvalidInput("source medical record does not belong to this pet")
	}
	return nil
}

// nextVisitIntentTimeBounds は暦日を JST sentinel 窓（00:00–00:15）にマップする。
func nextVisitIntentTimeBounds(intendedDate time.Time) (start, end time.Time) {
	jst := intendedDate.In(config.JST)
	start = time.Date(jst.Year(), jst.Month(), jst.Day(), 0, 0, 0, 0, config.JST)
	return start, start.Add(nextVisitIntentDurationMinutes * time.Minute)
}

// SetNextVisitIntent はペットの次回予定を書き込む。active tentative が既にある場合は
// 同一行を更新（upsert）し、なければ sentinel 時刻の tentative 行を作成する。
// upsert は既存行の更新のため input.CanUpdateExisting（=reservations:edit）を要求する。
// 第2戻り値は upsert で変更された旧 owner_id（未変更/新規は nil）— caller の
// post-commit タグ再同期が旧飼い主側の stale タグも除去できるようにする。
// 容量・時間枠・通知パスには乗らない。部分ユニーク索引違反は friendly error に変換する。
func (r *reservationRepository) SetNextVisitIntent(
	ctx context.Context,
	clinicID, petID uint64,
	input SetNextVisitIntentInput,
) (*model.Reservation, *uint64, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, nil, err
	}
	if input.IntendedDate.IsZero() {
		return nil, nil, apperrors.WrapInvalidInput("intended_date is required")
	}
	if err := r.acquireNextVisitIntentLock(ctx, clinicID, petID); err != nil {
		return nil, nil, err
	}
	// Codex #19: 事前検証〜ロック取得の待機中に JST 暦日境界を跨ぐと「本日以降」
	// 判定が stale になるため、intent lock 保持中の時点で日付窓を再検証する。
	// 上限側は待機で悪化し得ないため実害は past 境界のみだが、同一規則で再評価する。
	if err := validateNextVisitIntentDate(input.IntendedDate, r.now()); err != nil {
		return nil, nil, err
	}
	ownerID, err := r.FindPetOwnerInClinic(ctx, clinicID, petID)
	if err != nil {
		return nil, nil, err
	}
	if input.SourceMedicalRecordID != nil {
		if err := r.assertSourceMedicalRecordForPet(ctx, clinicID, *input.SourceMedicalRecordID, petID); err != nil {
			return nil, nil, err
		}
	}
	start, end := nextVisitIntentTimeBounds(input.IntendedDate)
	intended := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, config.JST)

	existing, err := r.findActiveTentativeForUpdate(ctx, clinicID, petID)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		if !input.CanUpdateExisting {
			return nil, nil, apperrors.WrapForbidden("既存の次回予定を更新するには予約の編集権限が必要です")
		}
		// Codex #19 再指摘（PR #621 review #3）: 上の再検証後にも参照カルテ
		// FOR SHARE・既存行 FOR UPDATE のロック待ちで暦日を跨ぎ得るため、
		// 行ロック確保後・UPDATE 書込み直前に時計を取り直して再検証する。
		if err := validateNextVisitIntentDate(input.IntendedDate, r.now()); err != nil {
			return nil, nil, err
		}
		fields := map[string]any{
			"intended_date": intended,
			"start_time":    start,
			"end_time":      end,
			// ペット譲渡で現 owner が変わっていたら intent 行の owner も追従する —
			// 旧 owner 宛の表示・リマインダー・タグ同期が残り続けるのを防ぐ。
			"owner_id": ownerID,
		}
		if input.SourceMedicalRecordID != nil {
			fields["source_medical_record_id"] = *input.SourceMedicalRecordID
		}
		updated, err := r.update(ctx, clinicID, existing.ID, fields)
		if err != nil {
			return nil, nil, err
		}
		var prevOwnerID *uint64
		if existing.OwnerID == nil || *existing.OwnerID != ownerID {
			prevOwnerID = existing.OwnerID
		}
		return updated, prevOwnerID, nil
	}

	reservationTypeID, err := r.findGeneralReservationTypeID(ctx, clinicID)
	if err != nil {
		return nil, nil, err
	}
	// Codex #19 再々指摘（PR #621 rereview）: Create は内部で actor の権限・所属を
	// FOR SHARE 検証する。それが最終日付検証より後に待機すると、ロック待ち中の
	// JST 暦日跨ぎで前日付の intended_date を INSERT し得る。ここで先に actor
	// 検証とロックを取得しておく（Create 内の再検証は既保持ロックで待機しない）。
	if err := assertReservationCreatedBy(ctx, r.db, clinicID, input.CreatedBy); err != nil {
		return nil, nil, err
	}
	// 最終検証 —— これ以降 INSERT までに他 transaction の commit を待つロック
	// 取得は残らない（Codex #19: UPDATE 経路と同じく書込直前に時計を取り直す）。
	if err := validateNextVisitIntentDate(input.IntendedDate, r.now()); err != nil {
		return nil, nil, err
	}
	route := nextVisitIntentRoute
	appointment := &model.Reservation{
		ClinicID:              clinicID,
		ReservationTypeID:     reservationTypeID,
		StartTime:             start,
		EndTime:               end,
		OwnerID:               &ownerID,
		PetID:                 &petID,
		VisitType:             model.VisitTypeRevisit,
		Status:                model.ReservationStatusTentative,
		Source:                model.ReservationSourceManual,
		ReservationRoute:      &route,
		IntendedDate:          &intended,
		SourceMedicalRecordID: input.SourceMedicalRecordID,
		CreatedBy:             input.CreatedBy,
		CustomerFields:        json.RawMessage("{}"),
	}
	if err := r.Create(ctx, appointment); err != nil {
		if apperrors.IsAlreadyExists(err) {
			return nil, nil, apperrors.WrapAlreadyExistsMessage("このペットの次回予定は既に登録されています")
		}
		return nil, nil, err
	}
	loaded, err := r.FindByID(ctx, clinicID, appointment.ID)
	if err != nil {
		return nil, nil, err
	}
	return loaded, nil, nil
}

// ClearNextVisitIntent はペットの active tentative を終了させる。
// 「終了」は通常キャンセル経路（Update isCancel）と同じ終端 semantics を取り、
// status=cancelled を記録してから論理削除する（sentinel status のまま消すと
// 履歴上「途中で消えた」状態になりキャンセル経路と不整合になる）。
// 他ステータスの予約には触れず、未存在時は冪等に成功する（found=false）。
func (r *reservationRepository) ClearNextVisitIntent(ctx context.Context, clinicID, petID uint64) (bool, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return false, err
	}
	if err := r.acquireNextVisitIntentLock(ctx, clinicID, petID); err != nil {
		return false, err
	}
	existing, err := r.findActiveTentativeForUpdate(ctx, clinicID, petID)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	cancelled := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Where("clinic_id = ? AND id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, existing.ID, model.ReservationStatusTentative).
		Update("status", model.ReservationStatusCancelled)
	if cancelled.Error != nil {
		return false, apperrors.FromGORM(cancelled.Error, "reservation", fmt.Sprintf("%d", existing.ID))
	}
	if cancelled.RowsAffected == 0 {
		return false, apperrors.WrapConflict("次回予定が同時に更新されました。再読み込みしてください")
	}
	// 行ロック保持中のため status は cancelled のまま変わらない。deleted_at 述語のみで確定する。
	result := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, existing.ID).
		Delete(&model.Reservation{})
	if result.Error != nil {
		return false, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", existing.ID))
	}
	if result.RowsAffected == 0 {
		return false, apperrors.WrapConflict("次回予定が同時に更新されました。再読み込みしてください")
	}
	return true, nil
}

// ConfirmTentative は tentative 予約を本予約（pending/confirmed）へ遷移する。
// caller の ambient tx が行ロックを保持し、status 述語が並行 confirm を fail-closed にする。
// intended_date と source_medical_record_id は provenance として維持する。
func (r *reservationRepository) ConfirmTentative(
	ctx context.Context,
	clinicID, id uint64,
	input ConfirmTentativeUpdate,
) (*model.Reservation, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	// Codex: LSTEP リマインダー送信との直列化のため、tentative mutation はすべて
	// (clinic,pet) intent advisory lock を通る必要がある。行ロック前にロックなし
	// read で pet を解決し intent lock → row lock の順序で取得する（absorb 経路と
	// 同一順序でデッドロックを回避）。Codex #16: pre-read は関連行ガードを通さない
	// FindPetIDForIntentLock を使う — 譲渡直後の stale owner 行を見失わないため。
	petID, err := r.FindPetIDForIntentLock(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if petID != nil {
		if err := r.acquireNextVisitIntentLock(ctx, clinicID, *petID); err != nil {
			return nil, err
		}
	}
	current, err := r.LockAndFindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if current.Status != model.ReservationStatusTentative {
		return nil, apperrors.WrapConflict("対象の予約は仮予約（次回予定）ではありません")
	}
	switch input.Status {
	case model.ReservationStatusPending, model.ReservationStatusConfirmed:
	default:
		return nil, apperrors.WrapInvalidInput("仮予約は pending または confirmed のみ確定できます")
	}
	if input.ReservationTypeID == 0 {
		return nil, apperrors.WrapInvalidInput("reservation_type_id is required to confirm a tentative appointment")
	}
	if input.DoctorID != nil && *input.DoctorID != 0 {
		if err := r.assertStaffAssignedToClinic(ctx, clinicID, *input.DoctorID); err != nil {
			return nil, err
		}
	}
	fields := map[string]any{
		"status":              input.Status,
		"start_time":          input.StartTime,
		"end_time":            input.EndTime,
		"reservation_type_id": input.ReservationTypeID,
	}
	// Codex #16: tentative 作成後にペットが譲渡されると行の owner_id は旧飼い主の
	// まま残る。intent lock + 行ロック保持中に pet を SHARE ロックして現飼い主へ
	// 追従させ、確定済み予約が旧飼い主宛に紐付き続けるのを防ぐ。
	if current.PetID != nil {
		currentOwnerID, err := r.FindPetOwnerInClinic(ctx, clinicID, *current.PetID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve current pet owner")
		}
		fields["owner_id"] = currentOwnerID
	}
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			fields["doctor_id"] = nil
		} else {
			fields["doctor_id"] = *input.DoctorID
		}
	}
	if input.IsDesignated != nil {
		fields["is_designated"] = *input.IsDesignated
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
		Where("status = ?", model.ReservationStatusTentative).
		Updates(fields)
	if result.Error != nil {
		// EMR-76: confirm の時間変更も excl_appointments_doctor_timerange に抵触しうる。
		if conflict := asReservationTimeConflict(result.Error); conflict != nil {
			return nil, conflict
		}
		return nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, findErr := r.FindByID(ctx, clinicID, id); findErr != nil {
			return nil, findErr
		}
		return nil, apperrors.WrapConflict("予約が同時に更新されました。再読み込みしてください")
	}
	return r.FindByID(ctx, clinicID, id)
}

// updateTentativeStatus は tentative 行限定の CAS 更新。
// service 層の FindByID はロックを取らないため、読み取り時点の tentative 判定と
// 書込時点の status がずれ得る。WHERE status='tentative' を必須述語にすることで、
// 並行 ConfirmTentative（tentative→pending/confirmed の CAS）に先勝ちされた行への
// checked_in/cancelled 上書きを RowsAffected=0 → conflict として fail-closed に弾く。
// write + reload を同一 tx に閉じるのは汎用 update と同じ read-after-write 契約。
func (r *reservationRepository) updateTentativeStatus(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
	validate tentativeTransitionValidator,
) (*model.Reservation, error) {
	var loaded *model.Reservation
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		// Codex: LSTEP リマインダー送信との直列化のため tentative mutation は
		// (clinic,pet) intent advisory lock を通す。CAS 前に tx 内 read で pet を
		// 解決し intent lock を取得する（lock 順序は intent → row で統一）。
		petID, findErr := r.FindPetIDForIntentLock(txCtx, clinicID, id)
		if findErr != nil {
			return findErr
		}
		if petID != nil {
			if lockErr := r.acquireNextVisitIntentLock(txCtx, clinicID, *petID); lockErr != nil {
				return lockErr
			}
		}
		if validate != nil {
			// Codex #3/#14/#19: ロック待ち中に commit され得る mutation（死亡・
			// intended_date 変更・並行昇格）を検出するため、post-lock で行を
			// 再読み込みして validator を適用する。validator が返す追加フィールド
			// （実受付時刻スタンプ等）は CAS 書込へマージする。
			locked, lockedErr := r.FindByID(txCtx, clinicID, id)
			if lockedErr != nil {
				return lockedErr
			}
			extra, validateErr := validate(txCtx, locked)
			if validateErr != nil {
				return validateErr
			}
			for k, v := range extra {
				fields[k] = v
			}
		}
		result := tx.Model(&model.Reservation{}).
			Where("clinic_id = ? AND id = ? AND status = ? AND deleted_at IS NULL",
				clinicID, id, model.ReservationStatusTentative).
			Updates(fields)
		if result.Error != nil {
			if conflict := asReservationTimeConflict(result.Error); conflict != nil {
				return conflict
			}
			return apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return apperrors.WrapConflict("次回予定が同時に更新されました。再読み込みしてください")
		}
		var reloadErr error
		loaded, reloadErr = r.FindByID(txCtx, clinicID, id)
		return reloadErr
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

// ---------------------------------------------------------------------------
// EMR-308 lifecycle absorb: tentative の同一行昇格
//
// すべての実予約作成経路（内部 POST /reservations・一括・LIFF・管理/walk-in）は
// INSERT 直前にこの共通入口を通し、ペットの active tentative があれば新規行を
// 作らず同じ id の tentative 行を実予約へ上書きする。intended_date /
// source_medical_record_id / reservation_route=record_intent の provenance は
// 昇格後も維持し、cancelled になった行を自動復活させない。
// ---------------------------------------------------------------------------

// AbsorbTentativeOnBooking は実予約作成時にペットの active tentative を同一行昇格する。
// tentative が無ければ通常 INSERT にフォールバックするため、各作成経路はこの
// 入口を Create の代わりに呼ぶだけで二重 INSERT を防げる。
// caller の ambient tx が (clinic,pet) advisory lock と行ロックを保持し、
// status=tentative 述語が並行昇格/キャンセルとの競合を fail-closed にする。
func (r *reservationRepository) AbsorbTentativeOnBooking(
	ctx context.Context,
	booking *model.Reservation,
) (*model.Reservation, *uint64, error) {
	return r.promoteOrCreateFromTentative(ctx, booking, false)
}

// PromoteTentativeOnCheckin は当日受付（checked_in 作成）時にペットの active
// tentative を同一行昇格する。general placeholder 区分は tentative 行のまま維持し、
// checked_in_at に実受付時刻をスタンプする。tentative 未存在時は通常 INSERT。
func (r *reservationRepository) PromoteTentativeOnCheckin(
	ctx context.Context,
	booking *model.Reservation,
) (*model.Reservation, *uint64, error) {
	return r.promoteOrCreateFromTentative(ctx, booking, true)
}

// promoteOrCreateFromTentative は booking.PetID が持つ active tentative を
// findActiveTentativeForUpdate で行ロックして同一行 UPDATE する共通実装。
// tentative 不在・pet 未解決（nil/0）のときは実予約を通常 INSERT して返す。
// toCheckin=true は当日受付昇格（status=checked_in・区分は placeholder 維持）。
func (r *reservationRepository) promoteOrCreateFromTentative(
	ctx context.Context,
	booking *model.Reservation,
	toCheckin bool,
) (*model.Reservation, *uint64, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, nil, err
	}
	if booking == nil {
		return nil, nil, apperrors.WrapInvalidInput("booking is required")
	}
	// 当日受付作成（toCheckin）は tentative 昇格・通常 INSERT のどちらでも実受付時刻を
	// checked_in_at にスタンプする（A2: actual check-in time）。
	if toCheckin && booking.CheckedInAt == nil {
		now := time.Now()
		booking.CheckedInAt = &now
	}
	// pet 未解決の booking は tentative 照合不能のため通常 INSERT にフォールバックする
	// （LIFF の未リンク顧客等、pet を持たない予約は tentative を持ち得ない）。
	if booking.PetID == nil || *booking.PetID == 0 {
		if err := r.Create(ctx, booking); err != nil {
			return nil, nil, err
		}
		return booking, nil, nil
	}
	if err := r.acquireNextVisitIntentLock(ctx, booking.ClinicID, *booking.PetID); err != nil {
		return nil, nil, err
	}
	existing, err := r.findActiveTentativeForUpdate(ctx, booking.ClinicID, *booking.PetID)
	if err != nil {
		return nil, nil, err
	}
	if existing == nil {
		if err := r.Create(ctx, booking); err != nil {
			return nil, nil, err
		}
		return booking, nil, nil
	}
	// Codex #14: 未来日の tentative は当日受付へ昇格しない。別日の来院意図行を
	// 受付行へ上書きすると intent が失われるため、未到来の tentative は温存して
	// 通常 INSERT にフォールバックする（intended_date は JST 暦日で比較）。
	if toCheckin && existing.IntendedDate != nil {
		intendedJST := existing.IntendedDate.In(config.JST)
		intendedDay := time.Date(intendedJST.Year(), intendedJST.Month(), intendedJST.Day(), 0, 0, 0, 0, config.JST)
		visitJST := booking.StartTime.In(config.JST)
		visitDay := time.Date(visitJST.Year(), visitJST.Month(), visitJST.Day(), 0, 0, 0, 0, config.JST)
		if intendedDay.After(visitDay) {
			if err := r.Create(ctx, booking); err != nil {
				return nil, nil, err
			}
			return booking, nil, nil
		}
	}
	if booking.DoctorID != nil && *booking.DoctorID != 0 {
		if err := r.assertStaffAssignedToClinic(ctx, booking.ClinicID, *booking.DoctorID); err != nil {
			return nil, nil, err
		}
	}
	fields := tentativeAbsorbFields(booking, toCheckin, time.Now())
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", booking.ClinicID, existing.ID).
		Where("status = ?", model.ReservationStatusTentative).
		Updates(fields)
	if result.Error != nil {
		// EMR-76: 昇格時の時刻変更も excl_appointments_doctor_timerange に抵触しうる。
		if conflict := asReservationTimeConflict(result.Error); conflict != nil {
			return nil, nil, conflict
		}
		return nil, nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", existing.ID))
	}
	if result.RowsAffected == 0 {
		return nil, nil, apperrors.WrapConflict("次回予定が同時に更新されました。再読み込みしてください")
	}
	promoted, err := r.FindByID(ctx, booking.ClinicID, existing.ID)
	if err != nil {
		return nil, nil, err
	}
	// existing was read under the same intent and row locks as this promotion.
	return promoted, existing.OwnerID, nil
}

// tentativeAbsorbFields は tentative 行へ実予約の値を上書きする列セットを組み立てる。
// id / clinic_id / pet_id（照合キー）と provenance（intended_date /
// source_medical_record_id / reservation_route=record_intent）は更新しない。
// toCheckin（当日受付昇格）でも要求区分を書き込む — general placeholder を維持すると
// trimming 当日受付が診療ワークフロー・カルテ自動作成へ誤振分けされるため。
// owner_id / line_customer_id / customer_fields / created_by は booking 側に実値が
// ある場合だけ上書きし、未解決（nil/空）なら tentative 側の provenance を残す。
func tentativeAbsorbFields(booking *model.Reservation, toCheckin bool, checkedInAt time.Time) map[string]any {
	status := booking.Status
	if toCheckin {
		status = model.ReservationStatusCheckedIn
	} else if status == "" || status == model.ReservationStatusTentative {
		// INSERT 側の DB default（'pending'）と同じ解決。昇格後に tentative を
		// 残さないため、空・tentative 指定も pending へ正規化する。
		status = model.ReservationStatusPending
	}
	visitType := booking.VisitType
	if visitType == "" {
		visitType = model.VisitTypeRevisit
	}
	source := booking.Source
	if source == "" {
		source = model.ReservationSourceManual
	}
	fields := map[string]any{
		"start_time":         booking.StartTime,
		"end_time":           booking.EndTime,
		"visit_type":         visitType,
		"is_designated":      booking.IsDesignated,
		"is_staff_delegated": booking.IsStaffDelegated,
		"status":             status,
		"notes":              booking.Notes,
		"source":             source,
	}
	fields["reservation_type_id"] = booking.ReservationTypeID
	if status == model.ReservationStatusCheckedIn {
		stamp := checkedInAt
		if booking.CheckedInAt != nil {
			stamp = *booking.CheckedInAt
		}
		fields["checked_in_at"] = stamp
	}
	if booking.DoctorID == nil {
		fields["doctor_id"] = nil
	} else {
		fields["doctor_id"] = *booking.DoctorID
	}
	if booking.OwnerID != nil {
		fields["owner_id"] = *booking.OwnerID
	}
	if booking.LineCustomerID != nil {
		fields["line_customer_id"] = *booking.LineCustomerID
	}
	if len(booking.CustomerFields) > 0 {
		fields["customer_fields"] = booking.CustomerFields
	}
	if booking.CreatedBy != nil {
		fields["created_by"] = *booking.CreatedBy
	}
	return fields
}
