// Package handler provides HTTP handler implementations for PermissionGroup entity.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// marshalAuditJSON は監査ログ用に値をJSONバイト列にシリアライズするヘルパー。
// nil の場合は nil を返す。エラー時は nil を返す（監査ログはベストエフォート）。
func marshalAuditJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// ---- PermissionGroup ----

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

	createInput := service.CreatePermissionGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}
	pg, err := h.svc.PermissionGroup.Create(c.Request.Context(), clinicID, &createInput)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ作成
	if staffID, ok := extractStaffID(c); ok {
		if auditErr := h.svc.Audit.Log(c.Request.Context(), &model.AuditLog{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupCreate,
			Resource:   "permission_group",
			ResourceID: &pg.ID,
			NewValue:   marshalAuditJSON(pg),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}); auditErr != nil {
			slog.ErrorContext(c.Request.Context(), "failed to log permission group creation", slog.String("error", auditErr.Error()))
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

	input := &service.UpdatePermissionGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	updated, err := h.svc.PermissionGroup.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ更新
	if staffID, ok := extractStaffID(c); ok {
		if auditErr := h.svc.Audit.Log(c.Request.Context(), &model.AuditLog{
			ClinicID:   &clinicID,
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupUpdate,
			Resource:   "permission_group",
			ResourceID: &id,
			NewValue:   marshalAuditJSON(updated),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}); auditErr != nil {
			slog.ErrorContext(c.Request.Context(), "failed to log permission group update", slog.String("error", auditErr.Error()))
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
	oldPG, getErr := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
	if getErr != nil && !errors.Is(getErr, apperrors.ErrNotFound) {
		slog.WarnContext(c.Request.Context(), "failed to fetch old permission group for audit",
			slog.String("error", getErr.Error()))
	}

	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限グループ削除
	if staffID, ok := extractStaffID(c); ok {
		auditLog := &model.AuditLog{
			ActorID:    &staffID,
			ActorType:  "staff",
			Action:     model.AuditActionPermissionGroupDelete,
			Resource:   "permission_group",
			ResourceID: &id,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.Header.Get("User-Agent"),
		}
		if oldPG != nil {
			auditLog.ClinicID = &oldPG.ClinicID
			auditLog.OldValue = marshalAuditJSON(oldPG)
		}
		if auditErr := h.svc.Audit.Log(c.Request.Context(), auditLog); auditErr != nil {
			slog.ErrorContext(c.Request.Context(), "failed to log permission group deletion", slog.String("error", auditErr.Error()))
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

	// Convert request rules to service Input DTO
	inputRules := make([]service.SetPermissionGroupRulesInput, 0, len(req.Rules))
	for _, r := range req.Rules {
		inputRules = append(inputRules, service.SetPermissionGroupRulesInput{
			Resource:  string(r.Resource),
			CanView:   r.CanView,
			CanCreate: r.CanCreate,
			CanEdit:   r.CanEdit,
			CanDelete: r.CanDelete,
		})
	}

	if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, inputRules, staffID); err != nil {
		RespondError(c, err)
		return
	}

	// 監査ログ: 権限ルール更新
	if auditErr := h.svc.Audit.Log(c.Request.Context(), &model.AuditLog{
		ActorID:    &staffID,
		ActorType:  "staff",
		Action:     model.AuditActionPermissionRulesUpdate,
		Resource:   "permission_group_rules",
		ResourceID: &id,
		NewValue:   marshalAuditJSON(inputRules),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.Header.Get("User-Agent"),
	}); auditErr != nil {
		slog.ErrorContext(c.Request.Context(), "failed to log permission rules update", slog.String("error", auditErr.Error()))
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
