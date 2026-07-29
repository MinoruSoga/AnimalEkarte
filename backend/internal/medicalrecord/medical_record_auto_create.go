package medicalrecord

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	reservationDraftCleanupFailedAction = "reservation.draft_cleanup_failed"
	reservationDraftCleanupTimeout      = 5 * time.Second
)

type reservationDraftCleanupFailureCategory string

const (
	reservationDraftCleanupDependencyConflict reservationDraftCleanupFailureCategory = "dependency_conflict"
	reservationDraftCleanupStateConflict      reservationDraftCleanupFailureCategory = "state_conflict"
	reservationDraftCleanupInternalError      reservationDraftCleanupFailureCategory = "internal_error"
)

func (s *medicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	if reservation == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation is nil")
		return
	}
	if reservation.ReservationType != nil && reservation.ReservationType.Category == model.ReservationTypeCategoryTrimming {
		slog.InfoContext(ctx, "autoCreateFromReservation: skipped — trimming reservation",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}
	if reservation.ReservationType != nil &&
		(strings.Contains(reservation.ReservationType.Name, "入院") ||
			strings.Contains(reservation.ReservationType.Name, "ホテル")) {
		slog.InfoContext(ctx, "autoCreateFromReservation: skipped — hospitalization reservation",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	// BUG-386: LINE予約で owner_id / pet_id が未設定の場合、line_customer から補完する
	if (reservation.PetID == nil || reservation.OwnerID == nil) &&
		reservation.LineCustomerID != nil &&
		s.lineCustomerRepo != nil {
		customer, err := s.lineCustomerRepo.FindByID(ctx, clinicID, *reservation.LineCustomerID)
		if err == nil && customer != nil && customer.OwnerID != nil {
			reservation.OwnerID = customer.OwnerID
			// ペットが1頭のみ登録されている場合は自動で紐付ける
			if reservation.PetID == nil && customer.Owner != nil && len(customer.Owner.Pets) == 1 {
				reservation.PetID = &customer.Owner.Pets[0].ID
			}
		}
	}

	if reservation.PetID == nil || reservation.OwnerID == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation has no pet_id or owner_id",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	dateStr := reservation.StartTime.In(config.JST).Format(time.DateOnly)
	var createResult medicalRecordCreateTxResult
	err := s.withTx(ctx, func(txCtx context.Context) error {
		acquired, err := s.repo.AcquireAutoCreateLock(txCtx, clinicID, *reservation.PetID, dateStr)
		if err != nil {
			slog.WarnContext(txCtx, "autoCreateFromReservation: failed to acquire auto-create lock",
				slog.Uint64("reservation_id", reservation.ID),
				slog.String("error", err.Error()))
			return err
		}
		if !acquired {
			slog.InfoContext(txCtx, "autoCreateFromReservation: skipped — auto-create lock is busy",
				slog.Uint64("reservation_id", reservation.ID),
				slog.Uint64("pet_id", *reservation.PetID),
				slog.String("date", dateStr))
			return nil
		}

		// advisory lock の保持中に同日同ペットの存在確認から Create 完了までを直列化する。
		total, err := s.repo.CountByPetAndDate(txCtx, clinicID, *reservation.PetID, dateStr)
		if err != nil {
			slog.WarnContext(txCtx, "autoCreateFromReservation: failed to check existing records",
				slog.Uint64("reservation_id", reservation.ID),
				slog.String("error", err.Error()))
			return err
		}
		if total > 0 {
			slog.InfoContext(txCtx, "autoCreateFromReservation: skipped — same-day record already exists",
				slog.Uint64("pet_id", *reservation.PetID),
				slog.String("date", dateStr))
			return nil
		}

		statusDraft := model.MedicalRecordStatusDraft
		input := CreateMedicalRecordInput{
			Date:          reservation.StartTime,
			OwnerID:       reservation.OwnerID,
			PetID:         reservation.PetID,
			AppointmentID: &reservation.ID,
			Status:        &statusDraft,
			VisitType:     model.VisitTypeRevisit,
		}
		if reservation.DoctorID != nil {
			input.DoctorID = reservation.DoctorID
		}

		created, err := s.createMedicalRecordInTx(txCtx, clinicID, &input)
		if err != nil {
			slog.WarnContext(txCtx, "autoCreateFromReservation: failed to create medical record",
				slog.Uint64("reservation_id", reservation.ID),
				slog.String("error", err.Error()))
			return err
		}
		resolvedRecord := created.record
		if created.existingHit != nil {
			resolvedRecord = created.existingHit
		}
		if resolvedRecord != nil &&
			resolvedRecord.Date.In(config.JST).Format(time.DateOnly) != dateStr {
			// The advisory key comes from the caller snapshot, while the create core locks and
			// rereads the authoritative appointment. If a concurrent reschedule changes the JST
			// day, roll this transaction back instead of inserting under a different day/key.
			return apperrors.WrapConflict("appointment date changed during medical record auto-create")
		}
		createResult = created
		return nil
	})
	if err != nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: auto-create transaction failed",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}
	if createResult.existingHit != nil {
		// Public Create returned this appointment-linked record before the extraction and
		// AutoCreate still ran the best-effort subrecord GetOrCreate path. Preserve that repair
		// without repeating create audit, welcome, or tag side effects.
		if err := s.CreateSubRecords(ctx, clinicID, createResult.existingHit.ID, CreateSubRecordsInput{}); err != nil {
			// MRC-04 / DEC-35: best-effort auto-create path must leave a durable failure signal.
			s.auditSubRecordsFailure(ctx, clinicID, createResult.existingHit.ID, reservation.ID, err)
		}
		return
	}
	if createResult.record == nil {
		return
	}

	s.runMedicalRecordCreatePostCommit(ctx, createResult)

	// サブテーブル（inquiry, clinical_plan）を空レコードで作成（best-effort + failure audit）
	if err := s.CreateSubRecords(ctx, clinicID, createResult.record.ID, CreateSubRecordsInput{}); err != nil {
		s.auditSubRecordsFailure(ctx, clinicID, createResult.record.ID, reservation.ID, err)
	}
}

// medicalRecordSubRecordsFailedAction is durable evidence for auto-create subrecord failures (MRC-04).
const medicalRecordSubRecordsFailedAction = "medical_record.subrecords_failed"

func (s *medicalRecordService) auditSubRecordsFailure(ctx context.Context, clinicID, medicalRecordID, reservationID uint64, cause error) {
	slog.ErrorContext(ctx, "autoCreateFromReservation: CreateSubRecords failed (best-effort)",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Uint64("reservation_id", reservationID),
		slog.String("error", cause.Error()))
	audit, ok := s.auditService.(AuditLogger)
	if !ok {
		slog.ErrorContext(ctx, "autoCreateFromReservation: subrecords failure audit dependency unavailable",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("medical_record_id", medicalRecordID),
			slog.Uint64("reservation_id", reservationID))
		return
	}
	auditCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), reservationDraftCleanupTimeout)
	defer cancel()
	if err := audit.LogEntry(auditCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     medicalRecordSubRecordsFailedAction,
		Resource:   model.AuditResourceReservation,
		ResourceID: &reservationID,
		Metadata: map[string]any{
			"medical_record_id": medicalRecordID,
			"failure":           cause.Error(),
		},
	}); err != nil {
		slog.ErrorContext(auditCtx, "autoCreateFromReservation: failed to audit CreateSubRecords failure",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("medical_record_id", medicalRecordID),
			slog.Uint64("reservation_id", reservationID),
			slog.String("error", err.Error()))
	}
}

// DeleteDraftFromReservation は予約キャンセル時に、その予約に紐づく draft カルテを論理削除する (#83 Q10、best-effort)。
// 診察開始済み(draft 以外)のカルテは削除しない。失敗してもキャンセル処理は止めない。
func (s *medicalRecordService) DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64) {
	// 予約キャンセルは既に確定済みのため、request cancellation や呼出元の ambient tx に
	// cleanup/audit を巻き込まない。一方で同期 best-effort 処理の上限は明示的に制限する。
	cleanupCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), reservationDraftCleanupTimeout)
	defer cancel()

	record, err := s.repo.FindByAppointmentID(cleanupCtx, clinicID, reservationID)
	if err != nil {
		slog.ErrorContext(cleanupCtx, "reservation cancel draft cleanup lookup failed (best-effort)",
			slog.Uint64("reservation_id", reservationID),
			slog.String("error", err.Error()))
		s.auditReservationDraftCleanupFailure(
			cleanupCtx, clinicID, reservationID, nil, reservationDraftCleanupInternalError,
		)
		return
	}
	if record == nil || record.Status != model.MedicalRecordStatusDraft {
		return
	}
	// 通常の削除経路へ委譲し、row lock・draft再確認・見積参照確認・監査を迂回しない。
	// Find後に状態が変わってもDelete側がtransaction内で再検証する。
	if err := s.Delete(cleanupCtx, clinicID, record.ID); err != nil {
		category := reservationDraftCleanupInternalError
		message := "reservation cancel draft cleanup failed (best-effort)"
		conflictKind, hasConflictKind := medicalRecordDeleteConflictKindFromError(err)
		switch {
		case hasConflictKind && conflictKind == medicalRecordDeleteDependencyConflict:
			category = reservationDraftCleanupDependencyConflict
			message = "reservation cancel draft cleanup blocked by dependency (best-effort)"
		case apperrors.IsConflict(err):
			category = reservationDraftCleanupStateConflict
			message = "reservation cancel draft cleanup blocked by record state change (best-effort)"
		}
		slog.ErrorContext(cleanupCtx, message,
			slog.Uint64("reservation_id", reservationID),
			slog.String("error", err.Error()))
		s.auditReservationDraftCleanupFailure(cleanupCtx, clinicID, reservationID, &record.ID, category)
	}
}

func (s *medicalRecordService) auditReservationDraftCleanupFailure(
	ctx context.Context,
	clinicID, reservationID uint64,
	medicalRecordID *uint64,
	category reservationDraftCleanupFailureCategory,
) {
	// cleanup lookup/Delete が timeout を使い切っても、失敗監査には独立した実行予算を与える。
	auditCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), reservationDraftCleanupTimeout)
	defer cancel()

	audit, ok := s.auditService.(AuditLogger)
	if !ok {
		slog.ErrorContext(auditCtx, "reservation cancel draft cleanup failure audit dependency unavailable",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("reservation_id", reservationID))
		return
	}

	metadata := map[string]any{
		"failure_category": string(category),
	}
	if medicalRecordID != nil {
		metadata["medical_record_id"] = *medicalRecordID
	}
	if err := audit.LogEntry(auditCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     reservationDraftCleanupFailedAction,
		Resource:   model.AuditResourceReservation,
		ResourceID: &reservationID,
		Metadata:   metadata,
	}); err != nil {
		slog.ErrorContext(auditCtx, "reservation cancel draft cleanup failure audit write failed",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("reservation_id", reservationID),
			slog.String("error", err.Error()))
	}
}

// resolveVisitTypeFromAppointment は AppointmentID 経由で Reservation.VisitType を取得する。
// AppointmentID なし、reservationRepo 未配線、Reservation 取得失敗の場合は空文字を返す。
func (s *medicalRecordService) resolveVisitTypeFromAppointment(ctx context.Context, record *model.MedicalRecord) string {
	if record.AppointmentID == nil || *record.AppointmentID == 0 {
		return ""
	}
	if s.reservationRepo == nil {
		return ""
	}
	res, err := s.reservationRepo.FindByID(ctx, record.ClinicID, *record.AppointmentID)
	if err != nil {
		slog.WarnContext(ctx, "reservation lookup failed for first-visit detection (non-fatal)",
			"appointment_id", *record.AppointmentID, "error", err)
		return ""
	}
	if res == nil {
		return ""
	}
	return string(res.VisitType)
}

// fallbackFirstVisitCheck は CountByOwnerID で初診判定する（飛び込みカルテ等のフォールバック）。
func (s *medicalRecordService) fallbackFirstVisitCheck(ctx context.Context, clinicID, ownerID uint64) bool {
	count, err := s.repo.CountByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.WarnContext(ctx, "first visit fallback check failed (non-fatal)",
			"owner_id", ownerID, "error", err)
		return false
	}
	return count == 0
}

// UpdateRecommendationReason は受診推奨理由を更新する（FEAT-381-2 Commit 2）。
// "" は NULL 保存。4値以外は InvalidInput。
