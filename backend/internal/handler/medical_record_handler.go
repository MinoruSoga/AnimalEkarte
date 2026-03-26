package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListMedicalRecords godoc
func (h *Handler) ListMedicalRecords(c *gin.Context) {
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

	records, total, err := h.svc.MedicalRecord.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newPaginatedResponse(toMedicalRecordResponseList(records), total, page, limit))
}

// GetMedicalRecord godoc
func (h *Handler) GetMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	record, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// buildMedicalRecord はリクエストから MedicalRecord モデルを組み立てる純粋関数。
// 日付解決・必須フィールド検証・string→uint64 変換・ステータス検証を行う。
func buildMedicalRecord(clinicID uint64, input *createMedicalRecordRequest) (*model.MedicalRecord, error) {
	// 1. 日付の解決: date (time.Time) または visit_date (string "YYYY-MM-DD")
	var recordDate time.Time
	switch {
	case input.Date != nil:
		recordDate = *input.Date
	case input.VisitDate != nil && *input.VisitDate != "":
		parsed, err := time.Parse("2006-01-02", *input.VisitDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid visit_date format (expected YYYY-MM-DD)")
		}
		recordDate = parsed
	default:
		recordDate = time.Now()
	}

	// 2. 必須フィールド検証
	if input.OwnerID == nil || *input.OwnerID == "" {
		return nil, apperrors.WrapInvalidInput("owner_id is required")
	}
	if input.PetID == nil || *input.PetID == "" {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}

	// 3. ID型の変換: string → uint64
	parseOptionalID := func(s *string, field string) (*uint64, error) {
		if s == nil || *s == "" {
			return nil, nil
		}
		id, err := strconv.ParseUint(*s, 10, 64)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid " + field + " (must be numeric)")
		}
		return &id, nil
	}

	ownerID, err := parseOptionalID(input.OwnerID, "owner_id")
	if err != nil {
		return nil, err
	}
	petID, err := parseOptionalID(input.PetID, "pet_id")
	if err != nil {
		return nil, err
	}
	doctorID, err := parseOptionalID(input.DoctorID, "doctor_id")
	if err != nil {
		return nil, err
	}
	reservationAppointmentID, err := parseOptionalID(input.ReservationAppointmentID, "reservation_appointment_id")
	if err != nil {
		return nil, err
	}

	// 4. モデル組み立て（RecordNo は service 層で自動生成）
	record := &model.MedicalRecord{
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		Date:                     recordDate,
		OwnerID:                  ownerID,
		PetID:                    petID,
		DoctorID:                 doctorID,
		ReservationAppointmentID: reservationAppointmentID,
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid status: " + err.Error())
		}
		record.Status = status
	}
	return record, nil
}

// createClinicalPlanIfNeeded は plan/assessment/diagnosis フィールドが存在する場合に
// ClinicalPlan を best-effort で作成・更新する。失敗してもカルテ作成は完了済みのためエラーログのみ。
func (h *Handler) createClinicalPlanIfNeeded(ctx context.Context, recordID uint64, input *createMedicalRecordRequest) {
	if input.Plan == nil && input.Assessment == nil && input.Diagnosis1CategoryID == nil && input.Diagnosis1NameID == nil {
		return
	}
	clinicalPlanInput := &service.UpdateClinicalPlanInput{
		TreatmentPolicy:  input.Plan,
		DiagnosisDetails: input.Assessment,
	}
	if input.Diagnosis1CategoryID != nil {
		clinicalPlanInput.DiagnosisCategoryID = input.Diagnosis1CategoryID
	}
	if input.Diagnosis1NameID != nil {
		clinicalPlanInput.DiagnosisNameID = input.Diagnosis1NameID
	}
	if input.Diagnosis2CategoryID != nil {
		clinicalPlanInput.Diagnosis2CategoryID = input.Diagnosis2CategoryID
	}
	if input.Diagnosis2NameID != nil {
		clinicalPlanInput.Diagnosis2NameID = input.Diagnosis2NameID
	}
	// best-effort: failure is intentionally ignored; service layer logs the error
	_, _ = h.svc.ClinicalPlan.Update(ctx, recordID, clinicalPlanInput)
}

// CreateMedicalRecord godoc
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	record, err := buildMedicalRecord(clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	ctx := c.Request.Context()
	if err := h.svc.MedicalRecord.Create(ctx, record); err != nil {
		RespondError(c, err)
		return
	}
	h.createClinicalPlanIfNeeded(ctx, record.ID, &input)
	c.JSON(http.StatusCreated, record)
}

// UpdateMedicalRecord godoc
func (h *Handler) UpdateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var input updateMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	record := &model.MedicalRecord{
		ID:                       id,
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		ReservationAppointmentID: input.ReservationAppointmentID,
	}
	if input.Date != nil {
		record.Date = *input.Date
	}
	if input.Status != "" {
		status, err := validateEnum(input.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid status: "+err.Error()))
			return
		}
		record.Status = status
	}

	ctx := c.Request.Context()
	if err := h.svc.MedicalRecord.Update(ctx, record); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// DeleteMedicalRecord godoc
func (h *Handler) DeleteMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.MedicalRecord.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterMedicalRecordRoutes はカルテ関連のルートを登録する
func (h *Handler) RegisterMedicalRecordRoutes(rg *gin.RouterGroup) {
	records := rg.Group("/medical-records")
	records.GET("", h.ListMedicalRecords)
	records.POST("", h.CreateMedicalRecord)
	records.GET("/:id", h.GetMedicalRecord)
	records.PATCH("/:id", h.UpdateMedicalRecord)
	records.DELETE("/:id", h.DeleteMedicalRecord)

	h.RegisterVitalRoutes(records)
	h.RegisterTreatmentRoutes(records)
	h.RegisterBillingReviewRoutes(records)
	h.RegisterRecordImageRoutes(records)
	h.RegisterTreatmentPlanMedicalRecordRoutes(records)
	h.RegisterClinicalPlanRoutes(records)
	h.RegisterCheckupRoutes(records)
	h.RegisterInquiryRoutes(records)
}
