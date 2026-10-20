package medicalrecord

// hospitalization_cage_occupancy.go — EMR-461: 入院/ホテル登録フォームの空きケージ候補用の
// 最小 read。連結ピッカー専用の GET /masters/cages/occupancy（master-hospitalization:edit）
// とは別面で、同院の入院作成／編集権限だけで占有ケージ ID 集合を返す。患者・入院情報は
// 返さない。候補表示用であり、保存時は create/update/cage-swap の割当ガードが現在状態を
// 再検査して競合を 409 で拒否する。

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// HospitalizationCageOccupancyReader は入院フォームの空き候補 read の面。
// HospitalizationService には足さず（非所有モックの互換維持）、handler が型アサーションで取得する。
type HospitalizationCageOccupancyReader interface {
	OccupiedCageIDsForAssignment(ctx context.Context, clinicID uint64, excludeHospitalizationID *uint64) ([]uint64, error)
}

// hospitalizationCageOccupancyResponse は GET /hospitalizations/cage-occupancy の応答。
type hospitalizationCageOccupancyResponse struct {
	OccupiedCageIDs []uint64 `json:"occupied_cage_ids"`
}

// OccupiedCageIDsForAssignment は医院内の占有ケージ ID（構成員 ID を含む）を返す。
// excludeHospitalizationID は編集中の入院で、その占有行だけを除外する（自己占有を
// 他滞在と誤認しない）。同医院に存在しない入院 ID は NotFound。
func (s *hospitalizationService) OccupiedCageIDsForAssignment(ctx context.Context, clinicID uint64, excludeHospitalizationID *uint64) ([]uint64, error) {
	mergeRepo, ok := s.cageRepo.(CageMergeRepository)
	if !ok {
		return nil, apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	if excludeHospitalizationID != nil {
		existing, err := s.hospRepo.FindByID(ctx, clinicID, *excludeHospitalizationID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to verify excluded hospitalization")
		}
		if existing == nil {
			return nil, apperrors.WrapNotFound("hospitalization", strconv.FormatUint(*excludeHospitalizationID, 10))
		}
	}
	ids, err := mergeRepo.ListOccupiedCageIDsExcludingHospitalization(ctx, clinicID, excludeHospitalizationID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list occupied cages")
	}
	if ids == nil {
		ids = []uint64{}
	}
	return ids, nil
}

// requireSelectedClinicHospitalizationWriteGrant は選択医院で hospitalization の
// create または edit を持つかを判定する（httpapi.RequireSelectedClinicGrant の any-of 版）。
// middleware の view は checker を attach する土台で、他院の grant では通さない。
func requireSelectedClinicHospitalizationWriteGrant(c *gin.Context) bool {
	if isAdmin, ok := httpapi.PeekIsSystemAdmin(c); ok && isAdmin {
		return true
	}
	check, ok := httpapi.PeekClinicPermissionChecker(c)
	clinicID, hasClinic := httpapi.PeekClinicID(c)
	if ok && hasClinic && clinicID != 0 {
		resource := string(model.ResourceHospitalization)
		if check(c, clinicID, resource, "create") || check(c, clinicID, resource, "edit") {
			return true
		}
	}
	httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
	return false
}

// GetCageOccupancy godoc
// @Summary 入院フォーム用の占有ケージID一覧
// @Description EMR-461。選択医院の未削除 admitted/reserved 入院が載るケージ ID を返す。exclude_hospitalization_id の入院行は除外する。
// @Tags Hospitalizations
// @Produce json
// @Param exclude_hospitalization_id query int false "編集中の入院ID"
// @Success 200 {object} hospitalizationCageOccupancyResponse
// @Router /hospitalizations/cage-occupancy [get]
func (h *HospitalizationHandler) GetCageOccupancy(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !requireSelectedClinicHospitalizationWriteGrant(c) {
		return
	}
	reader, ok := h.service.(HospitalizationCageOccupancyReader)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("hospitalization cage occupancy service is not configured"))
		return
	}
	var exclude *uint64
	if raw := c.Query("exclude_hospitalization_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			httpapi.RespondError(c, apperrors.WrapInvalidInput("exclude_hospitalization_id must be a positive id"))
			return
		}
		exclude = &id
	}
	ids, err := reader.OccupiedCageIDsForAssignment(c.Request.Context(), clinicID, exclude)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, hospitalizationCageOccupancyResponse{OccupiedCageIDs: ids})
}
