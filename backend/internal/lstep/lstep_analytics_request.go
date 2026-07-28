package lstep

import (
	"net/url"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type lstepMonthlyDeliveryStatsQuery struct {
	YearMonth string
}

func newLstepMonthlyDeliveryStatsQuery(values url.Values) (lstepMonthlyDeliveryStatsQuery, error) {
	query := lstepMonthlyDeliveryStatsQuery{YearMonth: values.Get("year_month")}
	if query.YearMonth == "" {
		return lstepMonthlyDeliveryStatsQuery{}, apperrors.WrapInvalidInput("year_month クエリパラメータは必須です")
	}
	return query, nil
}

type lstepVisitConversionQuery struct {
	YearMonth string
	Days      string
}

func newLstepVisitConversionQuery(values url.Values) (lstepVisitConversionQuery, error) {
	query := lstepVisitConversionQuery{
		YearMonth: values.Get("year_month"),
		Days:      values.Get("days"),
	}
	if query.YearMonth == "" {
		return lstepVisitConversionQuery{}, apperrors.WrapInvalidInput("year_month クエリパラメータは必須です")
	}
	return query, nil
}

func (q lstepVisitConversionQuery) toDays() (int, error) {
	if q.Days == "" {
		return 30, nil
	}
	days, err := strconv.Atoi(q.Days)
	if err != nil || days < 1 || days > maxVisitConversionDays {
		return 0, apperrors.WrapInvalidInput("days は 1 以上 365 以下の整数で指定してください")
	}
	return days, nil
}
