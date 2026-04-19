package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListTrimmings はトリミング予約一覧を返す（BE-119: appointments ベース）
func (h *Handler) ListTrimmings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uint64
	if petIDStr := c.Query("pet_id"); petIDStr != "" {
		id, err := strconv.ParseUint(petIDStr, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
			return
		}
		petID = &id
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	startDate, err := parseDateQuery(c, "start_date")
	if err != nil {
		RespondError(c, err)
		return
	}
	endDate, err := parseDateQuery(c, "end_date")
	if err != nil {
		RespondError(c, err)
		return
	}

	appts, total, err := h.svc.Trimming.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	responses := make([]trimmingResponse, 0, len(appts))
	for i := range appts {
		responses = append(responses, toTrimmingResponse(&appts[i]))
	}
	c.JSON(http.StatusOK, newPaginatedResponse(responses, total, page, limit))
}

// GetTrimming はトリミング予約詳細を返す
func (h *Handler) GetTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	appt, err := h.svc.Trimming.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingResponse(appt))
}

// CreateTrimming はトリミング予約を作成する
func (h *Handler) CreateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req createTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if req.ReservationTypeID == 0 {
		RespondError(c, apperrors.WrapInvalidInput("reservation_type_id は必須です"))
		return
	}

	input := &service.CreateTrimmingInput{
		ReservationTypeID: req.ReservationTypeID,
		PetID:             req.PetID,
		StaffID:           req.StaffID,
		CourseID:          req.CourseID,
		StyleRequest:      req.StyleRequest,
		BodyWeight:        req.BW,
		BodyTemperature:   req.BT,
		UsedShampoo:       req.UsedShampoo,
		UsedRibbon:        req.UsedRibbon,
		Remarks:           req.Remarks,
		StyleImage:        req.StyleImage,
		CompletedImage:    req.CompletedImage,
		OptionIDs:         req.OptionIDs,
	}
	if req.StartTime != nil {
		input.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		input.EndTime = *req.EndTime
	}
	if req.Status != "" {
		input.Status = model.ReservationStatus(req.Status)
	}
	if req.BWUnit != "" {
		input.BWUnit = model.BodyWeightUnit(req.BWUnit)
	}

	appt, err := h.svc.Trimming.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTrimmingResponse(appt))
}

// UpdateTrimming はトリミング予約を部分更新する
func (h *Handler) UpdateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateTrimmingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateTrimmingInput{
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		PetID:           req.PetID,
		StaffID:         req.StaffID,
		CourseID:        req.CourseID,
		StyleRequest:    req.StyleRequest,
		BodyWeight:      req.BW,
		BodyTemperature: req.BT,
		UsedShampoo:     req.UsedShampoo,
		UsedRibbon:      req.UsedRibbon,
		Remarks:         req.Remarks,
		StyleImage:      req.StyleImage,
		CompletedImage:  req.CompletedImage,
		OptionIDs:       req.OptionIDs,
	}
	if req.Status != nil {
		status := model.ReservationStatus(*req.Status)
		input.Status = &status
	}
	if req.BWUnit != nil {
		unit := model.BodyWeightUnit(*req.BWUnit)
		input.BWUnit = &unit
	}

	appt, err := h.svc.Trimming.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTrimmingResponse(appt))
}

// DeleteTrimming はトリミング予約を削除する（論理削除）
func (h *Handler) DeleteTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Trimming.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
