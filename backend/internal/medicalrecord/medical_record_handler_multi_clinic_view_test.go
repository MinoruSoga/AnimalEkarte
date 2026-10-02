package medicalrecord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-164 / 改善依頼 No.28①の許可側ケースを固定する:
// 「複数医院に所属していれば横断で飼主情報やカルテを閲覧できる（閲覧のみ）」。
// 医院A+Bの両方に medical-records:view 権限を持つスタッフは、一覧では
// clinic_ids=1,2 をそのままスコープにでき、詳細では選択医院Aのまま医院Bの
// カルテを開ける。付与無し医院への漏洩側は
// medical_record_handler_selected_clinic_b_grant_a_test.go が固定済み。
// 書き込みは ExtractClinicID で選択医院のみに束縛される（UpdateMedicalRecord 等）。
func membershipABGrantsABSelectedA(c *gin.Context) {
	c.Set("clinic_id", "1")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, id uint64, res, act string) bool {
		return (id == 1 || id == 2) && res == string(model.ResourceMedicalRecords) && act == "view"
	})
}

func TestListMedicalRecords_MembershipABGrantsAB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("clinic_ids=1,2 は両医院をスコープに渡し横断カルテを返す", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
			listFn: func(_ context.Context, clinicIDs []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				assert.Equal(t, []uint64{1, 2}, clinicIDs)
				return []model.MedicalRecord{
					{ID: 11, ClinicID: 1, RecordNo: "A-001"},
					{ID: 22, ClinicID: 2, RecordNo: "B-001"},
				}, 2, nil
			},
		}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantsABSelectedA(c)
		h.ListMedicalRecords(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"record_no":"A-001"`)
		assert.Contains(t, w.Body.String(), `"record_no":"B-001"`)
		assert.Contains(t, w.Body.String(), `"total":2`)
	})

	t.Run("clinic_ids 未指定は従来どおり選択医院のみにスコープする", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
			listFn: func(_ context.Context, clinicIDs []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				assert.Equal(t, []uint64{1}, clinicIDs)
				return []model.MedicalRecord{{ID: 11, ClinicID: 1, RecordNo: "A-001"}}, 1, nil
			},
		}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantsABSelectedA(c)
		h.ListMedicalRecords(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"record_no":"A-001"`)
		assert.NotContains(t, w.Body.String(), `"record_no":"B-001"`)
	})
}

func TestGetMedicalRecord_MembershipABGrantsAB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("選択医院Aのまま所属医院Bのカルテ詳細を閲覧できる", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
			getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
				require.Equal(t, []uint64{1, 2}, clinicIDs)
				return &model.MedicalRecord{ID: id, ClinicID: 2, RecordNo: "B-001"}, nil
			},
		}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "22"}}
		membershipABGrantsABSelectedA(c)
		h.GetMedicalRecord(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"clinic_id":2`)
	})
}
