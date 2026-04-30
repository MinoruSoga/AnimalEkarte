package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListVaccinations godoc
func (h *Handler) ListVaccinations(c *gin.Context) {
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

	vaccinations, total, err := h.svc.Vaccination.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(vaccinations, total, page, limit))
}

// GetVaccination godoc
func (h *Handler) GetVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	vaccination, err := h.svc.Vaccination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// CreateVaccination godoc
func (h *Handler) CreateVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input createVaccinationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// Parse Date (required)
	date, err := parseDate(input.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid date: %v", err)))
		return
	}
	if date == nil {
		RespondError(c, apperrors.WrapInvalidInput("date is required"))
		return
	}

	// Parse NextDate (optional)
	nextDate, err := parseDate(input.NextDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid next_date: %v", err)))
		return
	}

	var nextScheduleType *model.NextScheduleType
	if input.NextScheduleType != "" {
		nst := model.NextScheduleType(input.NextScheduleType)
		nextScheduleType = &nst
	}

	svcInput := &service.CreateVaccinationInput{
		MedicalRecordID:  input.MedicalRecordID,
		PetID:            input.PetID,
		VaccineID:        input.VaccineID,
		Date:             *date,
		DoctorID:         input.DoctorID,
		NextDate:         nextDate,
		NextScheduleType: nextScheduleType,
		Supplemental:     input.Supplemental,
		Lot1:             input.Lot1,
		Lot2:             input.Lot2,
		Lot3:             input.Lot3,
		Lot4:             input.Lot4,
		Remarks:          input.Remarks,
	}
	vaccination, err := h.svc.Vaccination.Create(c.Request.Context(), clinicID, svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-003: ワクチンタグ同期（best-effort、失敗しても作成レスポンスは返す）
	if vaccination.PetID != nil {
		if pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, *vaccination.PetID); err == nil {
			_ = h.svc.LstepTagSync.SyncVaccineTag(c.Request.Context(), clinicID, pet.OwnerID, vaccination.ID)
		}
	}
	c.Header("Location", fmt.Sprintf("/api/v1/vaccinations/%d", vaccination.ID))
	c.JSON(http.StatusCreated, toVaccinationResponse(vaccination))
}

// UpdateVaccination godoc
func (h *Handler) UpdateVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input updateVaccinationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// Parse Date (optional)
	date, err := parseDate(input.Date)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid date: %v", err)))
		return
	}

	// Parse NextDate (optional)
	nextDate, err := parseDate(input.NextDate)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid next_date: %v", err)))
		return
	}

	svcInput := service.UpdateVaccinationInput{
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		VaccineID:       input.VaccineID,
		Date:            date,
		DoctorID:        input.DoctorID,
		NextDate:        nextDate,
		Supplemental:    input.Supplemental,
		Lot1:            input.Lot1,
		Lot2:            input.Lot2,
		Lot3:            input.Lot3,
		Lot4:            input.Lot4,
		Remarks:         input.Remarks,
	}
	if input.NextScheduleType != nil {
		nst := model.NextScheduleType(*input.NextScheduleType)
		svcInput.NextScheduleType = &nst
	}

	vaccination, err := h.svc.Vaccination.Update(c.Request.Context(), clinicID, id, &svcInput)
	if err != nil {
		RespondError(c, err)
		return
	}
	// ISSUE-004: ワクチン更新後に飼い主全体のタグを DB 状態から再構築（best-effort）。
	// 種別ごとの最新接種日のみがタグとして残り、編集前の古い日付タグは解除される。
	if vaccination.PetID != nil {
		if pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, *vaccination.PetID); err == nil {
			_ = h.svc.LstepTagSync.ResyncOwnerVaccineTags(c.Request.Context(), clinicID, pet.OwnerID)
		}
	}
	c.JSON(http.StatusOK, toVaccinationResponse(vaccination))
}

// DeleteVaccination godoc
func (h *Handler) DeleteVaccination(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	// ISSUE-004: Delete 前に vaccination を取得してカテゴリタグ再計算に必要な owner_id を確保。
	// ソフトデリート後は FindByID が NotFound を返すため、owner 解決は削除前に行う必要がある。
	vaccination, err := h.svc.Vaccination.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	var ownerID uint64
	if vaccination.PetID != nil {
		if pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, *vaccination.PetID); err == nil {
			ownerID = pet.OwnerID
		}
	}
	if err := h.svc.Vaccination.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	// ISSUE-004: 削除後の DB 状態から飼い主全体のタグを再構築（best-effort）。
	// 同じ種別の他レコードがあればその最新日タグを残す。残らなければ全 vaccine_* を解除する。
	if ownerID != 0 {
		_ = h.svc.LstepTagSync.ResyncOwnerVaccineTags(c.Request.Context(), clinicID, ownerID)
	}
	c.Status(http.StatusNoContent)
}
