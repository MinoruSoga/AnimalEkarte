package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllReservations godoc
// @Summary 予約一覧取得
// @Description 登録されている予約の一覧を取得します。date パラメータが指定された場合は当日分のみ返します。
// @Tags reservations
// @Accept json
// @Produce json
// @Param date query string false "日付フィルタ（YYYY-MM-DD）。指定した場合は当日の予約のみ返す"
// @Param status query string false "ステータスでフィルタリング"
// @Success 200 {array} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllReservations(c *gin.Context) {
	ctx := c.Request.Context()

	date := c.Query("date")
	if date != "" {
		reservations, err := h.svc.GetReservationsByDate(ctx, date)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get reservations by date",
				slog.String("date", date),
				slog.String("error", err.Error()))
			h.handleError(c, err, "reservation", date)
			return
		}
		c.JSON(http.StatusOK, reservations)
		return
	}

	reservations, err := h.svc.GetAllReservations(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservations", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, reservations)
}

// GetReservationByID godoc
// @Summary 予約詳細取得
// @Description 指定されたIDの予約情報を取得します
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path string true "予約ID (UUID)"
// @Success 200 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [get]
func (h *Handler) GetReservationByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	reservation, err := h.svc.GetReservationByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "reservation", id)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// GetReservationsByPetID godoc
// @Summary ペットの予約一覧取得
// @Description 指定されたペットIDの予約一覧を取得します
// @Tags reservations
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/reservations [get]
func (h *Handler) GetReservationsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("id")

	reservations, err := h.svc.GetReservationsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "reservation", petID)
		return
	}

	c.JSON(http.StatusOK, reservations)
}

// GetReservationsByOwnerID godoc
// @Summary 飼い主の予約一覧取得
// @Description 指定された飼い主IDの予約一覧を取得します
// @Tags reservations
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/reservations [get]
func (h *Handler) GetReservationsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("id")

	reservations, err := h.svc.GetReservationsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "reservation", ownerID)
		return
	}

	c.JSON(http.StatusOK, reservations)
}

// CreateReservation godoc
// @Summary 予約作成
// @Description 新しい予約を作成します
// @Tags reservations
// @Accept json
// @Produce json
// @Param reservation body model.CreateReservationRequest true "予約情報"
// @Success 201 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations [post]
func (h *Handler) CreateReservation(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	reservation, err := h.svc.CreateReservation(ctx, &req)
	if err != nil {
		h.handleError(c, err, "reservation", "")
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// UpdateReservation godoc
// @Summary 予約更新
// @Description 既存の予約を更新します
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path string true "予約ID (UUID)"
// @Param reservation body model.UpdateReservationRequest true "予約情報"
// @Success 200 {object} model.ReservationAppointment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [put]
func (h *Handler) UpdateReservation(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	reservation, err := h.svc.UpdateReservation(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "reservation", id)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// DeleteReservation godoc
// @Summary 予約削除
// @Description 指定された予約を削除します
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path string true "予約ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reservations/{id} [delete]
func (h *Handler) DeleteReservation(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteReservation(ctx, id); err != nil {
		h.handleError(c, err, "reservation", id)
		return
	}

	c.Status(http.StatusNoContent)
}
