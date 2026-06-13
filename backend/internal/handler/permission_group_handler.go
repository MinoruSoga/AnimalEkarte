// Package handler provides HTTP handler implementations for PermissionGroup entity.
package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- PermissionGroup ----

func (h *Handler) RegisterPermissionGroupRoutes(masters *gin.RouterGroup) {
	masters.GET("/permission-groups", h.RequirePermission(string(model.ResourceMasterPermission), "view"), h.ListPermissionGroups)
	masters.GET("/permission-groups/:id", h.RequirePermission(string(model.ResourceMasterPermission), "view"), h.GetPermissionGroup)
	masters.POST("/permission-groups", h.RequirePermission(string(model.ResourceMasterPermission), "create"), h.CreatePermissionGroup)
	masters.PATCH("/permission-groups/reorder", h.RequirePermission(string(model.ResourceMasterPermission), "edit"), h.ReorderPermissionGroups)
	masters.PATCH("/permission-groups/:id", h.RequirePermission(string(model.ResourceMasterPermission), "edit"), h.UpdatePermissionGroup)
	masters.DELETE("/permission-groups/:id", h.RequirePermission(string(model.ResourceMasterPermission), "delete"), h.DeletePermissionGroup)
	masters.PUT("/permission-groups/:id/rules", h.RequirePermission(string(model.ResourceMasterPermission), "edit"), h.SetPermissionGroupRules)
}

// GetPermissionGroup godoc
func (h *Handler) GetPermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	pg, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPermissionGroupResponse(pg))
}

// ListPermissionGroups godoc
func (h *Handler) ListPermissionGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	groups, err := h.svc.PermissionGroup.List(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(groups, toPermissionGroupResponse))
}

// CreatePermissionGroup godoc
func (h *Handler) CreatePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createPermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	pg, err := h.svc.PermissionGroup.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ作成（#122: 最小JSON）
	if staffID, ok := extractStaffID(c); ok {
		if err := h.svc.Audit.LogEntry(c.Request.Context(), &service.AuditLogInput{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupCreate,
			Resource:   "permission_group",
			ResourceID: &pg.ID,
			NewValue:   map[string]any{"name": pg.Name},
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to write audit log for permission group create",
				"error", err, "resource_id", pg.ID, "actor_id", staffID)
		}
	}

	c.Header("Location", fmt.Sprintf("/v1/masters/permission-groups/%d", pg.ID))
	c.JSON(http.StatusCreated, toPermissionGroupResponse(pg))
}

// UpdatePermissionGroup godoc
func (h *Handler) UpdatePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updatePermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// 監査ログ用: 更新前の値を取得（#122 best-effort）
	oldPGForAudit, _ := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)

	updated, err := h.svc.PermissionGroup.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ更新（#122: OldValue/NewValue 最小JSON）
	if staffID, ok := extractStaffID(c); ok {
		auditInput := &service.AuditLogInput{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupUpdate,
			Resource:   "permission_group",
			ResourceID: &id,
			NewValue:   map[string]any{"name": updated.Name},
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}
		if oldPGForAudit != nil {
			auditInput.OldValue = map[string]any{"name": oldPGForAudit.Name}
		}
		if err := h.svc.Audit.LogEntry(c.Request.Context(), auditInput); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to write audit log for permission group update",
				"error", err, "resource_id", id, "actor_id", staffID)
		}
	}

	c.JSON(http.StatusOK, toPermissionGroupResponse(updated))
}

// DeletePermissionGroup godoc
func (h *Handler) DeletePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	// 削除前に old value を取得（監査ログ用）
	oldPG, _ := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)

	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ削除
	if staffID, ok := extractStaffID(c); ok {
		auditInput := service.AuditLogInput{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupDelete,
			Resource:   "permission_group",
			ResourceID: &id,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}
		if oldPG != nil {
			auditInput.OldValue = map[string]any{"name": oldPG.Name}
		}
		if err := h.svc.Audit.LogEntry(c.Request.Context(), &auditInput); err != nil {
			slog.WarnContext(c.Request.Context(), "failed to write audit log for permission group delete",
				"error", err, "resource_id", id, "actor_id", staffID)
		}
	}

	c.Status(http.StatusNoContent)
}

// SetPermissionGroupRules godoc
// PUTメソッドで全ルールを置き換える
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	// TASK-016: グループがこのクリニックに属することを確認（横断テナント書き換え防止）
	if _, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}

	var req setPermissionGroupRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-140 / BUG-146: バリデーションは service 層に委譲
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}

	inputRules := req.toServiceInput()
	if err := h.svc.PermissionGroup.UpdateRules(c.Request.Context(), id, inputRules, staffID); err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限ルール更新（#122: 最小JSON）
	if err := h.svc.Audit.LogEntry(c.Request.Context(), &service.AuditLogInput{
		ClinicID:   &clinicID,
		ActorID:    &staffID,
		ActorType:  "staff",
		Action:     model.AuditActionPermissionRulesUpdate,
		Resource:   "permission_group_rules",
		ResourceID: &id,
		NewValue:   map[string]any{"rule_count": len(inputRules)},
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.Header.Get("User-Agent"),
	}); err != nil {
		slog.WarnContext(c.Request.Context(), "failed to write audit log for permission group rules update",
			"error", err, "resource_id", id, "actor_id", staffID)
	}

	// Return updated group with rules
	pg, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPermissionGroupResponse(pg))
}

// ReorderPermissionGroups godoc
func (h *Handler) ReorderPermissionGroups(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.PermissionGroup.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
