package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListUsers godoc
// GET /users
// clinic_id は JWT から取得し、クエリパラメータによる上書きは禁止（テナント分離）
func (h *Handler) ListUsers(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	accounts, err := h.svc.UserAccount.ListUsers(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := make([]userResponse, 0, len(accounts))
	for i := range accounts {
		items = append(items, toUserResponse(&accounts[i]))
	}
	c.JSON(http.StatusOK, items)
}

// CreateUser godoc
// POST /users
func (h *Handler) CreateUser(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.CreateUserAccountInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		UserType:    model.UserType(req.UserType),
		StaffID:     req.StaffID,
		IsMain:      req.IsMain,
	}

	account, err := h.svc.UserAccount.CreateUser(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(account))
}

// GetUser godoc
// GET /users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	data, err := h.svc.UserAccount.GetWithMemberships(c.Request.Context(), strconv.FormatUint(id, 10))
	if err != nil {
		RespondError(c, err)
		return
	}
	groupIDs, err := h.svc.UserAccount.GetPermissionGroupIDs(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	memberships := make([]userMembershipResponse, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		memberships = append(memberships, userMembershipResponse{
			ClinicID: strconv.FormatUint(m.ClinicID, 10),
			IsMain:   m.IsMain,
		})
	}
	c.JSON(http.StatusOK, userDetailResponse{
		userResponse:       toUserResponse(&data.UserAccount),
		Memberships:        memberships,
		PermissionGroupIDs: groupIDs,
	})
}

// SetUserPermissionGroups godoc
// PUT /users/:id/permission-groups — ユーザーへのグループ割当を全置換する
// clinic_admin 以上のみ実行可能
func (h *Handler) SetUserPermissionGroups(c *gin.Context) {
	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType == model.UserTypeStaff {
		RespondError(c, apperrors.WrapForbidden("permission group assignment requires clinic_admin or above"))
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req struct {
		GroupIDs []uint64 `json:"group_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.UserAccount.SetPermissionGroups(c.Request.Context(), id, service.SetPermissionGroupsInput{
		GroupIDs: req.GroupIDs,
	}); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateUser godoc
// PATCH /users/:id
// user_type の変更は system_admin のみ許可
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.UpdateUserAccountInput{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Status:      req.Status,
		StaffID:     req.StaffID,
	}
	if req.UserType != nil {
		userType, ok := extractUserType(c)
		if !ok {
			return
		}
		if userType != model.UserTypeSystemAdmin {
			RespondError(c, apperrors.WrapForbidden("user_type change requires system_admin"))
			return
		}
		ut := model.UserType(*req.UserType)
		input.UserType = &ut
	}

	if err := h.svc.UserAccount.UpdateUser(c.Request.Context(), id, input); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteUser godoc
// DELETE /users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.UserAccount.DeleteUser(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterUserRoutes はユーザー管理関連のルートを登録する
func (h *Handler) RegisterUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.GET("", h.ListUsers)
	users.POST("", h.CreateUser)
	users.GET("/:id", h.GetUser)
	users.PATCH("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
	users.PUT("/:id/permission-groups", h.SetUserPermissionGroups)
}
