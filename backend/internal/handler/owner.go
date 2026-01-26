package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllOwners godoc
// @Summary 飼い主一覧取得
// @Description 登録されている飼い主の一覧を取得します
// @Tags owners
// @Accept json
// @Produce json
// @Param page query int false "ページ番号" default(1)
// @Param limit query int false "1ページあたりの件数" default(10)
// @Param search query string false "検索キーワード（名前、メール、電話番号）"
// @Success 200 {array} model.Owner
// @Failure 500 {object} map[string]string
// @Router /owners [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllOwners(c *gin.Context) {
	ctx := c.Request.Context()
	owners, err := h.svc.GetAllOwners(ctx)
	if err != nil {
		h.handleError(c, err, "owner", "")
		return
	}
	c.JSON(http.StatusOK, owners)
}

// GetOwnerByID godoc
// @Summary      Get an owner by ID
// @Description  Get details of a specific owner
// @Tags         owners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Owner ID"
// @Success      200  {object}  model.Owner
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /owners/{id} [get]
func (h *Handler) GetOwnerByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	owner, err := h.svc.GetOwnerByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "owner", id)
		return
	}
	c.JSON(http.StatusOK, owner)
}

// CreateOwner godoc
// @Summary      Create a new owner
// @Description  Create a new owner with the provided details
// @Tags         owners
// @Accept       json
// @Produce      json
// @Param        owner  body      model.CreateOwnerRequest  true  "Owner Data"
// @Success      201    {object}  model.Owner
// @Failure      400    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /owners [post]
func (h *Handler) CreateOwner(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidInput, "owner", "")
		return
	}

	owner, err := h.svc.CreateOwner(ctx, &req)
	if err != nil {
		h.handleError(c, err, "owner", "")
		return
	}
	c.JSON(http.StatusCreated, owner)
}

// UpdateOwner godoc
// @Summary      Update an owner
// @Description  Update details of an existing owner
// @Tags         owners
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Owner ID"
// @Param        owner  body      model.UpdateOwnerRequest  true  "Owner Data"
// @Success      200    {object}  model.Owner
// @Failure      400    {object}  ErrorResponse
// @Failure      404    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /owners/{id} [put]
func (h *Handler) UpdateOwner(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidInput, "owner", id)
		return
	}

	owner, err := h.svc.UpdateOwner(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "owner", id)
		return
	}
	c.JSON(http.StatusOK, owner)
}

// DeleteOwner godoc
// @Summary      Delete an owner
// @Description  Delete an owner by ID
// @Tags         owners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Owner ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /owners/{id} [delete]
func (h *Handler) DeleteOwner(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteOwner(ctx, id); err != nil {
		h.handleError(c, err, "owner", id)
		return
	}
	c.Status(http.StatusNoContent)
}
