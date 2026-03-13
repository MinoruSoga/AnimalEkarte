package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListReservations godoc
// @Summary 予約一覧取得
// @Description クリニックの予約一覧をページネーション付きで取得する
// @Tags Reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param date query string false "日付フィルター (YYYY-MM-DD)"
// @Param status query string false "ステータスフィルター"
// @Success 200 {object} handler.PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations [get]
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
	c.JSON(http.StatusOK, PaginatedResponse{Data: reservations, Total: total, Page: page, Limit: limit})
}

// GetReservation godoc
// @Summary 予約取得
// @Description 指定IDの予約を取得する
// @Tags Reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "予約ID"
// @Success 200 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [get]
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
// @Summary 予約作成
// @Description 新しい予約を作成する
// @Tags Reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.ReservationAppointment true "予約情報"
// @Success 201 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations [post]
func (h *Handler) CreateReservation(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input model.ReservationAppointment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ClinicID = clinicID
	if err := h.svc.Reservation.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateReservation godoc
// @Summary 予約更新
// @Description 指定IDの予約を更新する
// @Tags Reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "予約ID"
// @Param body body model.ReservationAppointment true "予約情報"
// @Success 200 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [put]
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
	var input model.ReservationAppointment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	input.ClinicID = clinicID
	if err := h.svc.Reservation.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteReservation godoc
// @Summary 予約削除
// @Description 指定IDの予約を削除する
// @Tags Reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "予約ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [delete]
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
