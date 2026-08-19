package medicalrecord

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// WithDeviceMasters attaches the BRT-96 master service. Optional in legacy lab-import tests.
func (h *LabImportHandler) WithDeviceMasters(service LabDeviceItemMasterService) *LabImportHandler {
	h.deviceMasters = service
	return h
}

// WithDeviceReceive attaches the BRT-97 wait/frames/board service. Optional in legacy tests.
func (h *LabImportHandler) WithDeviceReceive(service LabDeviceReceiveService) *LabImportHandler {
	h.deviceReceive = service
	return h
}

func (h *LabImportHandler) deviceReceiveService() (LabDeviceReceiveService, bool) {
	if h == nil || h.deviceReceive == nil {
		return nil, false
	}
	return h.deviceReceive, true
}

func (h *LabImportHandler) deviceMasterService() (LabDeviceItemMasterService, bool) {
	if h == nil || h.deviceMasters == nil {
		return nil, false
	}
	return h.deviceMasters, true
}

// ListLabDeviceItemMasters godoc
// GET /api/v1/lab-device-item-masters
func (h *LabImportHandler) ListLabDeviceItemMasters(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	sourceType := c.Query("source_type")
	items, err := svc.List(c.Request.Context(), clinicID, sourceType)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(items, toLabDeviceItemMasterResponse))
}

// EnsureLabDeviceItemMasters godoc
// POST /api/v1/lab-device-item-masters/ensure
func (h *LabImportHandler) EnsureLabDeviceItemMasters(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	inserted, items, err := svc.EnsureDefaults(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, labDeviceItemMasterEnsureResponse{
		InsertedCount: inserted,
		Items:         httpapi.MapSlice(items, toLabDeviceItemMasterResponse),
	})
}

// UpdateLabDeviceItemMaster godoc
// PATCH /api/v1/lab-device-item-masters/:id
func (h *LabImportHandler) UpdateLabDeviceItemMaster(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	var req updateLabDeviceItemMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	item, err := svc.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceItemMasterResponse(item))
}

// ListLabDevices godoc
// GET /api/v1/lab-devices
func (h *LabImportHandler) ListLabDevices(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	devices, err := svc.ListDevices(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(devices, toLabDeviceResponse))
}

// CreateLabDevice godoc
// POST /api/v1/lab-devices
func (h *LabImportHandler) CreateLabDevice(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	var req createLabDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	device, err := svc.CreateDevice(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLabDeviceResponse(device))
}

// UpdateLabDevice godoc
// PATCH /api/v1/lab-devices/:id
func (h *LabImportHandler) UpdateLabDevice(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	svc, ok := h.deviceMasterService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device item master service is not configured"))
		return
	}
	var req updateLabDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	device, err := svc.UpdateDevice(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceResponse(device))
}

// ReceiveLabDeviceFrames godoc
// POST /api/v1/lab-device/frames
func (h *LabImportHandler) ReceiveLabDeviceFrames(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	var req labDeviceFramesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	payload, err := decodeLabDevicePayloadBase64(req.PayloadBase64)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	result, err := svc.ReceiveFrames(c.Request.Context(), clinicID, payload, req.DeviceHint)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceReceiveResponse(result))
}

// PutLabDeviceWait godoc
// PUT /api/v1/lab-device/wait
func (h *LabImportHandler) PutLabDeviceWait(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	var req labDeviceWaitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	wait, err := svc.PutWait(c.Request.Context(), clinicID, staffID, req.PetID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceWaitResponse(wait))
}

// DeleteLabDeviceWait godoc
// DELETE /api/v1/lab-device/wait
func (h *LabImportHandler) DeleteLabDeviceWait(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	if err := svc.ClearWait(c.Request.Context(), clinicID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetLabDeviceBoard godoc
// GET /api/v1/lab-device/board
func (h *LabImportHandler) GetLabDeviceBoard(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	board, err := svc.Board(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceBoardResponse(board))
}

// GetLabDeviceUnlinked godoc
// GET /api/v1/lab-device/unlinked
func (h *LabImportHandler) GetLabDeviceUnlinked(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	cards, err := svc.Unlinked(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(cards, toLabDeviceJobCardResponse))
}

// GetLabDeviceStation godoc
// GET /api/v1/lab-device/station
func (h *LabImportHandler) GetLabDeviceStation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	station, err := svc.GetStation(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceStationResponse(station))
}

// PutLabDeviceStation godoc
// PUT /api/v1/lab-device/station
func (h *LabImportHandler) PutLabDeviceStation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	var req labDeviceStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	station, err := svc.PutStation(c.Request.Context(), clinicID, req.SlotsJSON)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceStationResponse(station))
}

// AttachLabDeviceJob godoc
// POST /api/v1/lab-imports/:job_id/attach
func (h *LabImportHandler) AttachLabDeviceJob(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	var req labDeviceAttachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	card, err := svc.Attach(c.Request.Context(), clinicID, jobID, req.PetID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceJobCardResponse(card))
}

// DetachLabDeviceJob godoc
// POST /api/v1/lab-imports/:job_id/detach
func (h *LabImportHandler) DetachLabDeviceJob(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := httpapi.ParseUUIDParam(c, "job_id")
	if !ok {
		return
	}
	svc, ok := h.deviceReceiveService()
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("lab device receive service is not configured"))
		return
	}
	card, err := svc.Detach(c.Request.Context(), clinicID, jobID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabDeviceJobCardResponse(card))
}
