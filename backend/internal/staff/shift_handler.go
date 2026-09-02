package staff

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListShiftEntries godoc
func (h *Handler) ListShiftEntries(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceShifts), "view") {
		return
	}

	filters, err := newListShiftEntriesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	shifts, err := h.svc.ShiftEntry.List(c.Request.Context(), clinicID, filters.YearMonth, filters.StaffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(shifts, toShiftResponse))
}

// CreateShiftEntry godoc
func (h *Handler) CreateShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createShiftRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	input, err := req.toServiceInput()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	shift, err := h.svc.ShiftEntry.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/shifts/%d", shift.ID))
	c.JSON(http.StatusCreated, toShiftResponse(shift))
}

// UpdateShiftEntry godoc
func (h *Handler) UpdateShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateShiftRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	shift, err := h.svc.ShiftEntry.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toShiftResponse(shift))
}

// DeleteShiftEntry godoc
func (h *Handler) DeleteShiftEntry(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.ShiftEntry.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetOnDutyStaffs は指定日に出勤しているスタッフ一覧を返す (BUG-344)
// GET /api/v1/shifts/on-duty-staffs?date=YYYY-MM-DD
func (h *Handler) GetOnDutyStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceShifts), "view") {
		return
	}
	date, err := newOnDutyStaffsQuery(c.Request.URL.Query()).toDate()
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}
	staffs, err := h.svc.ShiftEntry.GetOnDutyStaffs(c.Request.Context(), clinicID, date)
	if err != nil {
		RespondError(c, err)
		return
	}
	result := make([]staffSummaryResponse, 0, len(staffs))
	for i := range staffs {
		if s := toStaffSummary(&staffs[i]); s != nil {
			result = append(result, *s)
		}
	}
	c.JSON(http.StatusOK, result)
}

// RegisterShiftRoutes はシフト関連のルートを登録する
func (h *Handler) RegisterShiftRoutes(rg *gin.RouterGroup) {
	shifts := rg.Group("/shifts")
	shifts.GET("", h.RequirePermission(string(model.ResourceShifts), "view"), h.ListShiftEntries)
	shifts.GET("/on-duty-staffs", h.RequirePermission(string(model.ResourceShifts), "view"), h.GetOnDutyStaffs) // BUG-344
	shifts.POST("", h.RequirePermission(string(model.ResourceShifts), "create"), h.CreateShiftEntry)
	shifts.PATCH("/:id", h.RequirePermission(string(model.ResourceShifts), "edit"), h.UpdateShiftEntry)
	shifts.DELETE("/:id", h.RequirePermission(string(model.ResourceShifts), "delete"), h.DeleteShiftEntry)
}
