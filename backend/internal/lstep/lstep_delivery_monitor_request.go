package lstep

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

type lstepDeliveryMonitorQuery struct {
	ClinicID    uint64
	From        time.Time
	To          time.Time
	TriggerType string
	Status      string
	Page        int
	PerPage     int
}

func newLstepDeliveryMonitorSummaryQuery(clinicID uint64, values url.Values, now time.Time) (lstepDeliveryMonitorQuery, error) {
	from, to, err := parseLstepDeliveryMonitorDateRange(values, now)
	if err != nil {
		return lstepDeliveryMonitorQuery{}, err
	}
	return lstepDeliveryMonitorQuery{
		ClinicID:    clinicID,
		From:        from,
		To:          to,
		TriggerType: values.Get("trigger_type"),
	}, nil
}

func newLstepDeliveryMonitorLogsQuery(clinicID uint64, values url.Values, now time.Time) (lstepDeliveryMonitorQuery, error) {
	from, to, err := parseLstepDeliveryMonitorDateRange(values, now)
	if err != nil {
		return lstepDeliveryMonitorQuery{}, err
	}
	page, err := parseLstepDeliveryMonitorPage(values.Get("page"))
	if err != nil {
		return lstepDeliveryMonitorQuery{}, err
	}
	perPage, err := parseLstepDeliveryMonitorPerPage(values.Get("per_page"))
	if err != nil {
		return lstepDeliveryMonitorQuery{}, err
	}
	return lstepDeliveryMonitorQuery{
		ClinicID:    clinicID,
		From:        from,
		To:          to,
		TriggerType: values.Get("trigger_type"),
		Status:      values.Get("status"),
		Page:        page,
		PerPage:     perPage,
	}, nil
}

func (q *lstepDeliveryMonitorQuery) toSummaryServiceInput() GetDeliveryMonitorSummaryInput {
	return GetDeliveryMonitorSummaryInput{
		ClinicID:    q.ClinicID,
		From:        q.From,
		To:          q.To,
		TriggerType: q.TriggerType,
	}
}

func (q *lstepDeliveryMonitorQuery) toLogsServiceInput() *GetDeliveryMonitorLogsInput {
	return &GetDeliveryMonitorLogsInput{
		ClinicID:    q.ClinicID,
		From:        q.From,
		To:          q.To,
		TriggerType: q.TriggerType,
		Status:      q.Status,
		Page:        q.Page,
		PerPage:     q.PerPage,
	}
}

func parseLstepDeliveryMonitorDateRange(values url.Values, now time.Time) (from, to time.Time, err error) {
	defaultDate := now.In(config.JST).Format(time.DateOnly)
	fromStr := lstepDeliveryMonitorQueryDefault(values, "from", defaultDate)
	toStr := lstepDeliveryMonitorQueryDefault(values, "to", defaultDate)

	from, err = time.ParseInLocation(time.DateOnly, fromStr, config.JST)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("from は YYYY-MM-DD 形式で指定してください")
	}
	to, err = time.ParseInLocation(time.DateOnly, toStr, config.JST)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("to は YYYY-MM-DD 形式で指定してください")
	}
	return from, to.AddDate(0, 0, 1), nil
}

func parseLstepDeliveryMonitorPage(value string) (int, error) {
	if value == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 {
		return 0, apperrors.WrapInvalidInput("page は1以上の整数で指定してください")
	}
	return page, nil
}

func parseLstepDeliveryMonitorPerPage(value string) (int, error) {
	if value == "" {
		return 20, nil
	}
	perPage, err := strconv.Atoi(value)
	if err != nil || perPage < 1 || perPage > 100 {
		return 0, apperrors.WrapInvalidInput("per_page は1〜100の範囲で指定してください")
	}
	return perPage, nil
}

func lstepDeliveryMonitorQueryDefault(values url.Values, key, fallback string) string {
	if value := values.Get(key); value != "" {
		return value
	}
	return fallback
}
