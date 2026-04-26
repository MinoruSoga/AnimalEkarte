package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
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
}

// ownerAggregationListResponse は顧客集計一覧レスポンス。
type ownerAggregationListResponse struct {
	Total   int                        `json:"total"`
	Page    int                        `json:"page"`
	PerPage int                        `json:"per_page"`
	Owners  []ownerAggregationResponse `json:"owners"`
}

func toOwnerAggregationResponse(item service.OwnerAggregationItem) ownerAggregationResponse {
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
	}
	if item.LastVisitDate != nil {
		s := item.LastVisitDate.Format("2006-01-02")
		r.LastVisitDate = &s
	}
	if item.FirstVisitDate != nil {
		s := item.FirstVisitDate.Format("2006-01-02")
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

func toOwnerAggregationListResponse(result *service.ListOwnerAggregationResult) ownerAggregationListResponse {
	owners := make([]ownerAggregationResponse, 0, len(result.Items))
	for _, item := range result.Items {
		owners = append(owners, toOwnerAggregationResponse(item))
	}
	return ownerAggregationListResponse{
		Total:   result.Total,
		Page:    result.Page,
		PerPage: result.PerPage,
		Owners:  owners,
	}
}

// ListOwnerAggregation godoc
// GET /api/v1/owners/ltv — 累計診療費・来院回数でソート・フィルタ可能な顧客集計一覧を返す（BE-010）。
func (h *Handler) ListOwnerAggregation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	if pageErr != nil || page < 1 {
		RespondError(c, apperrors.WrapInvalidInput("page は1以上の整数で指定してください"))
		return
	}
	perPage, ppErr := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if ppErr != nil || perPage < 1 || perPage > 200 {
		RespondError(c, apperrors.WrapInvalidInput("per_page は1〜200の範囲で指定してください"))
		return
	}

	// min_total_fee / max_total_fee は FE エイリアス（min_total_amount 優先）
	var minTotalAmount, maxTotalAmount *int64
	minAmountKey := "min_total_amount"
	if c.Query(minAmountKey) == "" && c.Query("min_total_fee") != "" {
		minAmountKey = "min_total_fee"
	}
	if s := c.Query(minAmountKey); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("min_total_amount / min_total_fee は0以上の整数で指定してください"))
			return
		}
		minTotalAmount = &v
	}
	maxAmountKey := "max_total_amount"
	if c.Query(maxAmountKey) == "" && c.Query("max_total_fee") != "" {
		maxAmountKey = "max_total_fee"
	}
	if s := c.Query(maxAmountKey); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("max_total_amount / max_total_fee は0以上の整数で指定してください"))
			return
		}
		maxTotalAmount = &v
	}

	var minVisitCount *int64
	if s := c.Query("min_visit_count"); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("min_visit_count は0以上の整数で指定してください"))
			return
		}
		minVisitCount = &v
	}

	// has_line は FE エイリアス（line_linked 優先）
	lineLinked := c.Query("line_linked") == "true" || c.Query("has_line") == "true"

	// sort + order パラメータ（FE: sort=total_fee&order=desc）
	sort := resolveAggregationSort(c.Query("sort"))
	order := c.DefaultQuery("order", "desc")

	// year パラメータ（AGG-BE-001）
	var year *int
	if s := c.Query("year"); s != "" {
		v, pErr := strconv.Atoi(s)
		if pErr != nil || v < 1900 || v > 2100 {
			RespondError(c, apperrors.WrapInvalidInput("year は1900〜2100の整数で指定してください"))
			return
		}
		year = &v
	}

	// from / to パラメータ（YYYY-MM-DD形式、AGG-BE-001）
	var from, to *string
	if s := c.Query("from"); s != "" {
		from = &s
	}
	if s := c.Query("to"); s != "" {
		to = &s
	}

	// amount_basis パラメータ（AGG-BE-001）
	amountBasis := c.DefaultQuery("amount_basis", "gross_total_amount")

	// include_zero パラメータ（AGG-BE-001）
	includeZero := c.Query("include_zero") == "true"

	// search パラメータ（飼い主名部分一致、AGG-BE-001）
	search := c.Query("search")

	// period_preset パラメータ（AGG-BE-002）
	periodPreset := c.Query("period_preset")

	// max_visit_count パラメータ（AGG-BE-002）
	var maxVisitCount *int64
	if s := c.Query("max_visit_count"); s != "" {
		v, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil || v < 0 {
			RespondError(c, apperrors.WrapInvalidInput("max_visit_count は0以上の整数で指定してください"))
			return
		}
		maxVisitCount = &v
	}

	// last_visit_bucket パラメータ（AGG-BE-003）
	lastVisitBucket := c.Query("last_visit_bucket")

	// include_no_visit パラメータ（AGG-BE-003）
	includeNoVisit := c.Query("include_no_visit") == "true"

	// metric パラメータ（AGG-BE-004）— annual_sales | visit_count | last_visit
	metric := c.DefaultQuery("metric", "annual_sales")

	result, err := h.svc.Aggregation.ListOwnerAggregation(c.Request.Context(), clinicID, service.ListOwnerAggregationInput{
		Sort:            sort,
		MinTotalAmount:  minTotalAmount,
		MaxTotalAmount:  maxTotalAmount,
		MinVisitCount:   minVisitCount,
		LineLinked:      lineLinked,
		Page:            page,
		PerPage:         perPage,
		Year:            year,
		From:            from,
		To:              to,
		AmountBasis:     amountBasis,
		IncludeZero:     includeZero,
		Search:          search,
		PeriodPreset:    periodPreset,
		MaxVisitCount:   maxVisitCount,
		LastVisitBucket: lastVisitBucket,
		IncludeNoVisit:  includeNoVisit,
		Order:           order,
		Metric:          metric,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerAggregationListResponse(result))
}

// RegisterAggregationRoutes は顧客集計関連のルートを登録する（BE-010）。
func (h *Handler) RegisterAggregationRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinic_id")
	clinics.GET("/owners/aggregations", h.ListOwnerAggregation)
}
