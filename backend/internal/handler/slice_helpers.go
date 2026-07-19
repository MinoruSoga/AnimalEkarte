package handler

import "github.com/animal-ekarte/backend/internal/httpapi"

// DEPRECATED facade (BE9-2B): moved to internal/httpapi.MapSlice/ReorderRequest/NilIfEmpty.
// See context_helpers.go's file header for the rationale; delete once BE9-2F migrates every
// remaining internal/handler file to internal/httpapi directly.

func mapSlice[M, R any](items []M, f func(*M) R) []R {
	return httpapi.MapSlice(items, f)
}

// reorderRequest is a stable alias for httpapi.ReorderRequest so the ~26 existing call sites
// in this package keep compiling unchanged.
type reorderRequest = httpapi.ReorderRequest

func nilIfEmpty(s string) *string {
	return httpapi.NilIfEmpty(s)
}
