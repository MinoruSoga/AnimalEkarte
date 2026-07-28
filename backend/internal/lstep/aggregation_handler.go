package lstep

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ownerAggregationResponse は顧客集計の1件レスポンス。
type ownerAggregationResponse struct {
	OwnerID            string  `json:"owner_id"`
	OwnerName          string  `json:"owner_name"`
	TotalAmount        int64   `json:"total_amount"`
	TotalFee           int64   `json:"total_fee"` // total_amount の FE エイリアス
	TotalVisitCount    int64   `json:"total_visit_count"`
	AnnualVisitCount   int64   `json:"annual_visit_count"`
	LastVisitDate      *string `json:"last_visit_date,omitempty"`
	FirstVisitDate     *string `json:"first_visit_date,omitempty"`
	AnnualAmount       *int64  `json:"annual_amount,omitempty"`         // AGG-BE-001
	BillingCount       *int64  `json:"billing_count,omitempty"`         // AGG-BE-001
	PeriodVisitCount   *int64  `json:"period_visit_count,omitempty"`    // AGG-BE-001/002
	DaysSinceLastVisit *int    `json:"days_since_last_visit,omitempty"` // AGG-BE-003
	LastVisitBucket    *string `json:"last_visit_bucket,omitempty"`     // AGG-BE-003
	CPMStage           string  `json:"cpm_stage,omitempty"`             // ISSUE-006 タグ同期側と同一判定の CPM ステージ
}

// ownerAggregationListResponse は顧客集計一覧レスポンス。
type ownerAggregationListResponse struct {
	Total   int                        `json:"total"`
	Page    int                        `json:"page"`
	PerPage int                        `json:"per_page"`
	Owners  []ownerAggregationResponse `json:"owners"`
}

func toOwnerAggregationResponse(item *OwnerAggregationItem) ownerAggregationResponse {
	r := ownerAggregationResponse{
		OwnerID:            strconv.FormatUint(item.OwnerID, 10),
		OwnerName:          item.OwnerName,
		TotalAmount:        item.TotalAmount,
		TotalFee:           item.TotalAmount,
		TotalVisitCount:    item.TotalVisitCount,
		AnnualVisitCount:   item.AnnualVisitCount,
		AnnualAmount:       item.AnnualAmount,
		BillingCount:       item.BillingCount,
		PeriodVisitCount:   item.PeriodVisitCount,
		DaysSinceLastVisit: item.DaysSinceLastVisit,
		LastVisitBucket:    item.LastVisitBucket,
		CPMStage:           item.CPMStage,
	}
	if item.LastVisitDate != nil {
		s := item.LastVisitDate.In(time.Local).Format(time.DateOnly)
		r.LastVisitDate = &s
	}
	if item.FirstVisitDate != nil {
		s := item.FirstVisitDate.In(time.Local).Format(time.DateOnly)
		r.FirstVisitDate = &s
	}
	return r
}

// resolveAggregationSort は FE の sort パラメータをリポジトリの sort フィールドに変換する。
// order は別途パラメータで渡される。
func resolveAggregationSort(sortParam string) string {
	switch sortParam {
	case "total_fee", "total_amount":
		return "total_amount"
	case "annual_amount":
		return "annual_amount"
	case "annual_visit_count":
		return "annual_visit_count"
	case "visit_count", "period_visit_count":
		return "visit_count"
	case "total_visit_count":
		return "total_visit_count"
	case "last_visit_date", "last_visit":
		return "last_visit_date"
	case "days_since_last_visit":
		return "days_since_last_visit"
	case "owner_name":
		return "owner_name"
	default:
		return "total_amount"
	}
}

func toOwnerAggregationListResponse(result *ListOwnerAggregationResult) ownerAggregationListResponse {
	owners := make([]ownerAggregationResponse, 0, len(result.Items))
	for i := range result.Items {
		owners = append(owners, toOwnerAggregationResponse(&result.Items[i]))
	}
	return ownerAggregationListResponse{
		Total:   result.Total,
		Page:    result.Page,
		PerPage: result.PerPage,
		Owners:  owners,
	}
}

// ListOwnerAggregation godoc
// GET /api/v1/clinics/:clinic_id/owners/aggregations — 累計診療費・来院回数でソート・フィルタ可能な顧客集計一覧を返す（BE-010）。
func (h *Handler) ListOwnerAggregation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	query, err := newOwnerAggregationQuery(c.Request.URL.Query())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	result, err := h.aggregation.ListOwnerAggregation(c.Request.Context(), clinicID, query.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerAggregationListResponse(result))
}

// RegisterAggregationRoutes は顧客集計関連のルートを登録する（BE-010）。
func (h *Handler) RegisterAggregationRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/owners/aggregations", h.requirePermission(string(model.ResourceOwners), "view"), h.ListOwnerAggregation)
}
