package reservation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationHandler は予約 CRUD の HTTP handler。
type ReservationHandler struct {
	svc              ReservationService
	medicalRecord    medicalRecordAutoCreator
	liff             liffAvailability
	staffAssignments staffAssignmentFinder
}

// NewReservationHandler は ReservationHandler を構築する。
func NewReservationHandler(svc ReservationService, medicalRecord medicalRecordAutoCreator, liff liffAvailability, staffAssignments staffAssignmentFinder) *ReservationHandler {
	return &ReservationHandler{svc: svc, medicalRecord: medicalRecord, liff: liff, staffAssignments: staffAssignments}
}

// checkDoctorClinicAssignment は医師が指定クリニックに所属しているかを確認する共通ヘルパー
// （internal/handler/validation.go の同名メソッドの忠実な移植・appointment/liff 残留consumerは旧実装を継続使用）。
func (h *ReservationHandler) checkDoctorClinicAssignment(ctx context.Context, clinicID, doctorID uint64) error {
	if doctorID == 0 {
		return nil
	}
	assignments, err := h.staffAssignments.FindAllByStaffID(ctx, doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff assignment")
	}
	for i := range assignments {
		if assignments[i].ClinicID == clinicID {
			return nil
		}
	}
	return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
}

const reservationListMaxLimit = 1000

func toReservationAvailableTimeResponse(slot *TimeSlot) liffTimeSlotResponse {
	return liffTimeSlotResponse{StartTime: slot.StartTime, EndTime: slot.EndTime}
}

// ListReservations godoc
func (h *ReservationHandler) ListReservations(c *gin.Context) {
	// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePaginationWithMax(c, reservationListMaxLimit)
	if err != nil {
		respondError(c, err)
		return
	}

	q := newListReservationQuery(c.Request.URL.Query())
	filters, err := q.toServiceFilters()
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	reservations, total, err := h.svc.List(c.Request.Context(), clinicIDs, page, limit, filters.Date, filters.StartDate, filters.EndDate, filters.Status, filters.Source, filters.PetID, filters.OwnerID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(reservations, toReservationResponse), total, page, limit))
}

// GetReservation godoc
func (h *ReservationHandler) GetReservation(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := httpapi.ResolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	reservation, err := h.svc.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

// GetReservationAvailableTimes godoc
// GET /reservations/available-times?reservation_type_id=:id&staff_id=:id&date=YYYY-MM-DD
func (h *ReservationHandler) GetReservationAvailableTimes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	filters, err := newReservationAvailableTimesQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		respondError(c, err)
		return
	}
	if h.liff == nil {
		respondError(c, apperrors.WrapNotImplemented("予約可能時間の取得は未設定です"))
		return
	}
	// BUG-015: staff path allows inactive types via GetStaffAvailableTimes when present.
	// Falls back to GetAvailableTimes for mocks that only implement liffAvailability.
	var slots []TimeSlot
	if staffTimes, ok := h.liff.(interface {
		GetStaffAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
	}); ok {
		slots, err = staffTimes.GetStaffAvailableTimes(c.Request.Context(), clinicID, filters.ReservationTypeID, filters.StaffID, filters.Date)
	} else {
		slots, err = h.liff.GetAvailableTimes(c.Request.Context(), clinicID, filters.ReservationTypeID, filters.StaffID, filters.Date)
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(slots, toReservationAvailableTimeResponse))
}

// CreateReservation godoc
func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var input createReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput(clinicID, staffID)
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（クロスクリニック FK 防止）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			respondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Create(ctx, svcInput)
	if err != nil {
		respondError(c, err)
		return
	}
	// #83: 予約確定(confirmed)で作成された場合はカルテを best-effort で自動作成する
	if shouldAutoCreateMedicalRecordForReservation(reservation) && h.medicalRecord != nil {
		h.medicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}
	c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", reservation.ID))
	c.JSON(http.StatusCreated, toReservationResponse(reservation))
}

// CreateReservationBatch creates a single atomic shared doctor/time booking for selected pets.
func (h *ReservationHandler) CreateReservationBatch(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	var req createReservationBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input, pets, err := req.toServiceInput(clinicID, staffID)
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}
	if input.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, *input.DoctorID); err != nil {
			respondError(c, err)
			return
		}
	}
	reservations, err := h.svc.CreateBatch(c.Request.Context(), input, pets)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, httpapi.MapSlice(reservations, toReservationResponse))
}

// UpdateReservation godoc
func (h *ReservationHandler) UpdateReservation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateReservationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := input.toServiceInput()
	if err != nil {
		respondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// BUG-144: staff_id のクリニック所属チェック（Update時も）
	if svcInput.DoctorID != nil {
		if err := h.checkDoctorClinicAssignment(ctx, clinicID, *svcInput.DoctorID); err != nil {
			respondError(c, err)
			return
		}
	}

	reservation, err := h.svc.Update(ctx, clinicID, id, &svcInput)
	if err != nil {
		respondError(c, err)
		return
	}

	// 予約確定/受付済みの場合は通常カルテを best-effort で自動作成する(#83)。owner/pet 紐付け後の遅延作成(Q9)も兼ねる。
	if shouldAutoCreateMedicalRecordForReservation(reservation) && h.medicalRecord != nil {
		h.medicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
	}

	// #83 Q10: キャンセルされた予約に紐づく draft カルテを論理削除する（best-effort）
	if svcInput.Status != nil && *svcInput.Status == model.ReservationStatusCancelled && h.medicalRecord != nil {
		h.medicalRecord.DeleteDraftFromReservation(ctx, clinicID, reservation.ID)
	}

	c.JSON(http.StatusOK, toReservationResponse(reservation))
}

func shouldAutoCreateMedicalRecordForReservation(reservation *model.Reservation) bool {
	if reservation == nil {
		return false
	}
	// #83: 予約確定(confirmed)/受付済み(checked_in)の予約でカルテを作成する。
	// 更新後の reservation.Status で判定するため、owner/pet 紐付けの Update でも confirmed なら作成され、
	// owner/pet 未確定でスキップされた分が紐付け後に遅延作成される(Q9)。
	// 同日同ペットの重複は AutoCreateFromReservation 側でスキップされるため二重作成しない。
	if reservation.Status != model.ReservationStatusConfirmed && reservation.Status != model.ReservationStatusCheckedIn {
		return false
	}
	if reservation.ReservationType != nil && reservation.ReservationType.Category == model.ReservationTypeCategoryTrimming {
		return false
	}
	if reservation.ReservationType != nil &&
		(strings.Contains(reservation.ReservationType.Name, "入院") ||
			strings.Contains(reservation.ReservationType.Name, "ホテル")) {
		return false
	}
	return true
}

// DeleteReservation godoc
func (h *ReservationHandler) DeleteReservation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), clinicID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateReservationReservationRoute godoc
// PATCH /reservations/:id/reservation-route — 予約経路を更新する（FEAT-381-2）。
func (h *ReservationHandler) UpdateReservationReservationRoute(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchReservationReservationRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	reservation, err := h.svc.UpdateReservationRoute(
		c.Request.Context(), clinicID, id,
		req.toServiceInput(),
	)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toReservationResponse(reservation))
}
