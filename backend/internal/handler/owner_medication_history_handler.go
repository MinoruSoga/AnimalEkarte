package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// ownerMedicationHistoryResponse は飼い主の全ペット横断の投薬履歴1行（#158 飼主レポート）。
// どのペットの投薬かを各行に明示するため pet_id / pet_name を含む。
type ownerMedicationHistoryResponse struct {
	TreatmentID  uint64  `json:"treatment_id"`
	Date         string  `json:"date"` // YYYY-MM-DD（カルテ日付）
	PetID        uint64  `json:"pet_id"`
	PetName      string  `json:"pet_name"`
	MedicineName string  `json:"medicine_name"`
	AdminRoute   string  `json:"admin_route"`
	Quantity     float64 `json:"quantity"`
	DoctorName   string  `json:"doctor_name"`
}

func toOwnerMedicationHistoryResponse(r *service.OwnerMedicationHistoryItem) ownerMedicationHistoryResponse {
	return ownerMedicationHistoryResponse{
		TreatmentID:  r.TreatmentID,
		Date:         r.RecordDate.Format("2006-01-02"),
		PetID:        r.PetID,
		PetName:      r.PetName,
		MedicineName: r.MedicineName,
		AdminRoute:   r.AdminRoute,
		Quantity:     r.Quantity,
		DoctorName:   r.DoctorName,
	}
}

// GetOwnerMedicationHistory は飼い主の全ペット横断の投薬履歴を日付降順でページング返却する（#158 飼主レポート）。
// clinic 隔離は X-Clinic-ID から確定した単一医院スコープで行い、別医院の投薬は混入させない。
func (h *Handler) GetOwnerMedicationHistory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	rows, total, err := h.svc.MedicalRecord.GetOwnerMedicationHistory(c.Request.Context(), clinicID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(rows, toOwnerMedicationHistoryResponse), total, page, limit))
}
