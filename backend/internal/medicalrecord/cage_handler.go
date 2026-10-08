package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// CageHandler serves the cage master HTTP boundary. Moved from internal/handler (BE9-2D ⑥ Batch C).
type CageHandler struct {
	service CageService
	// merger は EMR-306 連結書込み面。配線サービスが CageMerger を実装しない
	// 場合（未配線モック含む）は nil で、merge/unmerge は 500 fail-closed。
	merger CageMerger
	// occupancy は EMR-306 連結ピッカーの占有読み取り面。同上に nil fail-closed。
	occupancy CageOccupancyReader
}

// NewCageHandler initializes a CageHandler.
func NewCageHandler(service CageService) *CageHandler {
	merger, _ := service.(CageMerger)
	occupancy, _ := service.(CageOccupancyReader)
	return &CageHandler{service: service, merger: merger, occupancy: occupancy}
}

// ---- Cage ----

// ListCages godoc
func (h *CageHandler) ListCages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterHospitalization), "view") {
		return
	}
	cageType := newListCageQuery(c.Request.URL.Query()).toServiceFilter()
	cages, err := h.service.List(c.Request.Context(), clinicID, cageType)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(cages, toCageResponse))
}

// GetCage godoc
func (h *CageHandler) GetCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterHospitalization), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	cage, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCageResponse(cage))
}

// CreateCage godoc
func (h *CageHandler) CreateCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	cage, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/cages/%d", cage.ID))
	c.JSON(http.StatusCreated, toCageResponse(cage))
}

// UpdateCage godoc
func (h *CageHandler) UpdateCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	cage, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toCageResponse(cage))
}

// ReorderCages godoc
func (h *CageHandler) ReorderCages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCage godoc
func (h *CageHandler) DeleteCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// OccupiedCageIDs godoc — EMR-306 review: 連結ピッカー占有判定専用の読み取り。
// GET /hospitalizations は hospitalization:view を要求するため、
// master-hospitalization:edit のみを持つケージ編集者が占有状況を取得できず
// fail-closed 側の全ケージ選択不可に陥る。本エンドポイントは merge 権限と同じ
// master-hospitalization:edit 配下で最小限の id 集合のみを返す（ページングなし）。
func (h *CageHandler) OccupiedCageIDs(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if h.occupancy == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("cage occupancy service is not configured"))
		return
	}
	ids, err := h.occupancy.OccupiedCageIDs(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if ids == nil {
		ids = []uint64{}
	}
	c.JSON(http.StatusOK, cageOccupancyResponse{OccupiedCageIDs: ids})
}

// MergeCage godoc — EMR-306: 代表ケージ :id のもとに member_cage_ids を一時連結する。
// 権限はルート側の perm ミドルウェア（master-hospitalization:edit）が担う。
func (h *CageHandler) MergeCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req mergeCageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if h.merger == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("cage merge service is not configured"))
		return
	}
	cage, err := h.merger.Merge(c.Request.Context(), clinicID, id, req.MemberCageIDs)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toCageResponse(cage))
}

// UnmergeCage godoc — EMR-306: 代表ケージ :id が率いる連結グループを解除する。
func (h *CageHandler) UnmergeCage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if h.merger == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("cage merge service is not configured"))
		return
	}
	if err := h.merger.Unmerge(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
