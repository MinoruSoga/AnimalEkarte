// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// staffListMaxLimit は全スタッフ一括取得の上限。スタッフ数は現実的に数十〜数百名程度のため全件返却で問題ない。
const staffListMaxLimit = 1000

// ---- Staff ----

// ListStaffs godoc
// FE互換: 直接配列を返す（ページネーション不要）
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	// NOTE: pagination パラメータは無視（全件返却）
	// 将来的にページネーション対応が必要な場合は、別エンドポイント化を検討
	staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, 1, staffListMaxLimit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(staffs, toStaffResponse))
}

// CreateStaff godoc
func (h *Handler) CreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	// BUG-145: email が指定されている場合は重複チェックを行い、Account を作成してスタッフに紐づける。
	// Account 作成・bcrypt ハッシュ化・パスワードバリデーションはすべて StaffService に委譲する。
	var staff *model.Staff
	var err error

	if req.hasAccountEmail() {
		staff, err = h.svc.Staff.CreateWithAccount(ctx, req.toCreateWithAccountServiceInput(clinicID))
	} else {
		staff, err = h.svc.Staff.Create(ctx, req.toCreateServiceInput(clinicID))
	}
	if err != nil {
		RespondError(c, err)
		return
	}

	// NOTE: Best-effort reload for Preload data. Create already succeeded.
	if reloaded, reloadErr := h.svc.Staff.GetByID(ctx, staff.ID); reloadErr == nil {
		staff = reloaded
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/staffs/%d", staff.ID))
	c.JSON(http.StatusCreated, toStaffResponse(staff))
}

// UpdateStaff godoc
func (h *Handler) UpdateStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	staff, err := h.svc.Staff.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toStaffResponse(staff))
}

// GetStaff godoc
func (h *Handler) GetStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	staff, err := h.svc.Staff.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponse(staff))
}

// DeleteStaff godoc
func (h *Handler) DeleteStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if actor := optionalStaffID(c); actor != nil && *actor == id {
		RespondError(c, apperrors.WrapInvalidInput("自分自身を削除することはできません"))
		return
	}
	if err := h.svc.Staff.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetStaffPermissionGroups godoc
// GET /v1/masters/staffs/:id/permission-groups
func (h *Handler) GetStaffPermissionGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	groupIDs, err := h.svc.Staff.GetPermissionGroupIDs(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"group_ids": groupIDs})
}

// SetStaffPermissionGroups godoc
// PUT /v1/masters/staffs/:id/permission-groups
func (h *Handler) SetStaffPermissionGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	var req setStaffPermissionGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.GroupIDs == nil {
		req.GroupIDs = []uint64{}
	}
	if err := h.svc.Staff.SetPermissionGroupIDs(c.Request.Context(), id, req.GroupIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"group_ids": req.GroupIDs})
}

// GetStaffClinicAssignments godoc
// GET /v1/masters/staffs/:id/clinics
func (h *Handler) GetStaffClinicAssignments(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	clinicIDs := make([]uint64, 0, len(assignments))
	for i := range assignments {
		clinicIDs = append(clinicIDs, assignments[i].ClinicID)
	}
	c.JSON(http.StatusOK, gin.H{"clinic_ids": clinicIDs})
}

// SetStaffClinicAssignments godoc
// PUT /v1/masters/staffs/:id/clinics
func (h *Handler) SetStaffClinicAssignments(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	var req setStaffClinicAssignmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.ClinicIDs == nil {
		req.ClinicIDs = []uint64{}
	}

	// 削除→作成をサービス層のトランザクションで実行する
	if err := h.svc.Staff.SetClinicAssignments(c.Request.Context(), id, req.ClinicIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"clinic_ids": req.ClinicIDs})
}

// GetStaffExcludedReservationTypes godoc
// GET /v1/masters/staffs/:id/excluded-reservation-types
func (h *Handler) GetStaffExcludedReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	ids, err := h.svc.Staff.GetExcludedReservationTypeIDs(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservation_type_ids": ids})
}

// SetStaffExcludedReservationTypes godoc
// PUT /v1/masters/staffs/:id/excluded-reservation-types
func (h *Handler) SetStaffExcludedReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	var req setStaffExcludedReservationTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.ReservationTypeIDs == nil {
		req.ReservationTypeIDs = []uint64{}
	}
	if err := h.svc.Staff.SetExcludedReservationTypeIDs(c.Request.Context(), id, req.ReservationTypeIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservation_type_ids": req.ReservationTypeIDs})
}

// GetStaffCapableReservationTypes godoc
// GET /v1/masters/staffs/:id/capable-reservation-types
func (h *Handler) GetStaffCapableReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	ids, err := h.svc.Staff.GetCapableReservationTypeIDs(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservation_type_ids": ids})
}

// SetStaffCapableReservationTypes godoc
// PUT /v1/masters/staffs/:id/capable-reservation-types
func (h *Handler) SetStaffCapableReservationTypes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyStaffClinicMembership(c, clinicID, id) {
		return
	}
	var req setStaffCapableReservationTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.ReservationTypeIDs == nil {
		req.ReservationTypeIDs = []uint64{}
	}
	if err := h.svc.Staff.SetCapableReservationTypeIDs(c.Request.Context(), clinicID, id, req.ReservationTypeIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservation_type_ids": req.ReservationTypeIDs})
}

// ReorderStaffs godoc
func (h *Handler) ReorderStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Staff.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
