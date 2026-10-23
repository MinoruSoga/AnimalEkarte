package medicalrecord

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const hospitalizationSnapshotMaxRows = 10000
const hospitalizationSnapshotTimeout = 10 * time.Second

type hospitalizationSnapshotLister interface {
	listSnapshot(context.Context, uint64, listHospitalizationFilters) ([]model.Hospitalization, int64, error)
}
type hospitalizationSnapshotFinder interface {
	findSnapshot(context.Context, uint64, listHospitalizationFilters) ([]model.Hospitalization, int64, error)
}

func (s *hospitalizationService) listSnapshot(ctx context.Context, clinicID uint64, filters listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
	finder, ok := s.hospRepo.(hospitalizationSnapshotFinder)
	if !ok {
		return nil, 0, apperrors.WrapNotImplemented("入院一覧の一括取得は未設定です")
	}
	rows, total, err := finder.findSnapshot(ctx, clinicID, filters)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to read hospitalization snapshot")
	}
	return rows, total, nil
}

func (h *HospitalizationHandler) maybeListHospitalizationSnapshot(c *gin.Context, clinicID uint64, filters listHospitalizationFilters) bool {
	values, present := c.Request.URL.Query()["snapshot"]
	if !present {
		return false
	}
	raw := c.Query("snapshot")
	if len(values) != 1 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("snapshot は1つだけ指定してください"))
		return true
	}
	if raw != "true" && raw != "false" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("snapshot は true または false を指定してください"))
		return true
	}
	if raw == "false" {
		return false
	}
	lister, ok := h.service.(hospitalizationSnapshotLister)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapNotImplemented("入院一覧の一括取得は未設定です"))
		return true
	}
	rows, total, err := lister.listSnapshot(c.Request.Context(), clinicID, filters)
	if err != nil {
		httpapi.RespondError(c, err)
		return true
	}
	projections, err := h.service.StayProjections(c.Request.Context(), clinicID, rows)
	if err != nil {
		httpapi.RespondError(c, err)
		return true
	}
	c.JSON(http.StatusOK, struct {
		Data     []HospitalizationResponse `json:"data"`
		Total    int64                     `json:"total"`
		Page     int                       `json:"page"`
		Limit    int                       `json:"limit"`
		Snapshot bool                      `json:"snapshot"`
	}{hospitalizationResponses(rows, projections), total, 1, hospitalizationSnapshotMaxRows, true})
	return true
}

func (r *hospitalizationRepository) findSnapshot(ctx context.Context, clinicID uint64, filters listHospitalizationFilters) ([]model.Hospitalization, int64, error) {
	rows := make([]model.Hospitalization, 0)
	ctx, cancel := context.WithTimeout(ctx, hospitalizationSnapshotTimeout)
	defer cancel()
	var total int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := hospitalizationListQuery(tx, clinicID, filters).Count(&total).Error; err != nil {
			return apperrors.FromGORM(err, "hospitalization", "")
		}
		if total > hospitalizationSnapshotMaxRows {
			return apperrors.WrapInvalidInput("入院が10000件を超えています。条件を絞り込んでください")
		}
		if err := hospitalizationListPreloads(hospitalizationListQuery(tx, clinicID, filters), clinicID).
			Order(hospitalizationListOrder).
			Limit(hospitalizationSnapshotMaxRows).
			Find(&rows).Error; err != nil {
			return apperrors.FromGORM(err, "hospitalization", "")
		}
		if int64(len(rows)) != total {
			return apperrors.WrapInternalServerError("hospitalization snapshot row count mismatch")
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
