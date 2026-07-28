package lstep

import (
	"net/url"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type lstepOwnersByTagQuery struct {
	TagName   string
	NameQuery string
	Format    string
	Page      int
	PerPage   int
}

func newLstepOwnersByTagQuery(values url.Values) (lstepOwnersByTagQuery, error) {
	query := lstepOwnersByTagQuery{
		TagName:   values.Get("tag"),
		NameQuery: values.Get("q"),
		Format:    values.Get("format"),
	}
	if query.TagName == "" {
		return lstepOwnersByTagQuery{}, apperrors.WrapInvalidInput("tag パラメータは必須です")
	}
	if query.isCSV() {
		return query, nil
	}

	page, err := parseLstepOwnersByTagPage(values.Get("page"))
	if err != nil {
		return lstepOwnersByTagQuery{}, err
	}
	perPage, err := parseLstepOwnersByTagPerPage(values.Get("per_page"))
	if err != nil {
		return lstepOwnersByTagQuery{}, err
	}
	query.Page = page
	query.PerPage = perPage
	return query, nil
}

func (q lstepOwnersByTagQuery) isCSV() bool {
	return q.Format == "csv"
}

func (q lstepOwnersByTagQuery) toServiceInput() ListOwnersByTagInput {
	return ListOwnersByTagInput{
		TagName:   q.TagName,
		NameQuery: q.NameQuery,
		Page:      q.Page,
		PerPage:   q.PerPage,
	}
}

func parseLstepOwnersByTagPage(value string) (int, error) {
	if value == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 {
		return 0, apperrors.WrapInvalidInput("page は1以上の整数で指定してください")
	}
	return page, nil
}

func parseLstepOwnersByTagPerPage(value string) (int, error) {
	if value == "" {
		return 20, nil
	}
	perPage, err := strconv.Atoi(value)
	if err != nil || perPage < 1 || perPage > 100 {
		return 0, apperrors.WrapInvalidInput("per_page は1〜100の範囲で指定してください")
	}
	return perPage, nil
}
