package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func optionalStringQueryFilter(value string) *string {
	return httpapi.OptionalString(value)
}

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	return httpapi.ParseOptionalDateOnlyField(value, field)
}

func parseOptionalDateTimeQueryFilter(value, field string) (*time.Time, error) {
	return httpapi.ParseOptionalDateTimeField(value, field)
}
