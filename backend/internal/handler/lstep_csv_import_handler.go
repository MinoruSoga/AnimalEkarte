package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type lstepCsvImportResponse struct {
	ID               string          `json:"id"`
	ClinicID         string          `json:"clinic_id"`
	CsvType          string          `json:"csv_type"`
	FileName         string          `json:"file_name"`
	UploadedByUserID *string         `json:"uploaded_by_user_id,omitempty"`
	RowCount         int             `json:"row_count"`
	SuccessCount     int             `json:"success_count"`
	ErrorCount       int             `json:"error_count"`
	Status           string          `json:"status"`
	ErrorLog         json.RawMessage `json:"error_log,omitempty"`
	ImportedAt       *string         `json:"imported_at,omitempty"`
	CreatedAt        string          `json:"created_at"`
}

func toLstepCsvImportResponse(m *model.LstepCsvImport) lstepCsvImportResponse {
	r := lstepCsvImportResponse{
		ID:           m.ID.String(),
		ClinicID:     strconv.FormatUint(m.ClinicID, 10),
		CsvType:      m.CsvType,
		FileName:     m.FileName,
		RowCount:     m.RowCount,
		SuccessCount: m.SuccessCount,
		ErrorCount:   m.ErrorCount,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
	}
	if m.UploadedByUserID != nil {
		s := strconv.FormatUint(*m.UploadedByUserID, 10)
		r.UploadedByUserID = &s
	}
	if len(m.ErrorLog) > 0 {
		r.ErrorLog = json.RawMessage(m.ErrorLog)
	}
	if m.ImportedAt != nil {
		s := m.ImportedAt.Format(time.RFC3339)
		r.ImportedAt = &s
	}
	return r
}

func toLstepCsvImportListResponse(imports []*model.LstepCsvImport) []lstepCsvImportResponse {
	resp := make([]lstepCsvImportResponse, len(imports))
	for i, imp := range imports {
		resp[i] = toLstepCsvImportResponse(imp)
	}
	return resp
}

// PostLstepCsvImportFriendAttributes godoc
// POST /api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes — 友だち属性 CSV インポート（FEAT-385）。
func (h *Handler) PostLstepCsvImportFriendAttributes(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // deferred Close on multipart file; error is non-actionable

	var actorID *uint64
	if staffID, ok2 := extractStaffID(c); ok2 {
		actorID = &staffID
	}

	imp, svcErr := h.svc.LstepCsvImport.ImportFriendAttributesCSV(c.Request.Context(), clinicID, header.Filename, file, actorID)
	if svcErr != nil {
		RespondError(c, svcErr)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/clinics/%d/lstep/csv-imports/%s", clinicID, imp.ID.String()))
	c.JSON(http.StatusCreated, toLstepCsvImportResponse(imp))
}

// ListLstepCsvImports godoc
// GET /api/v1/clinics/:clinic_id/lstep/csv-imports — インポート履歴一覧（FEAT-385）。
func (h *Handler) ListLstepCsvImports(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, parseErr := strconv.Atoi(v); parseErr == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	imports, err := h.svc.LstepCsvImport.ListByClinic(c.Request.Context(), clinicID, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepCsvImportListResponse(imports))
}

// RegisterLstepCsvImportRoutes は FEAT-385 CSV インポートのルートを登録する。
func (h *Handler) RegisterLstepCsvImportRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/clinics/:clinic_id/lstep/csv-imports")
	g.POST("/friend-attributes", h.RequirePermission(string(model.ResourceLstepCsvImport), "edit"), h.PostLstepCsvImportFriendAttributes)
	g.GET("", h.RequirePermission(string(model.ResourceLstepCsvImport), "view"), h.ListLstepCsvImports)
}
