// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

// ListStaffs godoc
// @Summary スタッフ一覧取得
// @Description クリニックに登録されているスタッフの一覧を返す。roleクエリで職種絞り込み可能。
// @Tags StaffMasters
// @Produce json
// @Security BearerAuth
// @Param role query string false "スタッフ職種でフィルタリング（例: veterinarian, nurse, trimmer, reception, manager）"
// @Success 200 {array} model.Staff
// @Failure 500 {object} map[string]string
// @Router /masters/staffs [get]
func (h *Handler) ListStaffs(c *gin.Context) {
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}
	staffs, err := h.svc.Staff.List(c.Request.Context(), role)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, staffs)
}

// CreateStaff godoc
// @Summary スタッフ作成
// @Description 新規スタッフを登録する。
// @Tags StaffMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param staff body model.Staff true "スタッフ情報"
// @Success 201 {object} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/staffs [post]
func (h *Handler) CreateStaff(c *gin.Context) {
	var input model.Staff
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Staff.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateStaff godoc
// @Summary スタッフ更新
// @Description 指定IDのスタッフ情報を更新する。
// @Tags StaffMasters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "スタッフID（UUID）"
// @Param staff body model.Staff true "更新するスタッフ情報"
// @Success 200 {object} model.Staff
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/staffs/{id} [put]
func (h *Handler) UpdateStaff(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Staff
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Staff.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteStaff godoc
// @Summary スタッフ削除
// @Description 指定IDのスタッフを論理削除する。
// @Tags StaffMasters
// @Security BearerAuth
// @Param id path string true "スタッフID（UUID）"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/staffs/{id} [delete]
func (h *Handler) DeleteStaff(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Staff.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
