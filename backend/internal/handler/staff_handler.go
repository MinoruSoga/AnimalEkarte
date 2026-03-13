// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- Staff ----

// ListStaffs godoc
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}
	staffs, err := h.svc.Staff.List(c.Request.Context(), clinicID, role)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffs)
}

// createStaffRequest はスタッフ登録リクエスト。
// スタッフ情報とシステムアカウント情報を同時に受け取る。
type createStaffRequest struct {
	Name          string          `json:"name"           binding:"required"`
	StaffRole     model.StaffRole `json:"staff_role"     binding:"required"`
	Email         string          `json:"email"          binding:"required,email"`
	Password      string          `json:"password"       binding:"required,min=8"`
	LicenseNumber string          `json:"license_number"`
	JobTitleID    *uint64         `json:"job_title_id"`
	SortOrder     int             `json:"sort_order"`
}

// CreateStaff godoc
func (h *Handler) CreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staff, err := h.svc.Staff.CreateWithAccount(c.Request.Context(), &service.CreateStaffInput{
		ClinicID:      clinicID,
		Name:          req.Name,
		StaffRole:     req.StaffRole,
		Email:         req.Email,
		Password:      req.Password,
		LicenseNumber: req.LicenseNumber,
		JobTitleID:    req.JobTitleID,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, staff)
}

type updateStaffInput struct {
	Name          string `json:"name"`
	StaffRole     string `json:"staff_role"`
	LicenseNumber string `json:"license_number"`
	JobTitleID    *uint64 `json:"job_title_id"`
	SortOrder     int    `json:"sort_order"`
	IsActive      *bool  `json:"is_active"`
}

// UpdateStaff godoc
func (h *Handler) UpdateStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateStaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff := &model.Staff{
		ID:            id,
		ClinicID:      clinicID,
		Name:          input.Name,
		LicenseNumber: input.LicenseNumber,
		JobTitleID:    input.JobTitleID,
		SortOrder:     input.SortOrder,
	}
	if input.StaffRole != "" {
		staff.StaffRole = model.StaffRole(input.StaffRole)
	}
	if input.IsActive != nil {
		staff.IsActive = *input.IsActive
	}

	if err := h.svc.Staff.Update(c.Request.Context(), staff); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staff)
}

// DeleteStaff godoc
func (h *Handler) DeleteStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Staff.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
