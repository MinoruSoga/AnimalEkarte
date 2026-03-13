// Package handler provides HTTP handler implementations for TrimmingCourse and TrimmingOption entities.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse ----

// ListTrimmingCourses godoc
// @Summary トリミングコース一覧取得
// @Description 登録されているトリミングコースの一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.TrimmingCourse
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-courses [get]
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	courses, err := h.svc.TrimmingCourse.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, courses)
}

// CreateTrimmingCourse godoc
// @Summary トリミングコース作成
// @Description 新しいトリミングコースを作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.TrimmingCourse true "トリミングコース情報"
// @Success 201 {object} model.TrimmingCourse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-courses [post]
func (h *Handler) CreateTrimmingCourse(c *gin.Context) {
	var input model.TrimmingCourse
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.TrimmingCourse.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateTrimmingCourse godoc
// @Summary トリミングコース更新
// @Description 指定IDのトリミングコースを更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングコースID"
// @Param request body model.TrimmingCourse true "トリミングコース情報"
// @Success 200 {object} model.TrimmingCourse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-courses/{id} [put]
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.TrimmingCourse
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.TrimmingCourse.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteTrimmingCourse godoc
// @Summary トリミングコース削除
// @Description 指定IDのトリミングコースを削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングコースID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-courses/{id} [delete]
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.TrimmingCourse.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- TrimmingOption ----

// ListTrimmingOptions godoc
// @Summary トリミングオプション一覧取得
// @Description 登録されているトリミングオプションの一覧を返す
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.TrimmingOption
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-options [get]
func (h *Handler) ListTrimmingOptions(c *gin.Context) {
	options, err := h.svc.TrimmingOption.List(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, options)
}

// CreateTrimmingOption godoc
// @Summary トリミングオプション作成
// @Description 新しいトリミングオプションを作成する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.TrimmingOption true "トリミングオプション情報"
// @Success 201 {object} model.TrimmingOption
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-options [post]
func (h *Handler) CreateTrimmingOption(c *gin.Context) {
	var input model.TrimmingOption
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.TrimmingOption.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateTrimmingOption godoc
// @Summary トリミングオプション更新
// @Description 指定IDのトリミングオプションを更新する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングオプションID"
// @Param request body model.TrimmingOption true "トリミングオプション情報"
// @Success 200 {object} model.TrimmingOption
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-options/{id} [put]
func (h *Handler) UpdateTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.TrimmingOption
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.TrimmingOption.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteTrimmingOption godoc
// @Summary トリミングオプション削除
// @Description 指定IDのトリミングオプションを削除する
// @Tags Masters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングオプションID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /masters/trimming-options/{id} [delete]
func (h *Handler) DeleteTrimmingOption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.TrimmingOption.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
