// Package handler provides HTTP handler implementations for PermissionGroup entity.
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- PermissionGroup ----

// GetPermissionGroup godoc
func (h *Handler) GetPermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	pg, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), id)
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
	c.JSON(http.StatusOK, toPermissionGroupResponseList(groups))
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

	pg := &model.PermissionGroup{
		ClinicID:    clinicID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}
	if err := h.svc.PermissionGroup.Create(c.Request.Context(), pg); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPermissionGroupResponse(pg))
}

// UpdatePermissionGroup godoc
func (h *Handler) UpdatePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	c.JSON(http.StatusOK, toPermissionGroupResponse(updated))
}

// DeletePermissionGroup godoc
func (h *Handler) DeletePermissionGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetPermissionGroupRules godoc
// PUTメソッドで全ルールを置き換える
func (h *Handler) SetPermissionGroupRules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	var req setPermissionGroupRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-140: 自分が所属するグループの master-permission edit を削除できないようにする
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	myGroupIDs, groupErr := h.svc.Staff.GetPermissionGroupIDs(c.Request.Context(), staffID)
	if groupErr == nil {
		isSelfGroup := false
		for _, gid := range myGroupIDs {
			if gid == id {
				isSelfGroup = true
				break
			}
		}
		if isSelfGroup {
			hasMasterPermEdit := false
			for _, r := range req.Rules {
				if r.Resource == string(model.ResourceMasterPermission) && r.CanEdit {
					hasMasterPermEdit = true
					break
				}
			}
			if !hasMasterPermEdit {
				RespondError(c, apperrors.WrapInvalidInput("自分が所属するグループの権限管理権限（master-permission edit）を削除することはできません"))
				return
			}
		}
	}

	// BUG-146: 入力バリデーション — 空文字・存在しないリソース名・重複を拒否
	seen := make(map[string]bool, len(req.Rules))
	for _, r := range req.Rules {
		if r.Resource == "" {
			RespondError(c, apperrors.WrapInvalidInput("リソース名が空です"))
			return
		}
		if !model.IsValidResource(r.Resource) {
			RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("無効なリソース名: %s", r.Resource)))
			return
		}
		if seen[r.Resource] {
			RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("リソース名が重複しています: %s", r.Resource)))
			return
		}
		seen[r.Resource] = true
	}

	// Convert request rules to model
	rules := make([]model.PermissionGroupRule, 0, len(req.Rules))
	for _, r := range req.Rules {
		rules = append(rules, model.PermissionGroupRule{
			Resource:  r.Resource,
			CanView:   r.CanView,
			CanCreate: r.CanCreate,
			CanEdit:   r.CanEdit,
			CanDelete: r.CanDelete,
		})
	}

	if err := h.svc.PermissionGroup.SetRules(c.Request.Context(), id, rules); err != nil {
		RespondError(c, err)
		return
	}

	// Return updated group with rules
	pg, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), id)
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
	var req reorderPermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.PermissionGroup.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}
