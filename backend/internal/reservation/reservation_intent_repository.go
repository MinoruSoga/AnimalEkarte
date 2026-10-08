package reservation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	lockKey := fmt.Sprintf("next-visit-intent:%d:%d", clinicID, petID)
	if err := persistence.DBOrTx(ctx, r.db).
		Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).
		Error; err != nil {
		return apperrors.FromGORM(err, "reservation", lockKey)
	}
	return nil
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
// active を優先し、未存在なら InvalidInput で fail-closed にする（列は NOT NULL）。
func (r *reservationRepository) findGeneralReservationTypeID(ctx context.Context, clinicID uint64) (uint64, error) {
	var reservationType model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Select("reservation_types.id").
		Where("clinic_id = ? AND category = ? AND deleted_at IS NULL", clinicID, model.ReservationTypeCategoryGeneral).
		Order("is_active DESC, id ASC").
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
// 容量・時間枠・通知パスには乗らない。部分ユニーク索引違反は friendly error に変換する。
func (r *reservationRepository) SetNextVisitIntent(
	ctx context.Context,
	clinicID, petID uint64,
	input SetNextVisitIntentInput,
) (*model.Reservation, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	if input.IntendedDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("intended_date is required")
	}
	if err := r.acquireNextVisitIntentLock(ctx, clinicID, petID); err != nil {
		return nil, err
	}
	ownerID, err := r.FindPetOwnerInClinic(ctx, clinicID, petID)
	if err != nil {
		return nil, err
	}
	if input.SourceMedicalRecordID != nil {
		if err := r.assertSourceMedicalRecordForPet(ctx, clinicID, *input.SourceMedicalRecordID, petID); err != nil {
			return nil, err
		}
	}
	start, end := nextVisitIntentTimeBounds(input.IntendedDate)
	intended := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, config.JST)

	existing, err := r.findActiveTentativeForUpdate(ctx, clinicID, petID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		fields := map[string]any{
			"intended_date": intended,
			"start_time":    start,
			"end_time":      end,
		}
		if input.SourceMedicalRecordID != nil {
			fields["source_medical_record_id"] = *input.SourceMedicalRecordID
		}
		return r.update(ctx, clinicID, existing.ID, fields)
	}

	reservationTypeID, err := r.findGeneralReservationTypeID(ctx, clinicID)
	if err != nil {
		return nil, err
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
			return nil, apperrors.WrapAlreadyExistsMessage("このペットの次回予定は既に登録されています")
		}
		return nil, err
	}
	return r.FindByID(ctx, clinicID, appointment.ID)
}

// ClearNextVisitIntent はペットの active tentative を論理削除する。
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
	result := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND id = ? AND status = ? AND deleted_at IS NULL",
			clinicID, existing.ID, model.ReservationStatusTentative).
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
) (*model.Reservation, error) {
	return r.promoteOrCreateFromTentative(ctx, booking, false)
}

// PromoteTentativeOnCheckin は当日受付（checked_in 作成）時にペットの active
// tentative を同一行昇格する。general placeholder 区分は tentative 行のまま維持し、
// checked_in_at に実受付時刻をスタンプする。tentative 未存在時は通常 INSERT。
func (r *reservationRepository) PromoteTentativeOnCheckin(
	ctx context.Context,
	booking *model.Reservation,
) (*model.Reservation, error) {
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
) (*model.Reservation, error) {
	if err := requireNextVisitIntentAmbientTransaction(ctx); err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, apperrors.WrapInvalidInput("booking is required")
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
			return nil, err
		}
		return booking, nil
	}
	if err := r.acquireNextVisitIntentLock(ctx, booking.ClinicID, *booking.PetID); err != nil {
		return nil, err
	}
	existing, err := r.findActiveTentativeForUpdate(ctx, booking.ClinicID, *booking.PetID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.Create(ctx, booking); err != nil {
			return nil, err
		}
		return booking, nil
	}
	if booking.DoctorID != nil && *booking.DoctorID != 0 {
		if err := r.assertStaffAssignedToClinic(ctx, booking.ClinicID, *booking.DoctorID); err != nil {
			return nil, err
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
			return nil, conflict
		}
		return nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", existing.ID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapConflict("次回予定が同時に更新されました。再読み込みしてください")
	}
	return r.FindByID(ctx, booking.ClinicID, existing.ID)
}

// tentativeAbsorbFields は tentative 行へ実予約の値を上書きする列セットを組み立てる。
// id / clinic_id / pet_id（照合キー）と provenance（intended_date /
// source_medical_record_id / reservation_route=record_intent）は更新しない。
// toCheckin=true（当日受付昇格）では区分を tentative 行の general placeholder のまま
// 維持するため reservation_type_id を書き換えず、checked_in_at に実受付時刻を入れる。
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
	if !toCheckin {
		fields["reservation_type_id"] = booking.ReservationTypeID
	}
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
