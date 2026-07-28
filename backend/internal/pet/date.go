package pet

import (
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type jsonDate struct {
	time.Time
}

func (d *jsonDate) UnmarshalJSON(data []byte) error {
	value := strings.Trim(string(data), `"`)
	if value == "null" || value == "" {
		return nil
	}
	parsed, err := httpapi.ParseFlexibleDate(value)
	if err != nil {
		return apperrors.WrapInvalidInput(httpapi.FlexibleDateInvalidInputMsg)
	}
	d.Time = parsed
	return nil
}

func jsonDatePtr(date *jsonDate) *time.Time {
	if date == nil {
		return nil
	}
	value := date.Time
	return &value
}
