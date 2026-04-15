package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// GetLiffSettings はLIFF公開設定を返す。
// GET /api/liff/:clinicId/settings
func (h *Handler) GetLiffSettings(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
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
	clinicID, ok := extractClinicIDFromParam(c)
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
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	courses, err := h.svc.Liff.GetCourses(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]liffCourseResponse, 0, len(courses))
	for i := range courses {
		resp = append(resp, toLiffCourseResponse(&courses[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// GetLiffStaffs はコース対応スタッフ一覧を返す。
// GET /api/liff/:clinicId/staffs?courseId=:id
func (h *Handler) GetLiffStaffs(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	typeIDStr := c.Query("courseId")
	typeID, err := strconv.ParseUint(typeIDStr, 10, 64)
	if err != nil || typeID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("invalid courseId"))
		return
	}
	staffs, err := h.svc.Liff.GetStaffs(c.Request.Context(), clinicID, typeID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]liffStaffResponse, 0, len(staffs))
	for i := range staffs {
		resp = append(resp, toLiffStaffResponse(&staffs[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// GetLiffAvailableDates は予約可能日付一覧を返す。
// GET /api/liff/:clinicId/available-dates?courseId=:id&staffId=:id
func (h *Handler) GetLiffAvailableDates(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	typeID, err := strconv.ParseUint(c.Query("courseId"), 10, 64)
	if err != nil || typeID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("invalid courseId"))
		return
	}
	staffID, _ := strconv.ParseUint(c.Query("staffId"), 10, 64) // 0 = 指名なし

	dates, window, err := h.svc.Liff.GetAvailableDates(c.Request.Context(), clinicID, typeID, staffID)
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
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	typeID, err := strconv.ParseUint(c.Query("courseId"), 10, 64)
	if err != nil || typeID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("invalid courseId"))
		return
	}
	staffID, _ := strconv.ParseUint(c.Query("staffId"), 10, 64)

	dateStr := c.Query("date")
	date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date: must be YYYY-MM-DD"))
		return
	}

	slots, err := h.svc.Liff.GetAvailableTimes(c.Request.Context(), clinicID, typeID, staffID, date)
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
	clinicID, ok := extractClinicIDFromParam(c)
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

	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid date: must be YYYY-MM-DD"))
		return
	}

	input := &service.CreateReservationInput{
		ReservationTypeID: req.TypeID,
		StaffID:           req.StaffID,
		Date:              date,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		CustomerFields:    req.CustomerFields,
		RequestText:       req.RequestText,
	}

	// 指名予約時: 医師がこのクリニックに所属しているか確認
	if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, req.StaffID); err != nil {
		RespondError(c, err)
		return
	}

	appt, err := h.svc.Liff.CreateReservation(c.Request.Context(), clinicID, customerID, input)
	if err != nil {
		// 予約制限エラーはフロントエンドに redirect_step を伝える
		if limErr, ok := service.IsReservationLimitError(err); ok {
			// NOTE: Intentional direct response — ReservationLimitError は RespondError で処理できないカスタムペイロード（redirect_step）を含む
			c.JSON(http.StatusConflict, gin.H{
				"error":         limErr.Error(),
				"code":          limErr.Code,
				"redirect_step": limErr.RedirectStep,
			})
			return
		}
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": appt.ID, "notes": appt.Notes})
}

// GetLiffMyReservations は顧客自身の予約一覧を返す。
// GET /api/liff/:clinicId/my-reservations
func (h *Handler) GetLiffMyReservations(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
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
	resp := make([]liffReservationResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toLiffReservationResponse(&items[i]))
	}
	c.JSON(http.StatusOK, resp)
}

// CancelLiffReservation は予約をキャンセルする。
// DELETE /api/liff/:clinicId/my-reservations/:id
func (h *Handler) CancelLiffReservation(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	idStr := c.Param("id")
	reservationID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || reservationID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("invalid reservation id"))
		return
	}

	if err := h.svc.Liff.CancelReservation(c.Request.Context(), clinicID, customerID, reservationID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
