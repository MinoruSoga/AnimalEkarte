package handler

import (
	"fmt"
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

	var source *string
	if s := c.Query("source"); s != "" {
		source = &s
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicID, page, limit, date, status, source, petID, ownerID)
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
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	reservation, err := h.svc.Reservation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// CreateReservation godoc
func (h *Handler) CreateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	var input createReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	source := model.ReservationSourceManual
	if input.Source == string(model.ReservationSourceLine) {
		source = model.ReservationSourceLine
	}
	reservation := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:          input.DoctorID,
		IsDesignated:      input.IsDesignated,
		Notes:             input.Notes,
		Source:            source,
		CreatedBy:         &staffID,
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

	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if reservation.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *reservation.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	if err := h.svc.Reservation.Create(ctx, reservation); err != nil {
		RespondError(c, err)
		return
	}
	// BE-007: 予約登録タグ同期（best-effort）
	if reservation.OwnerID != nil && reservation.Status != model.ReservationStatusCancelled {
		_ = h.svc.LstepTagSync.SyncReservationTag(ctx, clinicID, *reservation.OwnerID, reservation.StartTime)
	}
	c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", reservation.ID))
	c.JSON(http.StatusCreated, toReservationResponse(reservation))
}

// UpdateReservation godoc
func (h *Handler) UpdateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateReservationInput{
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:          input.DoctorID,
		IsDesignated:      input.IsDesignated,
		Notes:             input.Notes,
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

	// BUG-144: staff_id のクリニック所属チェック（Update時も）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Reservation.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}

	// BE-007: 予約ステータス変更タグ同期（best-effort）
	if svcInput.Status != nil && reservation.OwnerID != nil {
		switch *svcInput.Status {
		case model.ReservationStatusCancelled:
			_ = h.svc.LstepTagSync.SyncCancellationTag(ctx, clinicID, *reservation.OwnerID)
		case model.ReservationStatusConfirmed, model.ReservationStatusPending:
			_ = h.svc.LstepTagSync.SyncReservationTag(ctx, clinicID, *reservation.OwnerID, reservation.StartTime)
		}
	}
	// 受付済みに変更された場合はカルテを best-effort で自動作成する（BE-reception-auto-create-medical-record）
	if svcInput.Status != nil && *svcInput.Status == model.ReservationStatusCheckedIn {
		h.svc.MedicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}

	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// DeleteReservation godoc
func (h *Handler) DeleteReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	// ISSUE-004: Delete 前に reservation を取得して owner_id を確保。
	reservation, err := h.svc.Reservation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	if err := h.svc.Reservation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	// ISSUE-004: 削除後の DB 状態から reserved_* を再構築（best-effort）。
	// 削除した予約以外に未来予約があれば、その最新日タグを保持する。
	// canceled_visit / no_show_* は削除を「キャンセル」として扱わないため変更しない。
	if reservation.OwnerID != nil {
		_ = h.svc.LstepTagSync.ResyncOwnerReservationTags(c.Request.Context(), clinicID, *reservation.OwnerID)
	}
	c.Status(http.StatusNoContent)
}

// PatchReservationReservationRoute godoc
// PATCH /reservations/:id/reservation-route — 予約経路を更新する（FEAT-381-2）。
func (h *Handler) PatchReservationReservationRoute(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationReservationRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	reservation, err := h.svc.Reservation.UpdateReservationRoute(
		c.Request.Context(), clinicID, id,
		service.UpdateReservationRouteInput{Route: req.Route},
	)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// RegisterReservationRoutes は予約関連のルートを登録する
func (h *Handler) RegisterReservationRoutes(rg *gin.RouterGroup) {
	reservations := rg.Group("/reservations")
	reservations.GET("", h.RequirePermission(string(model.ResourceReservations), "view"), h.ListReservations)
	reservations.GET("/:id", h.RequirePermission(string(model.ResourceReservations), "view"), h.GetReservation)
	reservations.POST("", h.RequirePermission(string(model.ResourceReservations), "create"), h.CreateReservation)
	reservations.PATCH("/:id", h.RequirePermission(string(model.ResourceReservations), "edit"), h.UpdateReservation)
	reservations.PATCH("/:id/reservation-route", h.RequirePermission(string(model.ResourceReservations), "edit"), h.PatchReservationReservationRoute)
	reservations.DELETE("/:id", h.RequirePermission(string(model.ResourceReservations), "delete"), h.DeleteReservation)
}
