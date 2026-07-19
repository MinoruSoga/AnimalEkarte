package handler

import "github.com/animal-ekarte/backend/internal/httpapi"

// DEPRECATED facade (BE9-2B): moved to internal/httpapi.ParseBindError. See
// context_helpers.go's file header for the rationale; delete once BE9-2F migrates every
// remaining internal/handler file to internal/httpapi directly.
func parseBindError(err error) string {
	return httpapi.ParseBindError(err)
}
