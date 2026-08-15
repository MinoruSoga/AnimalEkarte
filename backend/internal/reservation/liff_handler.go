package reservation

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LiffHandler は LIFF 公開 API の HTTP handler。
type LiffHandler struct {
	svc              LiffService
	staffAssignments staffAssignmentFinder
}

// NewLiffHandler は LiffHandler を構築する。
func NewLiffHandler(svc LiffService, staffAssignments staffAssignmentFinder) *LiffHandler {
	return &LiffHandler{svc: svc, staffAssignments: staffAssignments}
}

// checkDoctorClinicAssignment は他 handler と同一の忠実移植（旧 *Handler 共有メソッド）。
func (h *LiffHandler) checkDoctorClinicAssignment(ctx context.Context, clinicID, doctorID uint64) error {
	if doctorID == 0 {
		return nil
	}
	assignments, err := h.staffAssignments.FindAllByStaffID(ctx, doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff assignment")
	}
	for i := range assignments {
		if assignments[i].ClinicID == clinicID {
			return nil
		}
	}
	return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
}

// GetLiffSettings はLIFF公開設定を返す。
// GET /api/liff/:clinicId/settings
func (h *LiffHandler) GetLiffSettings(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	setting, err := h.svc.GetSettings(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffSettingsResponse(setting))
}

// GetLiffProfile はLIFF顧客プロフィールを返す。
// GET /api/liff/:clinicId/profile
func (h *LiffHandler) GetLiffProfile(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := extractLiffCustomerID(c)
	if !ok {
		respondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}
	profile, err := h.svc.GetProfile(c.Request.Context(), clinicID, customerID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffProfileResponse(profile))
}

// GetLiffTypes はLIFF向け公開コース一覧を返す。
// GET /api/liff/:clinicId/courses
func (h *LiffHandler) GetLiffTypes(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courses, err := h.svc.GetCourses(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := httpapi.MapSlice(courses, toLiffCourseResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffStaffs はコース対応スタッフ一覧を返す。
// GET /api/liff/:clinicId/staffs?courseId=:id
func (h *LiffHandler) GetLiffStaffs(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courseID, err := newLiffCourseQuery(c.Request.URL.Query()).toCourseID()
	if err != nil {
		respondError(c, err)
		return
	}
	staffs, err := h.svc.GetStaffs(c.Request.Context(), clinicID, courseID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := httpapi.MapSlice(staffs, toLiffStaffResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffAvailableDates は予約可能日付一覧を返す。
// GET /api/liff/:clinicId/available-dates?courseId=:id&staffId=:id
func (h *LiffHandler) GetLiffAvailableDates(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	filters, err := newLiffAvailableDatesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		respondError(c, err)
		return
	}

	dates, window, err := h.svc.GetAvailableDates(c.Request.Context(), clinicID, filters.CourseID, filters.StaffID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := liffAvailableDatesResponse{
		Dates:  make([]liffAvailableDateResponse, 0, len(dates)),
		Window: window,
	}
	for _, d := range dates {
		resp.Dates = append(resp.Dates, liffAvailableDateResponse(d))
	}
	c.JSON(http.StatusOK, resp)
}

// GetLiffAvailableTimes は指定日の予約可能時間枠を返す。
// GET /api/liff/:clinicId/available-times?courseId=:id&staffId=:id&date=YYYY-MM-DD
func (h *LiffHandler) GetLiffAvailableTimes(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	filters, err := newLiffAvailableTimesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		respondError(c, err)
		return
	}

	slots, err := h.svc.GetAvailableTimes(c.Request.Context(), clinicID, filters.CourseID, filters.StaffID, filters.Date)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := make([]liffTimeSlotResponse, 0, len(slots))
	for _, s := range slots {
		resp = append(resp, liffTimeSlotResponse(s))
	}
	c.JSON(http.StatusOK, resp)
}

// CreateLiffReservation は予約を確定する。
// POST /api/liff/:clinicId/reservations
func (h *LiffHandler) CreateLiffReservation(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := extractLiffCustomerID(c)
	if !ok {
		respondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	var req liffCreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-LINE-012: 入力サイズ制限（DoS・DB 肥大化対策）。
	// 値の内容検証（スキーマ準拠・HTML エスケープ）は将来の機能追加時に強化する。
	if err := validateLiffReservationInput(&req); err != nil {
		respondError(c, err)
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	// 指名予約時のみ所属チェック（StaffID=0 は「指名なし」）
	if req.StaffID != 0 {
		if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, req.StaffID); err != nil {
			respondError(c, err)
			return
		}
	}

	appt, err := h.svc.CreateReservation(c.Request.Context(), clinicID, customerID, input)
	if err != nil {
		// 予約制限エラーはフロントエンドに redirect_step を伝える
		if limErr, ok := IsReservationLimitError(err); ok {
			extras := map[string]any{
				"redirect_step": limErr.RedirectStep,
			}
			respondErrorWithExtras(c, limErr, extras)
			return
		}
		respondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", appt.ID))
	c.JSON(http.StatusCreated, toLiffReservationCreatedResponse(appt))
}

// GetLiffTrimmingCourses はLIFF向けトリミングコース一覧を返す。
// GET /api/liff/:clinicId/trimming-courses
func (h *LiffHandler) GetLiffTrimmingCourses(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	courses, err := h.svc.GetTrimmingCourses(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := httpapi.MapSlice(courses, toLiffTrimmingCourseResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffTrimmingOptions はLIFF向けトリミングオプション一覧を返す。
// GET /api/liff/:clinicId/trimming-options
func (h *LiffHandler) GetLiffTrimmingOptions(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	options, err := h.svc.GetTrimmingOptions(c.Request.Context(), clinicID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := httpapi.MapSlice(options, toLiffTrimmingOptionResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffMyReservations は顧客自身の予約一覧を返す。
// GET /api/liff/:clinicId/my-reservations
func (h *LiffHandler) GetLiffMyReservations(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := extractLiffCustomerID(c)
	if !ok {
		respondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	items, err := h.svc.GetMyReservations(c.Request.Context(), clinicID, customerID)
	if err != nil {
		respondError(c, err)
		return
	}
	resp := httpapi.MapSlice(items, toLiffReservationResponse)
	c.JSON(http.StatusOK, resp)
}

// GetLiffHealthCard はLIFF健康手帳データを返す。
// GET /api/liff/:clinicId/health-card
func (h *LiffHandler) GetLiffHealthCard(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := extractLiffCustomerID(c)
	if !ok {
		respondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}
	result, err := h.svc.GetHealthCard(c.Request.Context(), clinicID, customerID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLiffHealthCardResponse(result))
}

// CancelLiffReservation は予約をキャンセルする。
// DELETE /api/liff/:clinicId/my-reservations/:id
func (h *LiffHandler) CancelLiffReservation(c *gin.Context) {
	clinicID, ok := httpapi.ParseIDParam(c, "clinicId")
	if !ok {
		return
	}
	customerID, ok := extractLiffCustomerID(c)
	if !ok {
		respondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}

	reservationID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.CancelReservation(c.Request.Context(), clinicID, customerID, reservationID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// extractLiffCustomerID は middleware.ExtractLiffCustomerID の複製
// （middleware は service を import しており reservation からの import は循環。
// key 文字列は middleware.LiffCustomerIDKey() と同一値であることを liff_handler_test の
// 契約テストで固定する）。
func extractLiffCustomerID(c *gin.Context) (uint64, bool) {
	v, exists := c.Get("liff_customer_id")
	if !exists {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
