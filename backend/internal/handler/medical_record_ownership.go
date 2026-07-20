package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// verifyMedicalRecordOwnership は clinicID + medicalRecordID の組み合わせを検証し、
// テナント分離を保証するヘルパー。検証済みの MedicalRecord と成否を返す。
//
// BE9-2D sub-batch④a: 元は medical_record_image_handler.go に定義されていたが、vital / image
// handler が internal/medicalrecord へ移設された。残留する treatment_plan_handler.go が引き続き
// このヘルパーを消費するため、共有カット点として handler package に残す（medicalrecord 側は
// route-level PermissionMiddleware + service.GetByID の consumer-side view でローカルに ownership を
// 担保する）。
func (h *Handler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) (*model.MedicalRecord, bool) {
	mr, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		RespondError(c, err)
		return nil, false
	}
	return mr, true
}
