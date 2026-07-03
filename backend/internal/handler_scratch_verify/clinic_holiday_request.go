package handler

import "net/url"

type listClinicHolidaysQuery struct {
	YearMonth string
}

func newListClinicHolidaysQuery(values url.Values) listClinicHolidaysQuery {
	return listClinicHolidaysQuery{YearMonth: values.Get("year_month")}
}
