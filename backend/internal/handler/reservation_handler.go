package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservations godoc
func (h *Handler) ListReservations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var date *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid date format, use YYYY-MM-DD"))
			return
		}
		date = &t
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var petID *uint64
	if s := c.Query("pet_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
			return
		}
		petID = &id
	}
	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicID, page, limit, date, status, petID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(reservations, total, page, limit))
}

// GetReservation godoc
func (h *Handler) GetReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	reservation, err := h.svc.Reservation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, reservation)
}

// CreateReservation godoc
func (h *Handler) CreateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	reservation := &model.ReservationAppointment{
		ClinicID:      clinicID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ServiceTypeID: input.ServiceTypeID,
		DoctorID:      input.DoctorID,
		IsDesignated:  input.IsDesignated,
		Notes:         input.Notes,
	}
	if input.VisitType != "" {
		vt, err := validateEnum(input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid visit_type: "+err.Error()))
			return
		}
		reservation.VisitType = vt
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		reservation.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.Reservation.Create(ctx, reservation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reservation)
}

// UpdateReservation godoc
func (h *Handler) UpdateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateReservationInput{
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ServiceTypeID: input.ServiceTypeID,
		DoctorID:      input.DoctorID,
		IsDesignated:  input.IsDesignated,
		Notes:         input.Notes,
	}
	if input.VisitType != nil {
		vt, err := validateEnum(*input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid visit_type: "+err.Error()))
			return
		}
		svcInput.VisitType = &vt
	}
	if input.Status != nil {
		status, err := validateEnum(*input.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		svcInput.Status = &status
	}

	ctx := c.Request.Context()
	reservation, err := h.svc.Reservation.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 受付済みに変更された場合はカルテを best-effort で自動作成する（BE-reception-auto-create-medical-record）
	if svcInput.Status != nil && *svcInput.Status == model.ReservationStatusCheckedIn {
		h.autoCreateMedicalRecord(ctx, clinicID, reservation)
	}

	c.JSON(http.StatusOK, reservation)
}

// autoCreateMedicalRecord は受付済みステータス変更時にカルテを best-effort で自動作成する。
// 失敗してもメイン処理（予約更新）には影響しない。
// 同日同ペットのカルテが既に存在する場合はスキップする（重複防止）。
func (h *Handler) autoCreateMedicalRecord(ctx context.Context, clinicID uint64, reservation *model.ReservationAppointment) {
	if reservation.PetID == nil || reservation.OwnerID == nil {
		slog.WarnContext(ctx, "autoCreateMedicalRecord: skipped — reservation has no pet_id or owner_id",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	// 同日同ペットのカルテ存在チェック（重複防止）
	dateStr := reservation.StartTime.Format("2006-01-02")
	_, total, err := h.svc.MedicalRecord.List(ctx, clinicID, reservation.PetID, nil, &dateStr, &dateStr, 1, 1)
	if err != nil {
		slog.WarnContext(ctx, "autoCreateMedicalRecord: failed to check existing records",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}
	if total > 0 {
		slog.InfoContext(ctx, "autoCreateMedicalRecord: skipped — same-day record already exists",
			slog.Uint64("pet_id", *reservation.PetID),
			slog.String("date", dateStr))
		return
	}

	record := &model.MedicalRecord{
		ClinicID:                 clinicID,
		Date:                     reservation.StartTime,
		OwnerID:                  reservation.OwnerID,
		PetID:                    reservation.PetID,
		ReservationAppointmentID: &reservation.ID,
		Status:                   model.MedicalRecordStatusDraft,
	}
	if reservation.DoctorID != nil {
		record.DoctorID = reservation.DoctorID
	}

	if err := h.svc.MedicalRecord.Create(ctx, record); err != nil {
		slog.WarnContext(ctx, "autoCreateMedicalRecord: failed to create medical record",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}

	// サブテーブル（inquiry, clinical_plan）を空レコードで作成（best-effort）
	if _, err := h.svc.Inquiry.Upsert(ctx, service.UpsertInquiryInput{MedicalRecordID: record.ID}); err != nil {
		slog.WarnContext(ctx, "autoCreateMedicalRecord: failed to upsert inquiry",
			slog.Uint64("medical_record_id", record.ID),
			slog.String("error", err.Error()))
	}
	if _, err := h.svc.ClinicalPlan.GetOrCreate(ctx, record.ID); err != nil {
		slog.WarnContext(ctx, "autoCreateMedicalRecord: failed to get or create clinical plan",
			slog.Uint64("medical_record_id", record.ID),
			slog.String("error", err.Error()))
	}
}

// DeleteReservation godoc
func (h *Handler) DeleteReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Reservation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterReservationRoutes は予約関連のルートを登録する
func (h *Handler) RegisterReservationRoutes(rg *gin.RouterGroup) {
	reservations := rg.Group("/reservations")
	reservations.GET("", h.ListReservations)
	reservations.POST("", h.CreateReservation)
	reservations.GET("/:id", h.GetReservation)
	reservations.PATCH("/:id", h.UpdateReservation)
	reservations.DELETE("/:id", h.DeleteReservation)
}
