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
// GET /users?clinic_id=xxx
func (h *Handler) ListUsers(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	// クエリパラメータで clinic_id が明示的に指定されていれば上書き
	if s := c.Query("clinic_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid clinic_id"))
			return
		}
		clinicID = id
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
	memberships := make([]userMembershipResponse, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		memberships = append(memberships, userMembershipResponse{
			ClinicID: strconv.FormatUint(m.ClinicID, 10),
			IsMain:   m.IsMain,
		})
	}
	c.JSON(http.StatusOK, userDetailResponse{
		userResponse: toUserResponse(&data.UserAccount),
		Memberships:  memberships,
	})
}

// UpdateUser godoc
// PATCH /users/:id
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

// GetUserPermissions godoc
// GET /users/:id/permissions?clinic_id=xxx
func (h *Handler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	perms, err := h.svc.UserAccount.GetPermissions(c.Request.Context(), userID, clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	items := make([]userPermissionResponse, 0, len(perms))
	for _, p := range perms {
		items = append(items, userPermissionResponse{Permission: string(p.Permission)})
	}
	c.JSON(http.StatusOK, items)
}

// SetUserPermissions godoc
// PUT /users/:id/permissions
func (h *Handler) SetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req setPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := &service.SetPermissionsInput{Permissions: req.Permissions}
	if err := h.svc.UserAccount.SetPermissions(c.Request.Context(), userID, clinicID, input); err != nil {
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
	users.GET("/:id/permissions", h.GetUserPermissions)
	users.PUT("/:id/permissions", h.SetUserPermissions)
}
