package clinic

import "net/url"

type ListClinicHolidaysQuery struct {
	YearMonth string
}

func NewListClinicHolidaysQuery(values url.Values) ListClinicHolidaysQuery {
	return ListClinicHolidaysQuery{YearMonth: values.Get("year_month")}
}
