package lstep

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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
		CreatedAt:    httpapi.LocalTimeRFC3339(m.CreatedAt),
	}
	uploadedBy := strconv.FormatUint(m.UploadedByUserID, 10)
	r.UploadedByUserID = &uploadedBy
	if len(m.ErrorLog) > 0 {
		r.ErrorLog = json.RawMessage(m.ErrorLog)
	}
	if m.ImportedAt != nil {
		s := httpapi.LocalTimeRFC3339(*m.ImportedAt)
		r.ImportedAt = &s
	}
	return r
}

func toLstepCsvImportListResponse(imports []*model.LstepCsvImport) []lstepCsvImportResponse {
	resp := make([]lstepCsvImportResponse, len(imports))
	for i, imp := range imports {
		resp[i] = toLstepCsvImportResponse(imp)
		resp[i].ErrorLog = nil
	}
	return resp
}

// ImportLstepFriendAttributesCsv godoc
// POST /api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes — 友だち属性 CSV インポート（FEAT-385）。
func (h *Handler) ImportLstepFriendAttributesCsv(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if c.Request.ContentLength > maxCSVUploadRequestBytes {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("CSV upload request exceeds size limit"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCSVUploadRequestBytes)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // deferred Close on multipart file; error is non-actionable

	staffID, ok2 := httpapi.ExtractStaffID(c)
	if !ok2 {
		return
	}

	imp, svcErr := h.csvImport.ImportFriendAttributesCSV(c.Request.Context(), clinicID, header.Filename, file, staffID)
	if svcErr != nil {
		httpapi.RespondError(c, svcErr)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/clinics/%d/lstep/csv-imports/%s", clinicID, imp.ID.String()))
	c.JSON(http.StatusCreated, toLstepCsvImportResponse(imp))
}

// ListLstepCsvImports godoc
// GET /api/v1/clinics/:clinic_id/lstep/csv-imports — インポート履歴一覧（FEAT-385）。
func (h *Handler) ListLstepCsvImports(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	limit := newListLstepCsvImportsQuery(c.Request.URL.Query()).toLimit()

	imports, err := h.csvImport.ListByClinic(c.Request.Context(), clinicID, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLstepCsvImportListResponse(imports))
}

// RegisterLstepCsvImportRoutes は FEAT-385 CSV インポートのルートを登録する。
func (h *Handler) RegisterLstepCsvImportRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/clinics/:clinic_id/lstep/csv-imports")
	g.POST(
		"/friend-attributes",
		h.requirePermission(string(model.ResourceLstepCsvImport), "edit"),
		newLstepCSVImportConcurrencyGate(maxConcurrentLstepCSVImports),
		h.ImportLstepFriendAttributesCsv,
	)
	g.GET("", h.requirePermission(string(model.ResourceLstepCsvImport), "view"), h.ListLstepCsvImports)
}
