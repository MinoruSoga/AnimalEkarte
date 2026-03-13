package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type createReservationInput struct {
	StartTime     time.Time `json:"start_time"      binding:"required"`
	EndTime       time.Time `json:"end_time"        binding:"required"`
	OwnerID       *uint64   `json:"owner_id"`
	PetID         *uint64   `json:"pet_id"`
	VisitType     string    `json:"visit_type"`
	ServiceTypeID uint64    `json:"service_type_id" binding:"required"`
	DoctorID      *uint64   `json:"doctor_id"`
	IsDesignated  bool      `json:"is_designated"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
}

type updateReservationInput struct {
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	OwnerID       *uint64    `json:"owner_id"`
	PetID         *uint64    `json:"pet_id"`
	VisitType     string     `json:"visit_type"`
	ServiceTypeID uint64     `json:"service_type_id"`
	DoctorID      *uint64    `json:"doctor_id"`
	IsDesignated  *bool      `json:"is_designated"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes"`
}

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}
	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	reservations, total, err := h.svc.Reservation.List(c.Request.Context(), clinicID, page, limit, date, status, petID, ownerID)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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
	var input createReservationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reservation := &model.ReservationAppointment{
		ClinicID:      clinicID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ServiceTypeID: input.ServiceTypeID,
		DoctorID:      input.DoctorID,
		IsDesignated:  input.IsDesignated,
		Notes:         input.Notes,
	}
	if input.VisitType != "" {
		vt, err := validateEnum(input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_type: " + err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		reservation.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.Reservation.Create(ctx, reservation); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateReservationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reservation := &model.ReservationAppointment{
		ID:            id,
		ClinicID:      clinicID,
		OwnerID:       input.OwnerID,
		PetID:         input.PetID,
		ServiceTypeID: input.ServiceTypeID,
		DoctorID:      input.DoctorID,
		Notes:         input.Notes,
	}
	if input.StartTime != nil {
		reservation.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		reservation.EndTime = *input.EndTime
	}
	if input.VisitType != "" {
		vt, err := validateEnum(input.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_type: " + err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		reservation.Status = status
	}
	if input.IsDesignated != nil {
		reservation.IsDesignated = *input.IsDesignated
	}

	ctx := c.Request.Context()
	if err := h.svc.Reservation.Update(ctx, reservation); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", reservation.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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
	reservations.POST("", h.CreateReservation)
	reservations.GET("/:id", h.GetReservation)
	reservations.PATCH("/:id", h.UpdateReservation)
	reservations.DELETE("/:id", h.DeleteReservation)
}
