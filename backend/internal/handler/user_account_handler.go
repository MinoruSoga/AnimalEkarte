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
// system_admin は任意ユーザーを取得可能。それ以外は同一クリニックに属するユーザーのみ。
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

	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		clinicID, ok := extractClinicID(c)
		if !ok {
			return
		}
		belongs := false
		for _, m := range data.Memberships {
			if m.ClinicID == clinicID {
				belongs = true
				break
			}
		}
		if !belongs {
			RespondError(c, apperrors.WrapForbidden("cannot access user from other clinic"))
			return
		}
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
// clinic_admin 以上のみ実行可能。system_admin 以外は同一クリニックのユーザーのみ変更可。
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

	// system_admin 以外はテナント検証
	if userType != model.UserTypeSystemAdmin {
		if ok := h.verifyUserClinicMembership(c, id); !ok {
			return
		}
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

// verifyUserClinicMembership は対象ユーザーが JWT クリニックに属しているかを検証するヘルパー。
// 属していない場合は 403 を返して false を返す。
func (h *Handler) verifyUserClinicMembership(c *gin.Context, userID uint64) bool {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return false
	}
	data, err := h.svc.UserAccount.GetWithMemberships(c.Request.Context(), strconv.FormatUint(userID, 10))
	if err != nil {
		RespondError(c, err)
		return false
	}
	for _, m := range data.Memberships {
		if m.ClinicID == clinicID {
			return true
		}
	}
	RespondError(c, apperrors.WrapForbidden("cannot access user from other clinic"))
	return false
}

// UpdateUser godoc
// PATCH /users/:id
// system_admin 以外は同一クリニックのユーザーのみ更新可。user_type の変更は system_admin のみ許可。
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		if ok := h.verifyUserClinicMembership(c, id); !ok {
			return
		}
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
// system_admin 以外は同一クリニックのユーザーのみ削除可。
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}

	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		if ok := h.verifyUserClinicMembership(c, id); !ok {
			return
		}
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
