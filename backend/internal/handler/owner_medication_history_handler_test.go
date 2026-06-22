package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// TestGetOwnerMedicationHistory は #158 投薬歴エンドポイントの正常系を検証する。
// 期待: clinicID/ownerID をサービスへ委譲し、date は YYYY-MM-DD、各行に pet 情報が乗る。
func TestGetOwnerMedicationHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mrSvc := &mockMedicalRecordService{
		getOwnerMedicationHistoryFn: func(_ context.Context, clinicID, ownerID uint64, _, _ int) ([]service.OwnerMedicationHistoryItem, int64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(42), ownerID)
			return []service.OwnerMedicationHistoryItem{
				{TreatmentID: 11, RecordDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), PetID: 2, PetName: "タマ", MedicineName: "院内調剤薬", AdminRoute: "皮下", Quantity: 1},
				{TreatmentID: 9, RecordDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), PetID: 1, PetName: "ポチ", MedicineName: "アモキシシリン", AdminRoute: "経口", Quantity: 2, DoctorName: "山田先生"},
			}, 2, nil
		},
	}
	h := newHandlerWithMedicalRecordSvc(mrSvc, &mockClinicalPlanService{})

	r := gin.New()
	r.GET("/owners/:id/medication-history",
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
		h.RequirePermission(string(model.ResourceMedicalRecords), "view"),
		h.GetOwnerMedicationHistory,
	)

	req := httptest.NewRequest(http.MethodGet, "/owners/42/medication-history", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp PaginatedResponse[[]ownerMedicationHistoryResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(2), resp.Total)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "タマ", resp.Data[0].PetName)
	assert.Equal(t, uint64(2), resp.Data[0].PetID)
	assert.Equal(t, "2026-06-15", resp.Data[0].Date)
	assert.Equal(t, "院内調剤薬", resp.Data[0].MedicineName)
	assert.Equal(t, "山田先生", resp.Data[1].DoctorName)
}

// TestGetOwnerMedicationHistory_ViewPermissionDenied は view 権限を持たない非 system_admin が 403 を受けることを検証する（P5）。
func TestGetOwnerMedicationHistory_ViewPermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{svc: &service.Services{
		MedicalRecord:       &mockMedicalRecordService{},
		ClinicalPlan:        &mockClinicalPlanService{},
		Inquiry:             &mockInquiryService{},
		LstepTagSync:        &mockLstepTagSyncService{},
		EffectivePermission: &mockEffectivePermissionService{},
	}}
	r := gin.New()
	r.GET("/owners/:id/medication-history",
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
		h.RequirePermission(string(model.ResourceMedicalRecords), "view"),
		h.GetOwnerMedicationHistory,
	)

	req := httptest.NewRequest(http.MethodGet, "/owners/42/medication-history", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
