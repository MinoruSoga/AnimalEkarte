package medicalrecord

// EMR-495 review round 1 regression tests:
//   A-1  absent / null diagnoses must stay "leave unchanged" end to end (explicit [] = clear)
//   A-2  diagnosis shape errors are rejected at request conversion, before any parent is created
//   SPEC-01  retired single diagnosis fields are rejected with 400 instead of being silently ignored

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func postCreateMedicalRecord(t *testing.T, h *MedicalRecordHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	c.Set("user_id", "1")
	h.CreateMedicalRecord(c)
	return w
}

func TestToDiagnosisInputs_NilIsDistinctFromExplicitEmpty(t *testing.T) {
	absent, err := toDiagnosisInputs(nil)
	require.NoError(t, err)
	assert.Nil(t, absent, "absent / null diagnoses must stay nil (leave unchanged)")

	cleared, err := toDiagnosisInputs([]clinicalPlanDiagnosisRequest{})
	require.NoError(t, err)
	require.NotNil(t, cleared, "explicit [] must stay a non-nil empty slice (clear)")
	assert.Empty(t, cleared)
}

func TestCreateMedicalRecordRequest_ToSubRecordsInput_DiagnosesAbsentVsEmpty(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantNil   bool
		wantEmpty bool
	}{
		{name: "key absent", body: `{"owner_id":"1","pet_id":"2"}`, wantNil: true},
		{name: "explicit null", body: `{"owner_id":"1","pet_id":"2","diagnoses":null}`, wantNil: true},
		{name: "explicit empty array", body: `{"owner_id":"1","pet_id":"2","diagnoses":[]}`, wantEmpty: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req createMedicalRecordRequest
			require.NoError(t, json.Unmarshal([]byte(tc.body), &req))
			input, err := req.toSubRecordsInput()
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, input.Diagnoses)
			}
			if tc.wantEmpty {
				require.NotNil(t, input.Diagnoses)
				assert.Empty(t, input.Diagnoses)
			}
		})
	}
}

// A-1: 同一予約の既存カルテへ再送された POST（既存ヒット）は診断を送らない。CreateSubRecords へ渡る
// Diagnoses は nil のままで、既存の診断行は置換（全クリア）されない。
func TestCreateMedicalRecord_WithoutDiagnoses_PassesNilToSubRecords(t *testing.T) {
	var got CreateSubRecordsInput
	called := false
	svc := &mockMedicalRecordService{
		createSubRecordsFn: func(_ context.Context, _, _ uint64, input CreateSubRecordsInput) error {
			called = true
			got = input
			return nil
		},
	}
	h := newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{})

	w := postCreateMedicalRecord(t, h, `{"owner_id":"10","pet_id":"20","visit_date":"2026-03-24","plan":"経過観察"}`)

	require.Equal(t, http.StatusCreated, w.Code)
	require.True(t, called)
	assert.Nil(t, got.Diagnoses, "a POST without diagnoses must not become a replace-with-empty request")
}

// A-1（サービス層）: 既存の診断行を持つ plan に対し、Diagnoses が nil の CreateSubRecords は
// 診断行を一切書かない（ReplaceDiagnoses / DeleteDiagnoses を呼ばない）。
func TestCreateSubRecords_NilDiagnosesKeepsExistingRows(t *testing.T) {
	replaceCalls := 0
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 7}, nil
		},
		listDiagnosesFn: func(_ context.Context, _, _ uint64) ([]model.ClinicalPlanDiagnosis, error) {
			return []model.ClinicalPlanDiagnosis{{ID: 1, ClinicalPlanID: 7, DiagnosisSlot: 1}}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error { return nil },
		replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
			replaceCalls++
			return nil
		},
		deleteDiagnosesFn: func(_ context.Context, _, _ uint64) error {
			replaceCalls++
			return nil
		},
	}
	svc := &medicalRecordService{
		inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: repo,
		diagTypeRepo: okDiagnosisTypeRepo(), diagNameRepo: okDiagnosisNameRepo(),
	}

	require.NoError(t, svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Plan: strPtr("policy")}))
	require.NoError(t, svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{}))

	assert.Zero(t, replaceCalls, "existing diagnosis rows must survive a request that omits diagnoses")
}

// A-1: 明示的な [] だけが全クリアとして ReplaceDiagnoses(空) に届く。
func TestCreateSubRecords_ExplicitEmptyDiagnosesClears(t *testing.T) {
	var replaced [][]ClinicalPlanDiagnosisInput
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 7}, nil
		},
		replaceDiagnosesFn: func(_ context.Context, _, _ uint64, rows []ClinicalPlanDiagnosisInput) error {
			replaced = append(replaced, rows)
			return nil
		},
	}
	svc := &medicalRecordService{
		inquiryRepo: &mockInquiryRepository{}, clinicalPlanRepo: repo,
		diagTypeRepo: okDiagnosisTypeRepo(), diagNameRepo: okDiagnosisNameRepo(),
	}

	require.NoError(t, svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{Diagnoses: []ClinicalPlanDiagnosisInput{}}))

	require.Len(t, replaced, 1)
	assert.Empty(t, replaced[0])
}

// A-2: 形の不正（slot 範囲外・51 行・slot 欠落）は親カルテ作成前に 400。親も子も一切作られない。
func TestCreateMedicalRecord_InvalidDiagnosisShape_Returns400WithoutPersisting(t *testing.T) {
	rows := make([]string, 0, maxClinicalPlanDiagnosisRows+1)
	for i := 0; i <= maxClinicalPlanDiagnosisRows; i++ {
		rows = append(rows, `{"diagnosis_slot":1}`)
	}
	cases := map[string]string{
		"slot 3":          `"diagnoses":[{"diagnosis_slot":3}]`,
		"slot 0":          `"diagnoses":[{"diagnosis_slot":0}]`,
		"51 rows":         `"diagnoses":[` + strings.Join(rows, ",") + `]`,
		"missing slot":    `"diagnoses":[{"diagnosis_name_snapshot":"x"}]`,
		"one bad of many": `"diagnoses":[{"diagnosis_slot":1},{"diagnosis_slot":3}]`,
		"slot wrong type": `"diagnoses":[{"diagnosis_slot":"one"}]`,
	}
	for name, diag := range cases {
		t.Run(name, func(t *testing.T) {
			createCalled, subCalled := false, false
			svc := &mockMedicalRecordService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					createCalled = true
					return &model.MedicalRecord{ID: 1}, nil
				},
				createSubRecordsFn: func(_ context.Context, _, _ uint64, _ CreateSubRecordsInput) error {
					subCalled = true
					return nil
				},
			}
			h := newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{})

			w := postCreateMedicalRecord(t, h, `{"owner_id":"10","pet_id":"20","visit_date":"2026-03-24",`+diag+`}`)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.False(t, createCalled, "no parent record may be created")
			assert.False(t, subCalled, "no child records may be written")
		})
	}
}

func TestUpdateClinicalPlanRequest_InvalidDiagnosisShapeIsRejectedAtConversion(t *testing.T) {
	slot3 := int16(3)
	_, err := updateClinicalPlanRequest{Diagnoses: &[]clinicalPlanDiagnosisRequest{{DiagnosisSlot: &slot3}}}.toServiceInput()
	require.Error(t, err)

	tooMany := make([]clinicalPlanDiagnosisRequest, maxClinicalPlanDiagnosisRows+1)
	slot1 := int16(1)
	for i := range tooMany {
		tooMany[i] = clinicalPlanDiagnosisRequest{DiagnosisSlot: &slot1}
	}
	_, err = updateClinicalPlanRequest{Diagnoses: &tooMany}.toServiceInput()
	require.Error(t, err)
}

func legacyCheck(t *testing.T, body string) error {
	t.Helper()
	keys, err := parseSingleJSONObject([]byte(body))
	require.NoError(t, err)
	return rejectLegacyDiagnosisFields(keys)
}

func TestRejectLegacyDiagnosisFields(t *testing.T) {
	for name := range legacyDiagnosisFieldNames {
		t.Run(name, func(t *testing.T) {
			err := legacyCheck(t, `{"`+name+`": null}`)
			require.Error(t, err, "presence alone is rejected, even with a null value")
			assert.Contains(t, err.Error(), "diagnoses[]")
			assert.Contains(t, err.Error(), name)
		})
	}
	assert.NoError(t, legacyCheck(t, `{"diagnoses":[{"diagnosis_slot":1,"diagnosis_type_id":3}],"physical_exam":"x"}`),
		"diagnosis_type_id inside a diagnoses[] row is the new shape and must not be rejected")
}

// R2-02: 事前検査は「ちょうど 1 つの JSON object + EOF」だけを通す。
func TestParseSingleJSONObject(t *testing.T) {
	valid := []string{`{}`, `{"a":1}`, ` {"a":1}` + "\n", `{"status":null}`}
	for _, body := range valid {
		_, err := parseSingleJSONObject([]byte(body))
		assert.NoError(t, err, body)
	}
	invalid := map[string]string{
		"trailing second object": `{"diagnosis_type_id":3,"treatment_policy":"x"} {}`,
		"trailing garbage":       `{"a":1} xyz`,
		"trailing comma value":   `{"a":1},`,
		"top-level array":        `[{"a":1}]`,
		"top-level scalar":       `123`,
		"top-level string":       `"x"`,
		"top-level null":         `null`,
		"truncated":              `{"a":`,
		"not json":               `not json`,
	}
	for name, body := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := parseSingleJSONObject([]byte(body))
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestRejectStatusKeyOnUpdate_DecidedByKeyPresence(t *testing.T) {
	for _, value := range []string{`null`, `"draft"`, `"finalized"`, `""`, `123`, `true`, `[]`, `["draft"]`, `{}`, `{"a":1}`} {
		t.Run(value, func(t *testing.T) {
			keys, err := parseSingleJSONObject([]byte(`{"status":` + value + `}`))
			require.NoError(t, err)
			err = rejectStatusKeyOnUpdate(keys)
			require.Error(t, err)
			assert.True(t, errors.Is(err, apperrors.ErrUnprocessable))
		})
	}
	keys, err := parseSingleJSONObject([]byte(`{"date":"2026-03-24"}`))
	require.NoError(t, err)
	assert.NoError(t, rejectStatusKeyOnUpdate(keys))
}

func patchMedicalRecordRaw(t *testing.T, h *MedicalRecordHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	c.Set("user_id", "1")
	h.UpdateMedicalRecord(c)
	return w
}

func failingUpdateService(t *testing.T) *mockMedicalRecordService {
	return &mockMedicalRecordService{
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
			t.Fatal("Update must not be called")
			return nil, nil
		},
	}
}

// R2-01: PATCH は status キーの存在だけで固定 422。値の型・null に依存せず、service を呼ばない。
func TestUpdateMedicalRecord_StatusKeyPresentIs422ForAnyValue(t *testing.T) {
	values := map[string]string{
		"null":             `null`,
		"string draft":     `"draft"`,
		"string finalized": `"finalized"`,
		"empty string":     `""`,
		"number":           `123`,
		"bool":             `true`,
		"array":            `["draft"]`,
		"empty array":      `[]`,
		"object":           `{"a":1}`,
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(failingUpdateService(t), &mockClinicalPlanService{})
			w := patchMedicalRecordRaw(t, h, `{"status":`+value+`}`)
			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), apperrors.CodeMedicalRecordStatusRetired)
		})
	}
	t.Run("null combined with valid update fields", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(failingUpdateService(t), &mockClinicalPlanService{})
		w := patchMedicalRecordRaw(t, h, `{"status":null,"date":"2026-03-24T00:00:00Z","version":2,"visit_type":"first"}`)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	})
}

// POST の status 扱いは変更なし: 未指定 / null / draft / 空は受理、finalized は 422。
func TestCreateMedicalRecord_StatusBehaviourUnchanged(t *testing.T) {
	for _, status := range []string{`"draft"`, `""`, `null`} {
		created := false
		svc := &mockMedicalRecordService{
			createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
				created = true
				return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date}, nil
			},
		}
		w := postCreateMedicalRecord(t, newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{}),
			`{"owner_id":"10","pet_id":"20","status":`+status+`}`)
		assert.Equal(t, http.StatusCreated, w.Code, status)
		assert.True(t, created, status)
	}
	svc := &mockMedicalRecordService{
		createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
			t.Fatal("Create must not be called for status finalized")
			return nil, nil
		},
	}
	w := postCreateMedicalRecord(t, newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{}),
		`{"owner_id":"10","pet_id":"20","status":"finalized"}`)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// R2-02: 末尾の別 JSON・ゴミ・最上位の配列/スカラー・null は 400。旧診断キーを後続値に隠しても書込は起きない。
func TestMalformedBodies_Return400WithoutWrite(t *testing.T) {
	bodies := map[string]string{
		"legacy key then trailing object": `{"diagnosis_type_id":3,"treatment_policy":"x"} {}`,
		"trailing garbage":                `{"date":"2026-03-24T00:00:00Z"} garbage`,
		"top-level array":                 `[{"date":"2026-03-24T00:00:00Z"}]`,
		"top-level scalar":                `42`,
		"top-level null":                  `null`,
		"empty body":                      ``,
	}
	for name, body := range bodies {
		t.Run("update/"+name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(failingUpdateService(t), &mockClinicalPlanService{})
			w := patchMedicalRecordRaw(t, h, body)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
		t.Run("create/"+name, func(t *testing.T) {
			svc := &mockMedicalRecordService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					t.Fatal("Create must not be called")
					return nil, nil
				},
				createSubRecordsFn: func(_ context.Context, _, _ uint64, _ CreateSubRecordsInput) error {
					t.Fatal("CreateSubRecords must not be called")
					return nil
				},
			}
			w := postCreateMedicalRecord(t, newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{}), body)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
		t.Run("clinical-plan/"+name, func(t *testing.T) {
			svc := &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateClinicalPlanInput) (*ClinicalPlanView, error) {
					t.Fatal("clinical plan Update must not be called")
					return nil, nil
				},
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/8/clinical-plan", strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "8"}}
			setClinicID(c)
			c.Set("user_id", "42")
			newHandlerWithClinicalPlanSvc(svc).UpdateClinicalPlan(c)
			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
}

func TestUpdateClinicalPlan_LegacyDiagnosisFields_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, field := range []string{"diagnosis_type_id", "diagnosis_name_id", "diagnosis_category_id", "diagnosis2_type_id", "diagnosis_2_name_id"} {
		t.Run(field, func(t *testing.T) {
			svc := &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateClinicalPlanInput) (*ClinicalPlanView, error) {
					t.Fatal("service must not be called when a retired diagnosis field is sent")
					return nil, nil
				},
			}
			h := newHandlerWithClinicalPlanSvc(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/8/clinical-plan",
				bytes.NewReader([]byte(`{"physical_exam":"x","`+field+`":5}`)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "8"}}
			setClinicID(c)
			c.Set("user_id", "42")

			h.UpdateClinicalPlan(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "diagnoses[]")
		})
	}
}

func TestCreateMedicalRecord_LegacyDiagnosisFields_Returns400WithoutPersisting(t *testing.T) {
	for _, field := range []string{"diagnosis_1_category_id", "diagnosis_1_name_id", "diagnosis1_name_id", "diagnosis_2_type_id", "diagnosis2_name_id", "diagnosis_type_id"} {
		t.Run(field, func(t *testing.T) {
			createCalled := false
			svc := &mockMedicalRecordService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					createCalled = true
					return &model.MedicalRecord{ID: 1}, nil
				},
			}
			h := newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{})

			w := postCreateMedicalRecord(t, h, `{"owner_id":"10","pet_id":"20","`+field+`":3}`)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "diagnoses[]")
			assert.False(t, createCalled)
		})
	}
}

func TestUpdateMedicalRecord_LegacyDiagnosisFields_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockMedicalRecordService{
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
			t.Fatal("service must not be called when a retired diagnosis field is sent")
			return nil, nil
		},
	}
	h := newHandlerWithMedicalRecordSvc(svc, &mockClinicalPlanService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"diagnosis_name_id":9}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	c.Set("user_id", "1")

	h.UpdateMedicalRecord(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
