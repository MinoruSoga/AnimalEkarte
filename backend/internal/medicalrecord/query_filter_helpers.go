package medicalrecord

import "github.com/animal-ekarte/backend/internal/httpapi"

func optionalStringQueryFilter(value string) *string {
	return httpapi.OptionalString(value)
}

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	return httpapi.ParseOptionalDateOnlyField(value, field)
}

func nilIfEmpty(s string) *string {
	return httpapi.OptionalString(s)
}
