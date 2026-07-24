package reservation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
			SELECT a.id, a.status
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
		RETURNING candidate.status
	`

	var previousStatus model.ReservationStatus
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
		Scan(&previousStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return NoShowTransition{}, nil
	}
	if err != nil {
		return NoShowTransition{}, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return NoShowTransition{Changed: true, PreviousStatus: previousStatus}, nil
}

func (r *reservationRepository) CreateForTrimming(
	ctx context.Context,
	clinicID uint64,
	input CreateTrimmingReservationInput,
) (*model.Reservation, error) {
	if err := requireTrimmingAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	if err := validateTrimmingCreateStatus(input.Status); err != nil {
		return nil, err
	}
	if err := r.assertActiveTrimmingReservationType(ctx, clinicID, input.ReservationTypeID); err != nil {
		return nil, err
	}
	ownerID, err := r.resolveTrimmingOwnerForPet(ctx, clinicID, nil, input.PetID)
	if err != nil {
		return nil, err
	}
	if input.DoctorID != nil {
		if err := r.assertStaffAssignedToClinic(ctx, clinicID, *input.DoctorID); err != nil {
			return nil, err
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
	if err := r.Create(ctx, appointment); err != nil {
		return nil, err
	}
	return appointment, nil
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
