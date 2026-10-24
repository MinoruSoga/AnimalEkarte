package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const reservationDraftCleanupTimeout = 5 * time.Second

func (s *medicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	if reservation == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation is nil")
		return
	}
	// 通常カルテを自動作成するのは business_kind が general の予約区分だけ。trimming / inpatient / hotel と NULL（未分類）は
	// 作成しない（001_init.sql L8893-8894: 旧 category は dispatch の fallback ではない。予約区分名の部分一致による
	// 入院・ホテル判定は廃止した）。
	if reservation.ReservationType != nil && !reservation.ReservationType.HasBusinessKind(model.ReservationBusinessKindGeneral) {
		slog.InfoContext(ctx, "autoCreateFromReservation: skipped — reservation type is not a general business kind",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	// BUG-386: LINE予約で owner_id / pet_id が未設定の場合、line_customer から補完する
	if (reservation.PetID == nil || reservation.OwnerID == nil) &&
		reservation.LineCustomerID != nil &&
		s.lineCustomerRepo != nil {
		s.fillOwnerPetFromLineCustomer(ctx, clinicID, reservation)
	}

	if reservation.PetID == nil || reservation.OwnerID == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation has no pet_id or owner_id",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	dateStr := reservation.StartTime.In(config.JST).Format(time.DateOnly)
	var createResult medicalRecordCreateTxResult
	err := s.withTx(ctx, func(txCtx context.Context) error {
		created, err := s.autoCreateFromReservationInTx(txCtx, clinicID, reservation, dateStr)
		if err != nil {
			return err
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

func (s *medicalRecordService) autoCreateFromReservationInTx(
	txCtx context.Context,
	clinicID uint64,
	reservation *model.Reservation,
	dateStr string,
) (medicalRecordCreateTxResult, error) {
	acquired, err := s.repo.AcquireAutoCreateLock(txCtx, clinicID, *reservation.PetID, dateStr)
	if err != nil {
		slog.WarnContext(txCtx, "autoCreateFromReservation: failed to acquire auto-create lock",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return medicalRecordCreateTxResult{}, err
	}
	if !acquired {
		slog.InfoContext(txCtx, "autoCreateFromReservation: skipped — auto-create lock is busy",
			slog.Uint64("reservation_id", reservation.ID),
			slog.Uint64("pet_id", *reservation.PetID),
			slog.String("date", dateStr))
		return medicalRecordCreateTxResult{}, nil
	}

	// advisory lock の保持中に同日同ペットの存在確認から Create 完了までを直列化する。
	total, err := s.repo.CountByPetAndDate(txCtx, clinicID, *reservation.PetID, dateStr)
	if err != nil {
		slog.WarnContext(txCtx, "autoCreateFromReservation: failed to check existing records",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return medicalRecordCreateTxResult{}, err
	}
	if total > 0 {
		slog.InfoContext(txCtx, "autoCreateFromReservation: skipped — same-day record already exists",
			slog.Uint64("pet_id", *reservation.PetID),
			slog.String("date", dateStr))
		return medicalRecordCreateTxResult{}, nil
	}

	input := CreateMedicalRecordInput{
		Date:          reservation.StartTime,
		OwnerID:       reservation.OwnerID,
		PetID:         reservation.PetID,
		AppointmentID: &reservation.ID,
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
		return medicalRecordCreateTxResult{}, err
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
		return medicalRecordCreateTxResult{}, apperrors.WrapConflict("appointment date changed during medical record auto-create")
	}
	return created, nil
}

func (s *medicalRecordService) fillOwnerPetFromLineCustomer(
	ctx context.Context,
	clinicID uint64,
	reservation *model.Reservation,
) {
	customer, err := s.lineCustomerRepo.FindByID(ctx, clinicID, *reservation.LineCustomerID)
	if err != nil || customer == nil || customer.OwnerID == nil {
		return
	}
	reservation.OwnerID = customer.OwnerID
	if reservation.PetID == nil && customer.Owner != nil && len(customer.Owner.Pets) == 1 {
		reservation.PetID = &customer.Owner.Pets[0].ID
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

// DeleteDraftFromReservation は予約キャンセル時の呼び出し口で、EMR-495 P1 からカルテを削除しない
// （ログのみの no-op）。旧実装は status=draft のカルテを論理削除していたが、spec（medical-records/lifecycle §3）は
// 「通常予約取消からの自動削除は禁止」「draftという状態だけでは削除可と判定しない」とするため、
// status 列が無い現在は記入済みカルテを巻き込まないよう何も消さない。呼び出し元の予約取消は影響を受けない。
func (s *medicalRecordService) DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64) {
	slog.InfoContext(ctx, "reservation cancel: medical record auto-delete is not provided; record kept",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_id", reservationID))
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
