package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

	// 同日同ペットのカルテ存在チェック（重複防止）
	dateStr := reservation.StartTime.Format("2006-01-02")
	_, total, err := s.repo.FindAll(ctx, []uint64{clinicID}, repository.MedicalRecordListFilters{
		PetID:     reservation.PetID,
		StartDate: &dateStr,
		EndDate:   &dateStr,
	}, 1, 1)
	if err != nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: failed to check existing records",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}
	if total > 0 {
		slog.InfoContext(ctx, "autoCreateFromReservation: skipped — same-day record already exists",
			slog.Uint64("pet_id", *reservation.PetID),
			slog.String("date", dateStr))
		return
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

	record, err := s.Create(ctx, clinicID, &input)
	if err != nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: failed to create medical record",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}

	// サブテーブル（inquiry, clinical_plan）を空レコードで作成（best-effort）
	s.CreateSubRecords(ctx, clinicID, record.ID, CreateSubRecordsInput{})
}

// DeleteDraftFromReservation は予約キャンセル時に、その予約に紐づく draft カルテを論理削除する (#83 Q10、best-effort)。
// 診察開始済み(draft 以外)のカルテは削除しない。失敗してもキャンセル処理は止めない。
func (s *medicalRecordService) DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64) {
	if err := s.repo.DeleteDraftByAppointmentID(ctx, clinicID, reservationID); err != nil {
		slog.WarnContext(ctx, "failed to delete draft medical record on reservation cancel (best-effort)",
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
