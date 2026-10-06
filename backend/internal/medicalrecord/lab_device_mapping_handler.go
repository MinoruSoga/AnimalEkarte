package medicalrecord

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type labMappingDraftItemResponse struct {
	DeviceItemCode    string  `json:"device_item_code"`
	DeviceUnit        string  `json:"device_unit"`
	ChartUnit         string  `json:"chart_unit"`
	UnitMismatch      bool    `json:"unit_mismatch"`
	ExamTypeFieldID   *uint64 `json:"exam_type_field_id,omitempty"`
	ProposedFieldName string  `json:"proposed_field_name"`
	Source            string  `json:"source"`
	LLMError          string  `json:"llm_error"`
	Confirmed         bool    `json:"confirmed"`
}

type labMappingDraftResponse struct {
	ID           uint64                        `json:"id"`
	Status       string                        `json:"status"`
	SourceType   string                        `json:"source_type"`
	Species      string                        `json:"species"`
	ErrorMessage string                        `json:"error_message"`
	Items        []labMappingDraftItemResponse `json:"items"`
}

type labMappingGenerationResponse struct {
	Stopped bool `json:"stopped"`
}

type confirmLabMappingItemRequest struct {
	DeviceItemCode  string `json:"device_item_code"`
	ExamTypeFieldID uint64 `json:"exam_type_field_id"`
	ValueShape      string `json:"value_shape"`
	ChartUnit       string `json:"chart_unit"`
}

type confirmLabMappingRequest struct {
	Items []confirmLabMappingItemRequest `json:"items"`
}

type putLabMappingGenerationRequest struct {
	Stopped bool `json:"stopped"`
}

// WithMappingDrafts attaches the EMR-297 draft service. Optional in legacy lab-import tests.
func (h *LabImportHandler) WithMappingDrafts(service LabMappingDraftService) *LabImportHandler {
	h.mappingDrafts = service
	return h
}

func (h *LabImportHandler) mappingDraftService() (LabMappingDraftService, bool) {
	if h == nil || h.mappingDrafts == nil {
		return nil, false
	}
	return h.mappingDrafts, true
}

// CreateLabMappingDraft godoc
// POST /api/v1/lab-device/mapping-drafts
func (h *LabImportHandler) CreateLabMappingDraft(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "create") {
		return
	}
	svc, ok := h.mappingDraftService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab mapping draft service is not configured"))
		return
	}
	var req labDeviceFramesRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, labDeviceFramesMaxRequestBytes)
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("lab mapping draft request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	payload, err := decodeLabDevicePayloadBase64(req.PayloadBase64)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	view, err := svc.Create(c.Request.Context(), clinicID, payload, req.DeviceHint)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLabMappingDraftResponse(view))
}

// GetLabMappingDraft godoc
// GET /api/v1/lab-device/mapping-drafts/:id
func (h *LabImportHandler) GetLabMappingDraft(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	svc, ok := h.mappingDraftService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab mapping draft service is not configured"))
		return
	}
	view, err := svc.Get(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabMappingDraftResponse(view))
}

// ConfirmLabMappingDraft godoc
// POST /api/v1/lab-device/mapping-drafts/:id/confirm
func (h *LabImportHandler) ConfirmLabMappingDraft(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "edit") {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	svc, ok := h.mappingDraftService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab mapping draft service is not configured"))
		return
	}
	var req confirmLabMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	items := make([]ConfirmLabMappingItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, ConfirmLabMappingItem(item))
	}
	view, err := svc.Confirm(c.Request.Context(), clinicID, id, staffID, ConfirmLabMappingInput{Items: items})
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabMappingDraftResponse(view))
}

// PutLabMappingGeneration godoc
// PUT /api/v1/lab-device/mapping-generation
func (h *LabImportHandler) PutLabMappingGeneration(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "edit") {
		return
	}
	svc, ok := h.mappingDraftService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab mapping draft service is not configured"))
		return
	}
	var req putLabMappingGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := svc.SetGenerationStopped(c.Request.Context(), clinicID, req.Stopped); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, labMappingGenerationResponse(req))
}

func toLabMappingDraftResponse(view *LabMappingDraftView) labMappingDraftResponse {
	items := make([]labMappingDraftItemResponse, 0, len(view.Items))
	for _, item := range view.Items {
		items = append(items, labMappingDraftItemResponse{
			DeviceItemCode:    item.DeviceItemCode,
			DeviceUnit:        item.DeviceUnit,
			ChartUnit:         item.ChartUnit,
			UnitMismatch:      item.UnitMismatch,
			ExamTypeFieldID:   item.ExamTypeFieldID,
			ProposedFieldName: item.ProposedFieldName,
			Source:            item.ItemSource,
			LLMError:          item.LLMError,
			Confirmed:         item.Confirmed,
		})
	}
	return labMappingDraftResponse{
		ID:           view.Draft.ID,
		Status:       view.Draft.Status,
		SourceType:   view.Draft.SourceType,
		Species:      view.Draft.Species,
		ErrorMessage: view.Draft.ErrorMessage,
		Items:        items,
	}
}
