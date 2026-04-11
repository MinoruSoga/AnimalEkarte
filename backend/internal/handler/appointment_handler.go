package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListReservations godoc
func (h *Handler) ListReservations(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var date *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid date format, use YYYY-MM-DD"))
			return
		}
		date = &t
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var petID *uint64
	if s := c.Query("pet_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
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

	var source *string
	if s := c.Query("source"); s != "" {
		source = &s
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicID, page, limit, date, status, source, petID, ownerID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(reservations, total, page, limit))
}

// GetReservation godoc
func (h *Handler) GetReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	reservation, err := h.svc.Reservation.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, reservation)
}

// CreateReservation godoc
func (h *Handler) CreateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	source := model.ReservationSourceManual
	if input.Source == string(model.ReservationSourceLine) {
		source = model.ReservationSourceLine
	}
	reservation := &model.Appointment{
		ClinicID:      clinicID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ReservationTypeID:     input.ReservationTypeID,
		DoctorID:      input.DoctorID,
		IsDesignated:  input.IsDesignated,
		Notes:         input.Notes,
		Source:        source,
	}
	if input.VisitType != "" {
		vt, err := validateEnum(input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid visit_type: "+err.Error()))
			return
		}
		reservation.VisitType = vt
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		reservation.Status = status
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if reservation.DoctorID != nil {
		assignments, asgErr := h.svc.StaffClinicAssignment.FindByStaffID(ctx, *reservation.DoctorID)
		if asgErr != nil {
			RespondError(c, apperrors.Wrap(asgErr, "failed to verify staff assignment"))
			return
		}
		assigned := false
		for _, a := range assignments {
			if a.ClinicID == clinicID {
				assigned = true
				break
			}
		}
		if !assigned {
			RespondError(c, apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません"))
			return
		}
	}

	if err := h.svc.Reservation.Create(ctx, reservation); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reservation)
}

// UpdateReservation godoc
func (h *Handler) UpdateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	svcInput := service.UpdateReservationInput{
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:      input.DoctorID,
		IsDesignated:  input.IsDesignated,
		Notes:         input.Notes,
	}
	if input.VisitType != nil {
		vt, err := validateEnum(*input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid visit_type: "+err.Error()))
			return
		}
		svcInput.VisitType = &vt
	}
	if input.Status != nil {
		status, err := validateEnum(*input.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		svcInput.Status = &status
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（Update時も）
	if svcInput.DoctorID != nil {
		assignments, asgErr := h.svc.StaffClinicAssignment.FindByStaffID(ctx, *svcInput.DoctorID)
		if asgErr != nil {
			RespondError(c, apperrors.Wrap(asgErr, "failed to verify staff assignment"))
			return
		}
		assigned := false
		for _, a := range assignments {
			if a.ClinicID == clinicID {
				assigned = true
				break
			}
		}
		if !assigned {
			RespondError(c, apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません"))
			return
		}
	}

	reservation, err := h.svc.Reservation.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 受付済みに変更された場合はカルテを best-effort で自動作成する（BE-reception-auto-create-medical-record）
	if svcInput.Status != nil && *svcInput.Status == model.ReservationStatusCheckedIn {
		h.svc.MedicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}

	c.JSON(http.StatusOK, reservation)
}

// DeleteReservation godoc
func (h *Handler) DeleteReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Reservation.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterReservationRoutes は予約関連のルートを登録する
func (h *Handler) RegisterReservationRoutes(rg *gin.RouterGroup) {
	reservations := rg.Group("/reservations")
	reservations.GET("", h.ListReservations)
	reservations.GET("/:id", h.GetReservation)
	reservations.POST("", h.RequirePermission(string(model.ResourceReservations), "create"), h.CreateReservation)
	reservations.PATCH("/:id", h.RequirePermission(string(model.ResourceReservations), "edit"), h.UpdateReservation)
	reservations.DELETE("/:id", h.RequirePermission(string(model.ResourceReservations), "delete"), h.DeleteReservation)
}
