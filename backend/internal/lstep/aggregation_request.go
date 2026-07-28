package lstep

import (
	"net/url"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type ownerAggregationQuery struct {
	Sort            string
	MinTotalAmount  *int64
	MaxTotalAmount  *int64
	MinVisitCount   *int64
	LineLinked      bool
	Page            int
	PerPage         int
	Year            *int
	From            *string
	To              *string
	AmountBasis     string
	IncludeZero     bool
	Search          string
	PeriodPreset    string
	MaxVisitCount   *int64
	LastVisitBucket string
	IncludeNoVisit  bool
	Order           string
	Metric          string
	CPMStage        string
}

func newOwnerAggregationQuery(values url.Values) (ownerAggregationQuery, error) {
	page, err := parseOwnerAggregationPage(values.Get("page"))
	if err != nil {
		return ownerAggregationQuery{}, err
	}
	perPage, err := parseOwnerAggregationPerPage(values.Get("per_page"))
	if err != nil {
		return ownerAggregationQuery{}, err
	}

	minTotalAmount, err := parseOwnerAggregationAmount(
		firstOwnerAggregationQueryValue(values, "min_amount", "min_total_amount", "min_total_fee"),
		"min_amount / min_total_amount / min_total_fee は0以上の整数で指定してください",
	)
	if err != nil {
		return ownerAggregationQuery{}, err
	}
	maxTotalAmount, err := parseOwnerAggregationAmount(
		firstOwnerAggregationQueryValue(values, "max_amount", "max_total_amount", "max_total_fee"),
		"max_amount / max_total_amount / max_total_fee は0以上の整数で指定してください",
	)
	if err != nil {
		return ownerAggregationQuery{}, err
	}
	minVisitCount, err := parseOwnerAggregationAmount(values.Get("min_visit_count"), "min_visit_count は0以上の整数で指定してください")
	if err != nil {
		return ownerAggregationQuery{}, err
	}
	maxVisitCount, err := parseOwnerAggregationAmount(values.Get("max_visit_count"), "max_visit_count は0以上の整数で指定してください")
	if err != nil {
		return ownerAggregationQuery{}, err
	}
	year, err := parseOwnerAggregationYear(values.Get("year"))
	if err != nil {
		return ownerAggregationQuery{}, err
	}

	return ownerAggregationQuery{
		Sort:            resolveAggregationSort(values.Get("sort")),
		MinTotalAmount:  minTotalAmount,
		MaxTotalAmount:  maxTotalAmount,
		MinVisitCount:   minVisitCount,
		LineLinked:      values.Get("line_linked") == "true" || values.Get("has_line") == "true",
		Page:            page,
		PerPage:         perPage,
		Year:            year,
		From:            optionalStringQueryFilter(values.Get("from")),
		To:              optionalStringQueryFilter(values.Get("to")),
		AmountBasis:     ownerAggregationQueryDefault(values, "amount_basis", "gross_total_amount"),
		IncludeZero:     values.Get("include_zero") == "true",
		Search:          values.Get("search"),
		PeriodPreset:    values.Get("period_preset"),
		MaxVisitCount:   maxVisitCount,
		LastVisitBucket: values.Get("last_visit_bucket"),
		IncludeNoVisit:  values.Get("include_no_visit") == "true",
		Order:           ownerAggregationQueryDefault(values, "order", "desc"),
		Metric:          ownerAggregationQueryDefault(values, "metric", "annual_sales"),
		CPMStage:        values.Get("cpm_stage"),
	}, nil
}

func (q *ownerAggregationQuery) toServiceInput() *ListOwnerAggregationInput {
	return &ListOwnerAggregationInput{
		Sort:            q.Sort,
		MinTotalAmount:  q.MinTotalAmount,
		MaxTotalAmount:  q.MaxTotalAmount,
		MinVisitCount:   q.MinVisitCount,
		LineLinked:      q.LineLinked,
		Page:            q.Page,
		PerPage:         q.PerPage,
		Year:            q.Year,
		From:            q.From,
		To:              q.To,
		AmountBasis:     q.AmountBasis,
		IncludeZero:     q.IncludeZero,
		Search:          q.Search,
		PeriodPreset:    q.PeriodPreset,
		MaxVisitCount:   q.MaxVisitCount,
		LastVisitBucket: q.LastVisitBucket,
		IncludeNoVisit:  q.IncludeNoVisit,
		Order:           q.Order,
		Metric:          q.Metric,
		CPMStage:        q.CPMStage,
	}
}

func parseOwnerAggregationPage(value string) (int, error) {
	if value == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 {
		return 0, apperrors.WrapInvalidInput("page は1以上の整数で指定してください")
	}
	return page, nil
}

func parseOwnerAggregationPerPage(value string) (int, error) {
	if value == "" {
		return 50, nil
	}
	perPage, err := strconv.Atoi(value)
	if err != nil || perPage < 1 || perPage > 200 {
		return 0, apperrors.WrapInvalidInput("per_page は1〜200の範囲で指定してください")
	}
	return perPage, nil
}

func parseOwnerAggregationAmount(value, message string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, apperrors.WrapInvalidInput(message)
	}
	return &parsed, nil
}

func parseOwnerAggregationYear(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}
	year, err := strconv.Atoi(value)
	if err != nil || year < 1900 || year > 2100 {
		return nil, apperrors.WrapInvalidInput("year は1900〜2100の整数で指定してください")
	}
	return &year, nil
}

func firstOwnerAggregationQueryValue(values url.Values, keys ...string) string {
	for _, key := range keys {
		if value := values.Get(key); value != "" {
			return value
		}
	}
	return ""
}

func ownerAggregationQueryDefault(values url.Values, key, fallback string) string {
	if value := values.Get(key); value != "" {
		return value
	}
	return fallback
}

func optionalStringQueryFilter(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
