package handler

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// tagNamePattern は手動付与可能なタグ名の形式（英数字・アンダースコア・ハイフン、1〜100文字）。
var tagNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,100}$`)

// GetOwnerLstepTags godoc
// GET /owners/:id/lstep/tags — 飼い主のLステップタグ一覧を返す（BE-019）。
// LINE未連携時は 200 + tags:[] を返す。
func (h *Handler) GetOwnerLstepTags(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	result, err := h.svc.LstepTag.GetOwnerTags(c.Request.Context(), clinicID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepTagsResponse(result))
}

// PostOwnerLstepTag godoc
// POST /owners/:id/lstep/tags — 飼い主に手動でLステップタグを付与する（BE-019）。
func (h *Handler) PostOwnerLstepTag(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req postOwnerLstepTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if !tagNamePattern.MatchString(req.TagName) {
		RespondError(c, apperrors.WrapInvalidInput("tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）"))
		return
	}

	var actorIDPtr *uint64
	if actorID, ok := extractStaffID(c); ok {
		actorIDPtr = &actorID
	}

	if err := h.svc.LstepTag.AddOwnerTag(c.Request.Context(), clinicID, ownerID, req.TagName, actorIDPtr); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, lstepAddTagResponse{TagName: req.TagName, Added: true})
}

// DeleteOwnerLstepTag godoc
// DELETE /owners/:id/lstep/tags/:tag_name — 飼い主からLステップタグを手動解除する（BE-019）。
// 存在しないタグでも冪等に 204 を返す。
func (h *Handler) DeleteOwnerLstepTag(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	tagName := c.Param("tag_name")
	if tagName == "" {
		RespondError(c, apperrors.WrapInvalidInput("tag_name は必須です"))
		return
	}

	var actorIDPtr *uint64
	if actorID, ok := extractStaffID(c); ok {
		actorIDPtr = &actorID
	}

	if err := h.svc.LstepTag.RemoveOwnerTag(c.Request.Context(), clinicID, ownerID, tagName, actorIDPtr); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
