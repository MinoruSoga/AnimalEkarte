package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

const reservationListMaxLimit = 1000

func toReservationAvailableTimeResponse(slot *service.TimeSlot) liffTimeSlotResponse {
	return liffTimeSlotResponse{StartTime: slot.StartTime, EndTime: slot.EndTime}
}

// ListReservations godoc
func (h *Handler) ListReservations(c *gin.Context) {
	// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
	clinicIDs, ok := resolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := parsePaginationWithMax(c, reservationListMaxLimit)
	if err != nil {
		RespondError(c, err)
		return
	}

	q := newListReservationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicIDs, page, limit, filters.Date, filters.StartDate, filters.EndDate, filters.Status, filters.Source, filters.PetID, filters.OwnerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(reservations, toReservationResponse), total, page, limit))
}

// GetReservation godoc
func (h *Handler) GetReservation(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := resolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	reservation, err := h.svc.Reservation.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// GetReservationAvailableTimes godoc
// GET /reservations/available-times?reservation_type_id=:id&staff_id=:id&date=YYYY-MM-DD
func (h *Handler) GetReservationAvailableTimes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	filters, err := newReservationAvailableTimesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}
	if h.svc.Liff == nil {
		RespondError(c, apperrors.WrapNotImplemented("予約可能時間の取得は未設定です"))
		return
	}
	slots, err := h.svc.Liff.GetAvailableTimes(c.Request.Context(), clinicID, filters.ReservationTypeID, filters.StaffID, filters.Date)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(slots, toReservationAvailableTimeResponse))
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

	svcInput, err := input.toServiceInput(clinicID, staffID)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			RespondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Reservation.Create(ctx, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	// #83: 予約確定(confirmed)で作成された場合はカルテを best-effort で自動作成する
	if shouldAutoCreateMedicalRecordForReservation(reservation) && h.svc.MedicalRecord != nil {
		h.svc.MedicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
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

	svcInput, err := input.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
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

	// 予約確定/受付済みの場合は通常カルテを best-effort で自動作成する(#83)。owner/pet 紐付け後の遅延作成(Q9)も兼ねる。
	if shouldAutoCreateMedicalRecordForReservation(reservation) && h.svc.MedicalRecord != nil {
		h.svc.MedicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}

	// #83 Q10: キャンセルされた予約に紐づく draft カルテを論理削除する（best-effort）
	if svcInput.Status != nil && *svcInput.Status == model.ReservationStatusCancelled && h.svc.MedicalRecord != nil {
		h.svc.MedicalRecord.DeleteDraftFromReservation(ctx, clinicID, reservation.ID)
	}

	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

func shouldAutoCreateMedicalRecordForReservation(reservation *model.Reservation) bool {
	if reservation == nil {
		return false
	}
	// #83: 予約確定(confirmed)/受付済み(checked_in)の予約でカルテを作成する。
	// 更新後の reservation.Status で判定するため、owner/pet 紐付けの Update でも confirmed なら作成され、
	// owner/pet 未確定でスキップされた分が紐付け後に遅延作成される(Q9)。
	// 同日同ペットの重複は AutoCreateFromReservation 側でスキップされるため二重作成しない。
	if reservation.Status != model.ReservationStatusConfirmed && reservation.Status != model.ReservationStatusCheckedIn {
		return false
	}
	if reservation.ReservationType != nil && reservation.ReservationType.Category == model.ReservationTypeCategoryTrimming {
		return false
	}
	if reservation.ReservationType != nil &&
		(strings.Contains(reservation.ReservationType.Name, "入院") ||
			strings.Contains(reservation.ReservationType.Name, "ホテル")) {
		return false
	}
	return true
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
	if err := h.svc.Reservation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
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
		req.toServiceInput(),
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
	reservations.GET("/available-times", h.RequirePermission(string(model.ResourceReservations), "view"), h.GetReservationAvailableTimes)
	reservations.GET("/:id", h.RequirePermission(string(model.ResourceReservations), "view"), h.GetReservation)
	reservations.POST("", h.RequirePermission(string(model.ResourceReservations), "create"), h.CreateReservation)
	reservations.PATCH("/:id", h.RequirePermission(string(model.ResourceReservations), "edit"), h.UpdateReservation)
	reservations.PATCH("/:id/reservation-route", h.RequirePermission(string(model.ResourceReservations), "edit"), h.PatchReservationReservationRoute)
	reservations.DELETE("/:id", h.RequirePermission(string(model.ResourceReservations), "delete"), h.DeleteReservation)
}
