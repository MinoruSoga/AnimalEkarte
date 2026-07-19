package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DEPRECATED facade (BE9-2B): moved to internal/httpapi.LocalTime/LocalTimePtr/
// LocalTimeRFC3339. See context_helpers.go's file header for the rationale; delete once
// BE9-2F migrates every remaining internal/handler file to internal/httpapi directly.

func localTime(t time.Time) time.Time {
	return httpapi.LocalTime(t)
}

func localTimePtr(t *time.Time) *time.Time {
	return httpapi.LocalTimePtr(t)
}

func localTimeRFC3339(t time.Time) string {
	return httpapi.LocalTimeRFC3339(t)
}
