// Package handler provides HTTP handler implementations for DiagnosisCategory and DiagnosisName entities.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DiagnosisCategory ----

// ListDiagnosisCategories godoc
// @Summary 診断カテゴリ一覧取得
// @Description 登録されている診断カテゴリの一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.DiagnosisCategory
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-categories [get]
func (h *Handler) ListDiagnosisCategories(c *gin.Context) {
	categories, err := h.svc.DiagnosisCategory.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}

// CreateDiagnosisCategory godoc
// @Summary 診断カテゴリ作成
// @Description 新しい診断カテゴリを作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DiagnosisCategory true "診断カテゴリ情報"
// @Success 201 {object} model.DiagnosisCategory
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-categories [post]
func (h *Handler) CreateDiagnosisCategory(c *gin.Context) {
	var input model.DiagnosisCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.DiagnosisCategory.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateDiagnosisCategory godoc
// @Summary 診断カテゴリ更新
// @Description 指定IDの診断カテゴリを更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診断カテゴリUUID"
// @Param request body model.DiagnosisCategory true "診断カテゴリ情報"
// @Success 200 {object} model.DiagnosisCategory
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-categories/{id} [put]
func (h *Handler) UpdateDiagnosisCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.DiagnosisCategory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.DiagnosisCategory.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteDiagnosisCategory godoc
// @Summary 診断カテゴリ削除
// @Description 指定IDの診断カテゴリを削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診断カテゴリUUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-categories/{id} [delete]
func (h *Handler) DeleteDiagnosisCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisCategory.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- DiagnosisName ----

// ListDiagnosisNames godoc
// @Summary 診断名一覧取得
// @Description 登録されている診断名の一覧を返す。category_idクエリパラメータでカテゴリ絞り込み可能
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category_id query string false "診断カテゴリUUID（省略時は全件返す）"
// @Success 200 {array} model.DiagnosisName
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-names [get]
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
	var names []model.DiagnosisName
	var err error

	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, parseErr := uuid.Parse(catIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		names, err = h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), catID)
	} else {
		names, err = h.svc.DiagnosisName.List(c.Request.Context())
	}

	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, names)
}

// CreateDiagnosisName godoc
// @Summary 診断名作成
// @Description 新しい診断名を作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DiagnosisName true "診断名情報"
// @Success 201 {object} model.DiagnosisName
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-names [post]
func (h *Handler) CreateDiagnosisName(c *gin.Context) {
	var input model.DiagnosisName
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.DiagnosisName.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateDiagnosisName godoc
// @Summary 診断名更新
// @Description 指定IDの診断名を更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診断名UUID"
// @Param request body model.DiagnosisName true "診断名情報"
// @Success 200 {object} model.DiagnosisName
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-names/{id} [put]
func (h *Handler) UpdateDiagnosisName(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.DiagnosisName
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.DiagnosisName.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteDiagnosisName godoc
// @Summary 診断名削除
// @Description 指定IDの診断名を削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "診断名UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/diagnosis-names/{id} [delete]
func (h *Handler) DeleteDiagnosisName(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DiagnosisName.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
