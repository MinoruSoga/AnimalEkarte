package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	record, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

// CreateMedicalRecord godoc
// FE↔BE契約統一: visit_date (string) + date (time.Time) 両対応、record_no自動生成、ID型変換、ClinicalPlan自動作成
func (h *Handler) CreateMedicalRecord(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input createMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	ctx := c.Request.Context()

	// 1. 日付の解決: date (time.Time) または visit_date (string "YYYY-MM-DD")
	var recordDate time.Time
	if input.Date != nil {
		recordDate = *input.Date
	} else if input.VisitDate != nil && *input.VisitDate != "" {
		parsed, err := time.Parse("2006-01-02", *input.VisitDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_date format (expected YYYY-MM-DD)"})
			return
		}
		recordDate = parsed
	} else {
		recordDate = time.Now()
	}

	// 2. ID型の変換: string → uint64
	var ownerID, petID, doctorID, reservationAppointmentID *uint64
	if input.OwnerID == nil || *input.OwnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}
	if input.PetID == nil || *input.PetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pet_id is required"})
		return
	}

	if input.OwnerID != nil && *input.OwnerID != "" {
		id, err := strconv.ParseUint(*input.OwnerID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id (must be numeric)"})
			return
		}
		ownerID = &id
	}
	if input.PetID != nil && *input.PetID != "" {
		id, err := strconv.ParseUint(*input.PetID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id (must be numeric)"})
			return
		}
		petID = &id
	}
	if input.DoctorID != nil && *input.DoctorID != "" {
		id, err := strconv.ParseUint(*input.DoctorID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor_id (must be numeric)"})
			return
		}
		doctorID = &id
	}
	if input.ReservationAppointmentID != nil && *input.ReservationAppointmentID != "" {
		id, err := strconv.ParseUint(*input.ReservationAppointmentID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation_appointment_id (must be numeric)"})
			return
		}
		reservationAppointmentID = &id
	}

	// 3. MedicalRecord モデル組み立て（RecordNoはservice層で自動生成）
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
			return
		}
		record.Status = status
	}

	// 5. MedicalRecord 作成
	if err := h.svc.MedicalRecord.Create(ctx, record); err != nil {
		RespondError(c, err)
		return
	}

	// 6. ClinicalPlan の自動作成・更新（chief_complaint, plan, assessment, diagnosis_* が送られた場合）
	if input.Plan != nil || input.Assessment != nil || input.Diagnosis1CategoryID != nil || input.Diagnosis1NameID != nil {
		clinicalPlanInput := &service.UpdateClinicalPlanInput{
			TreatmentPolicy:  input.Plan,       // plan → TreatmentPolicy
			DiagnosisDetails: input.Assessment, // assessment → DiagnosisDetails
		}
		// Diagnosis1 を優先（複数診断の場合は拡張必要）
		if input.Diagnosis1CategoryID != nil {
			clinicalPlanInput.DiagnosisCategoryID = input.Diagnosis1CategoryID
		}
		if input.Diagnosis1NameID != nil {
			clinicalPlanInput.DiagnosisNameID = input.Diagnosis1NameID
		}
		// v9.1: Diagnosis2 も追加
		if input.Diagnosis2CategoryID != nil {
			clinicalPlanInput.Diagnosis2CategoryID = input.Diagnosis2CategoryID
		}
		if input.Diagnosis2NameID != nil {
			clinicalPlanInput.Diagnosis2NameID = input.Diagnosis2NameID
		}
		if _, err := h.svc.ClinicalPlan.Update(ctx, record.ID, clinicalPlanInput); err != nil {
			slog.ErrorContext(ctx, "failed to update clinical plan",
				slog.String("error", err.Error()),
				slog.Uint64("medical_record_id", record.ID))
			// ClinicalPlan 作成失敗は医療記録は作成済みなので、エラーログのみ
		}
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateMedicalRecordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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
