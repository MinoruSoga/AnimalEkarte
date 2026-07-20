package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DEPRECATED facade (BE9-2C): moved to internal/httpapi.ParseIDParam/ParsePagination/
// ParsePaginationWithMax/ParseUUIDParam (BE9-2A target:httpapi, ADR-006 — this file was the
// last of the 6-file httpapi extraction deferred by BE9-2B because the manualarticle pilot
// didn't need it; internal/medicalrecord's master-CRUD slice does). See
// context_helpers.go's file header for the rationale; delete once BE9-2F migrates every
// remaining internal/handler file to internal/httpapi directly.
//
// httpapi.ParseOptionalUint64Query and httpapi.DefaultMaxPaginationLimit have no facade
// here: this package's sole caller (diagnosis_handler.go's ListDiagnosisNames/
// ListDiagnosisNamesAll) moved to internal/medicalrecord with this batch, and no other
// internal/handler file used them (verified — golangci-lint's unused check would otherwise
// flag a facade with zero callers). Not-yet-migrated handler files that need optional-uint64
// query parsing can call httpapi.ParseOptionalUint64Query directly.

func parsePagination(c *gin.Context) (page, limit int, err error) {
	return httpapi.ParsePagination(c)
}

func parsePaginationWithMax(c *gin.Context, maxLimit int) (page, limit int, err error) {
	return httpapi.ParsePaginationWithMax(c, maxLimit)
}

func parseIDParam(c *gin.Context, key string) (uint64, bool) {
	return httpapi.ParseIDParam(c, key)
}

func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, bool) {
	return httpapi.ParseUUIDParam(c, key)
}
