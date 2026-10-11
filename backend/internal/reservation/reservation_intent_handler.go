package reservation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// setNextVisitIntentRequest は PUT /reservations/pets/:petId/next-visit-intent の body。
// intended_date は暦日（YYYY-MM-DD）で、JST の日付として解釈する。
type setNextVisitIntentRequest struct {
	IntendedDate          string  `json:"intended_date" binding:"required"`
	SourceMedicalRecordID *uint64 `json:"source_medical_record_id"`
}

func (r *setNextVisitIntentRequest) toServiceInput(staffID *uint64) (SetNextVisitIntentInput, error) {
	intendedDate, err := ParseJSTDate(r.IntendedDate)
	if err != nil {
		return SetNextVisitIntentInput{}, err
	}
	return SetNextVisitIntentInput{
		IntendedDate:          intendedDate,
		SourceMedicalRecordID: r.SourceMedicalRecordID,
		CreatedBy:             staffID,
	}, nil
}

// confirmTentativeRequest は POST /reservations/:id/confirm-tentative の body。
// tentative には時刻・担当医が未確定のため start_time/end_time は必須。
type confirmTentativeRequest struct {
	StartTime         time.Time `json:"start_time" binding:"required"`
	EndTime           time.Time `json:"end_time" binding:"required"`
	ReservationTypeID *uint64   `json:"reservation_type_id"`
	DoctorID          *uint64   `json:"doctor_id"`
	IsDesignated      *bool     `json:"is_designated"`
	Status            *string   `json:"status" binding:"omitempty,oneof=pending confirmed"`
	Notes             *string   `json:"notes"`
}

func (r *confirmTentativeRequest) toServiceInput() *ConfirmTentativeInput {
	input := &ConfirmTentativeInput{
		StartTime:         r.StartTime,
		EndTime:           r.EndTime,
		ReservationTypeID: r.ReservationTypeID,
		DoctorID:          r.DoctorID,
		IsDesignated:      r.IsDesignated,
		Notes:             r.Notes,
	}
	if r.Status != nil {
		status := model.ReservationStatus(*r.Status)
		input.Status = &status
	}
	return input
}

// SetNextVisitIntent godoc
// PUT /reservations/pets/:petId/next-visit-intent — ペットの次回予定を登録/更新する（EMR-308）。
func (h *CRUDHandler) SetNextVisitIntent(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "petId")
	if !ok {
		return
	}
	var req setNextVisitIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input, err := req.toServiceInput(&staffID)
	if err != nil {
		respondError(c, err)
		return
	}
	// EMR-308: PUT は upsert — 既存 intent の上書きは更新操作のため reservations:edit を
	// repo のロック内で要求する。ルート middleware の create チェックとは別に、ここで
	// edit 権限を解決して input へ載せる（checker 不在時は fail-closed で false）。
	if check, ok := httpapi.PeekClinicPermissionChecker(c); ok {
		input.CanUpdateExisting = check(c, clinicID, string(model.ResourceReservations), "edit")
	}
	reservation, err := h.svc.SetNextVisitIntent(c.Request.Context(), clinicID, petID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// ClearNextVisitIntent godoc
// DELETE /reservations/pets/:petId/next-visit-intent — ペットの次回予定を除去する（EMR-308）。
func (h *CRUDHandler) ClearNextVisitIntent(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "petId")
	if !ok {
		return
	}
	expected, err := parseClearNextVisitIntentExpectation(c)
	if err != nil {
		respondError(c, err)
		return
	}
	if expected != nil {
		guarded, ok := h.svc.(guardedNextVisitIntentClearer)
		if !ok {
			respondError(c, apperrors.WrapInternalServerError("guarded intent clear dependency is required"))
			return
		}
		err = guarded.ClearNextVisitIntentIfMatches(c.Request.Context(), clinicID, petID, *expected)
	} else {
		err = h.svc.ClearNextVisitIntent(c.Request.Context(), clinicID, petID)
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ConfirmTentativeAppointment godoc
// POST /reservations/:id/confirm-tentative — tentative 予約を本予約へ確定する（EMR-308）。
func (h *CRUDHandler) ConfirmTentativeAppointment(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req confirmTentativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input := req.toServiceInput()
	ctx := c.Request.Context()
	reservation, err := h.svc.ConfirmTentative(ctx, clinicID, id, input)
	if err != nil {
		respondError(c, err)
		return
	}
	// 確定後は通常予約と同じ挙動: confirmed 時のカルテ自動作成（best-effort）を共有する。
	if shouldAutoCreateMedicalRecordForReservation(reservation) && h.medicalRecord != nil {
		h.autoCreateMedicalRecordAfterConfirm(ctx, clinicID, reservation)
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}
