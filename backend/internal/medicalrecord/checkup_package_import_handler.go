package medicalrecord

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

const maxCheckupPackageImportBodyBytes = 1 << 20 // 1 MiB

// CheckupPackageImportHandler serves dry-run and apply for versioned checkup packages.
type CheckupPackageImportHandler struct {
	service CheckupPackageImportService
}

// NewCheckupPackageImportHandler initializes a CheckupPackageImportHandler.
func NewCheckupPackageImportHandler(service CheckupPackageImportService) *CheckupPackageImportHandler {
	return &CheckupPackageImportHandler{service: service}
}

type checkupPackageImportRequest struct {
	Manifest json.RawMessage `json:"manifest" binding:"required"`
}

// PreviewCheckupPackageImport godoc
// POST /api/v1/checkup-package-imports/preview — dry-run (domain write zero).
func (h *CheckupPackageImportHandler) PreviewCheckupPackageImport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	raw, err := readCheckupPackageImportBody(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if h.service == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("checkup package import service is not configured"))
		return
	}
	receipt, err := h.service.Preview(c.Request.Context(), clinicID, actorID, raw)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, receipt)
}

// ApplyCheckupPackageImport godoc
// POST /api/v1/checkup-package-imports — apply with explicit action.
func (h *CheckupPackageImportHandler) ApplyCheckupPackageImport(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	raw, err := readCheckupPackageImportBody(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if h.service == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("checkup package import service is not configured"))
		return
	}
	receipt, err := h.service.Apply(c.Request.Context(), clinicID, actorID, raw)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, receipt)
}

func readCheckupPackageImportBody(c *gin.Context) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCheckupPackageImportBodyBytes)
	var req checkupPackageImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	if len(req.Manifest) == 0 || string(req.Manifest) == "null" {
		return nil, apperrors.WrapInvalidInput("manifest is required")
	}
	return []byte(req.Manifest), nil
}
