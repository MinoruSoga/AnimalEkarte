// Package staff provides staff HTTP handlers.
package staff

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
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
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	passwordReplacement := req.Password != nil && *req.Password != ""
	var credentialActorID uint64
	if passwordReplacement {
		credentialActorID, ok = httpapi.ExtractStaffID(c)
		if !ok {
			return
		}
		if h.hasPermission == nil ||
			!h.hasPermission(
				c,
				string(model.ResourceMasterPermission),
				"edit",
			) {
			RespondError(c, apperrors.WrapForbidden("forbidden"))
			return
		}
	}

	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	authorizedClinicIDs, ok := httpapi.ExtractClinicIDs(c)
	if !ok {
		return
	}
	input := req.toServiceInput()
	input.AuthorizedClinicIDs = authorizedClinicIDs
	input.IsSystemAdmin = isSystemAdmin
	if actor := optionalStaffID(c); actor != nil {
		input.ActorStaffID = *actor
	}
	if passwordReplacement {
		input.CredentialAudit = &CredentialMutationAudit{
			ClinicID:      clinicID,
			ActorStaffID:  credentialActorID,
			TargetStaffID: id,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.Header.Get("User-Agent"),
		}
	}

	staff, err := h.svc.Staff.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponse(staff))
}

// GetStaff godoc
func (h *Handler) GetStaff(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	var (
		staff *model.Staff
		err   error
	)
	if peekedSystemAdmin(c) {
		staff, err = h.svc.Staff.GetByID(c.Request.Context(), id)
	} else {
		staff, err = h.svc.Staff.GetByIDInClinic(c.Request.Context(), clinicID, id)
	}
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
	if err := h.svc.Staff.Delete(c.Request.Context(), clinicID, id, peekedSystemAdmin(c)); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetStaffPermissionGroups godoc
// GET /v1/masters/staffs/:id/permission-groups
func (h *Handler) GetStaffPermissionGroups(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	groupIDs, err := h.svc.Staff.GetPermissionGroupIDs(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffPermissionGroupsResponse{GroupIDs: groupIDs})
}

// SetStaffPermissionGroups godoc
// PUT /v1/masters/staffs/:id/permission-groups
func (h *Handler) SetStaffPermissionGroups(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	var req setStaffPermissionGroupsRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if req.GroupIDs == nil {
		req.GroupIDs = []uint64{}
	}
	if err := h.svc.Staff.SetPermissionGroupIDs(c.Request.Context(), clinicID, id, req.GroupIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffPermissionGroupsResponse(req))
}

// GetStaffClinicAssignments godoc
// GET /v1/masters/staffs/:id/clinics
func (h *Handler) GetStaffClinicAssignments(c *gin.Context) {
	_, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}

	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	authorizedClinicIDs, ok := httpapi.ExtractClinicIDs(c)
	if !ok {
		return
	}
	authorizedClinicIDs, ok = httpapi.FilterClinicIDsForPermission(
		c,
		authorizedClinicIDs,
		string(model.ResourceMasterStaff),
		"view",
	)
	if !ok {
		return
	}
	if h.svc.StaffClinicAssignment == nil {
		RespondError(c, apperrors.WrapInternalServerError(
			"staff clinic assignment service is not configured",
		))
		return
	}

	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffClinicAssignmentsResponse{
		ClinicIDs: visibleStaffClinicIDs(id, assignments, authorizedClinicIDs, isSystemAdmin),
	})
}

func visibleStaffClinicIDs(
	staffID uint64,
	assignments []model.StaffClinicAssignment,
	authorizedClinicIDs []uint64,
	isSystemAdmin bool,
) []uint64 {
	authorized := make(map[uint64]struct{}, len(authorizedClinicIDs))
	if !isSystemAdmin {
		for _, clinicID := range authorizedClinicIDs {
			authorized[clinicID] = struct{}{}
		}
	}

	visible := make([]uint64, 0, len(assignments))
	seen := make(map[uint64]struct{}, len(assignments))
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.StaffID != staffID || assignment.ClinicID == 0 || assignment.DeletedAt.Valid {
			continue
		}
		if _, duplicate := seen[assignment.ClinicID]; duplicate {
			continue
		}
		if !isSystemAdmin {
			if _, allowed := authorized[assignment.ClinicID]; !allowed {
				continue
			}
		}
		seen[assignment.ClinicID] = struct{}{}
		visible = append(visible, assignment.ClinicID)
	}
	slices.Sort(visible)
	return visible
}

// SetStaffClinicAssignments godoc
// PUT /v1/masters/staffs/:id/clinics
func (h *Handler) SetStaffClinicAssignments(c *gin.Context) {
	_, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	var req setStaffClinicAssignmentsRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	normalizedClinicIDs := make([]uint64, 0, len(req.ClinicIDs))
	seenClinicIDs := make(map[uint64]struct{}, len(req.ClinicIDs))
	for _, clinicID := range req.ClinicIDs {
		if _, duplicate := seenClinicIDs[clinicID]; duplicate {
			continue
		}
		seenClinicIDs[clinicID] = struct{}{}
		normalizedClinicIDs = append(normalizedClinicIDs, clinicID)
	}

	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	authorizedClinicIDs, ok := httpapi.ExtractClinicIDs(c)
	if !ok {
		return
	}
	authorizedClinicIDs, ok = httpapi.FilterClinicIDsForPermission(
		c,
		authorizedClinicIDs,
		string(model.ResourceMasterStaff),
		"edit",
	)
	if !ok {
		return
	}
	if !isSystemAdmin {
		for _, clinicID := range normalizedClinicIDs {
			if !slices.Contains(authorizedClinicIDs, clinicID) {
				RespondError(c, apperrors.WrapForbidden("cannot assign staff outside authorized clinics"))
				return
			}
		}
	}

	if err := h.svc.Staff.SetClinicAssignments(c.Request.Context(), &SetClinicAssignmentsInput{
		StaffID:             id,
		ClinicIDs:           normalizedClinicIDs,
		AuthorizedClinicIDs: authorizedClinicIDs,
		IsSystemAdmin:       isSystemAdmin,
	}); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffClinicAssignmentsResponse{ClinicIDs: normalizedClinicIDs})
}

// GetStaffExcludedReservationTypes godoc
// GET /v1/masters/staffs/:id/excluded-reservation-types
func (h *Handler) GetStaffExcludedReservationTypes(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	ids, err := h.svc.Staff.GetExcludedReservationTypeIDs(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffReservationTypesResponse{ReservationTypeIDs: ids})
}

// SetStaffExcludedReservationTypes godoc
// PUT /v1/masters/staffs/:id/excluded-reservation-types
func (h *Handler) SetStaffExcludedReservationTypes(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	var req setStaffExcludedReservationTypesRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if req.ReservationTypeIDs == nil {
		req.ReservationTypeIDs = []uint64{}
	}
	if err := h.svc.Staff.SetExcludedReservationTypeIDs(c.Request.Context(), clinicID, id, req.ReservationTypeIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffReservationTypesResponse(req))
}

// GetStaffCapableReservationTypes godoc
// GET /v1/masters/staffs/:id/capable-reservation-types
func (h *Handler) GetStaffCapableReservationTypes(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	ids, err := h.svc.Staff.GetCapableReservationTypeIDs(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffReservationTypesResponse{ReservationTypeIDs: ids})
}

// SetStaffCapableReservationTypes godoc
// PUT /v1/masters/staffs/:id/capable-reservation-types
func (h *Handler) SetStaffCapableReservationTypes(c *gin.Context) {
	clinicID, id, ok := h.resolveStaffWithClinic(c)
	if !ok {
		return
	}
	var req setStaffCapableReservationTypesRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if req.ReservationTypeIDs == nil {
		req.ReservationTypeIDs = []uint64{}
	}
	if err := h.svc.Staff.SetCapableReservationTypeIDs(c.Request.Context(), clinicID, id, req.ReservationTypeIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffReservationTypesResponse(req))
}

// ReorderStaffs godoc
func (h *Handler) ReorderStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := bindStaffJSON(c, &req); err != nil {
		RespondError(c, err)
		return
	}
	if err := h.svc.Staff.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
