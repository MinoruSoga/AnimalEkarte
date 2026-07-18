package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetLiffSettings はLIFF公開設定を返す。
// GET /api/liff/:clinicId/settings
func (h *Handler) GetLiffSettings(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	setting, err := h.svc.Liff.GetSettings(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffSettingsResponse(setting))
}

// GetLiffProfile はLIFF顧客プロフィールを返す。
// GET /api/liff/:clinicId/profile
func (h *Handler) GetLiffProfile(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}
	profile, err := h.svc.Liff.GetProfile(c.Request.Context(), clinicID, customerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffProfileResponse(profile))
}

// GetLiffTypes はLIFF向け公開コース一覧を返す。
// GET /api/liff/:clinicId/courses
func (h *Handler) GetLiffTypes(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courses, err := h.svc.Liff.GetCourses(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := mapSlice(courses, toLiffCourseResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffStaffs はコース対応スタッフ一覧を返す。
// GET /api/liff/:clinicId/staffs?courseId=:id
func (h *Handler) GetLiffStaffs(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courseID, err := newLiffCourseQuery(c.Request.URL.Query()).toCourseID()
	if err != nil {
		RespondError(c, err)
		return
	}
	staffs, err := h.svc.Liff.GetStaffs(c.Request.Context(), clinicID, courseID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := mapSlice(staffs, toLiffStaffResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffAvailableDates は予約可能日付一覧を返す。
// GET /api/liff/:clinicId/available-dates?courseId=:id&staffId=:id
func (h *Handler) GetLiffAvailableDates(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	filters, err := newLiffAvailableDatesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	dates, window, err := h.svc.Liff.GetAvailableDates(c.Request.Context(), clinicID, filters.CourseID, filters.StaffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := liffAvailableDatesResponse{
		Dates:  make([]liffAvailableDateResponse, 0, len(dates)),
		Window: window,
	}
	for _, d := range dates {
		resp.Dates = append(resp.Dates, liffAvailableDateResponse{
			Date:      d.Date,
			Weekday:   d.Weekday,
			Available: d.Available,
			Reason:    d.Reason,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// GetLiffAvailableTimes は指定日の予約可能時間枠を返す。
// GET /api/liff/:clinicId/available-times?courseId=:id&staffId=:id&date=YYYY-MM-DD
func (h *Handler) GetLiffAvailableTimes(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	filters, err := newLiffAvailableTimesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	slots, err := h.svc.Liff.GetAvailableTimes(c.Request.Context(), clinicID, filters.CourseID, filters.StaffID, filters.Date)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]liffTimeSlotResponse, 0, len(slots))
	for _, s := range slots {
		resp = append(resp, liffTimeSlotResponse{StartTime: s.StartTime, EndTime: s.EndTime})
	}
	c.JSON(http.StatusOK, resp)
}

// CreateLiffReservation は予約を確定する。
// POST /api/liff/:clinicId/reservations
func (h *Handler) CreateLiffReservation(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	var req liffCreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-LINE-012: 入力サイズ制限（DoS・DB 肥大化対策）。
	// 値の内容検証（スキーマ準拠・HTML エスケープ）は将来の機能追加時に強化する。
	if err := validateLiffReservationInput(&req); err != nil {
		RespondError(c, err)
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	// 指名予約時のみ所属チェック（StaffID=0 は「指名なし」）
	if req.StaffID != 0 {
		if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, req.StaffID); err != nil {
			RespondError(c, err)
			return
		}
	}

	appt, err := h.svc.Liff.CreateReservation(c.Request.Context(), clinicID, customerID, input)
	if err != nil {
		// 予約制限エラーはフロントエンドに redirect_step を伝える
		if limErr, ok := service.IsReservationLimitError(err); ok {
			extras := map[string]any{
				"redirect_step": limErr.RedirectStep,
			}
			RespondErrorWithExtras(c, limErr, extras)
			return
		}
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", appt.ID))
	c.JSON(http.StatusCreated, toLiffReservationCreatedResponse(appt))
}

// GetLiffTrimmingCourses はLIFF向けトリミングコース一覧を返す。
// GET /api/liff/:clinicId/trimming-courses
func (h *Handler) GetLiffTrimmingCourses(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courses, err := h.svc.Liff.GetTrimmingCourses(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := mapSlice(courses, toLiffTrimmingCourseResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffTrimmingOptions はLIFF向けトリミングオプション一覧を返す。
// GET /api/liff/:clinicId/trimming-options
func (h *Handler) GetLiffTrimmingOptions(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	options, err := h.svc.Liff.GetTrimmingOptions(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := mapSlice(options, toLiffTrimmingOptionResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffMyReservations は顧客自身の予約一覧を返す。
// GET /api/liff/:clinicId/my-reservations
func (h *Handler) GetLiffMyReservations(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	items, err := h.svc.Liff.GetMyReservations(c.Request.Context(), clinicID, customerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := mapSlice(items, toLiffReservationResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffHealthCard はLIFF健康手帳データを返す。
// GET /api/liff/:clinicId/health-card
func (h *Handler) GetLiffHealthCard(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}
	result, err := h.svc.Liff.GetHealthCard(c.Request.Context(), clinicID, customerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffHealthCardResponse(result))
}

// CancelLiffReservation は予約をキャンセルする。
// DELETE /api/liff/:clinicId/my-reservations/:id
func (h *Handler) CancelLiffReservation(c *gin.Context) {
	clinicID, ok := parseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	reservationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.Liff.CancelReservation(c.Request.Context(), clinicID, customerID, reservationID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
