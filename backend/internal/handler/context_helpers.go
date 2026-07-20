package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DEPRECATED facade (BE9-2B): the real implementations moved to internal/httpapi
// (BE9-2A target:httpapi, ADR-006) so domain packages like internal/manualarticle can use
// them without importing internal/handler. These unexported wrappers exist only because
// ~269 files in this package still call them unqualified; delete this file once BE9-2F
// migrates every remaining internal/handler file to internal/httpapi directly.
//
// extractContextUint64, extractClinicIDs, and parseClinicIDsParam were removed (BE9-2C):
// their only remaining callers were diagnosis_handler.go/exam_type_handler.go/
// chief_complaint_handler.go, which moved to internal/medicalrecord in this batch — zero
// callers left in internal/handler (verified: golangci-lint's unused check + grep). Not-yet-
// migrated handler files needing them can call httpapi.ExtractContextUint64/
// ExtractClinicIDs/ParseClinicIDsParam directly.

func extractStaffID(c *gin.Context) (uint64, bool) {
	return httpapi.ExtractStaffID(c)
}

func optionalStaffID(c *gin.Context) *uint64 {
	return httpapi.OptionalStaffID(c)
}

func extractClinicID(c *gin.Context) (uint64, bool) {
	return httpapi.ExtractClinicID(c)
}

func authorizeClinicIDs(c *gin.Context, requested []uint64) bool {
	return httpapi.AuthorizeClinicIDs(c, requested)
}

func resolveListClinicIDs(c *gin.Context) ([]uint64, bool) {
	return httpapi.ResolveListClinicIDs(c)
}

func resolveAllClinicIDs(c *gin.Context) ([]uint64, bool) {
	return httpapi.ResolveAllClinicIDs(c)
}

func extractIsSystemAdmin(c *gin.Context) (isSystemAdmin, ok bool) {
	return httpapi.ExtractIsSystemAdmin(c)
}
