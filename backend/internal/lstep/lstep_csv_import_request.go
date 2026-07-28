package lstep

import (
	"net/url"
	"strconv"
)

type listLstepCsvImportsQuery struct {
	Limit string
}

func newListLstepCsvImportsQuery(values url.Values) listLstepCsvImportsQuery {
	return listLstepCsvImportsQuery{Limit: values.Get("limit")}
}

func (q listLstepCsvImportsQuery) toLimit() int {
	if q.Limit == "" {
		return 20
	}
	limit, err := strconv.Atoi(q.Limit)
	if err != nil || limit < 1 || limit > 100 {
		return 20
	}
	return limit
}
