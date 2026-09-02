package trimming

import "github.com/animal-ekarte/backend/internal/httpapi"

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	return httpapi.ParseOptionalDateOnlyField(value, field)
}
