package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock CheckupFieldResultService ----

type mockCheckupFieldResultService struct {
	listFieldsFn        func(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
	listByCheckupFn     func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error)
	listByPetFn         func(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	replaceForCheckupFn func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error)
}

func (m *mockCheckupFieldResultService) ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	return m.listFieldsFn(ctx, clinicID, checkupTypeID)
}

func (m *mockCheckupFieldResultService) ListByCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	return m.listByCheckupFn(ctx, clinicID, medicalRecordID, checkupID)
}

func (m *mockCheckupFieldResultService) ListByPet(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	return m.listByPetFn(ctx, clinicID, petID)
}

func (m *mockCheckupFieldResultService) ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
	return m.replaceForCheckupFn(ctx, clinicID, medicalRecordID, checkupID, actorID, inputs)
}

func newHandlerWithCheckupFieldResultSvc(svc CheckupFieldResultService) *CheckupHandler {
	return NewCheckupHandler(nil, svc)
}

// ---- mock CheckupTypeFieldService（EMR-225 フィールド定義 write 側） ----

type mockCheckupTypeFieldService struct {
	createFieldFn   func(ctx context.Context, clinicID, checkupTypeID uint64, input *CreateCheckupTypeFieldInput) (*model.CheckupTypeField, error)
	updateFieldFn   func(ctx context.Context, clinicID, checkupTypeID, fieldID uint64, input *UpdateCheckupTypeFieldInput) (*model.CheckupTypeField, error)
	deleteFieldFn   func(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error
	reorderFieldsFn func(ctx context.Context, clinicID, checkupTypeID uint64, ids []uint64) error
}

func (m *mockCheckupTypeFieldService) CreateField(ctx context.Context, clinicID, checkupTypeID uint64, input *CreateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
	return m.createFieldFn(ctx, clinicID, checkupTypeID, input)
}

func (m *mockCheckupTypeFieldService) UpdateField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64, input *UpdateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
	return m.updateFieldFn(ctx, clinicID, checkupTypeID, fieldID, input)
}

func (m *mockCheckupTypeFieldService) DeleteField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error {
	return m.deleteFieldFn(ctx, clinicID, checkupTypeID, fieldID)
}

func (m *mockCheckupTypeFieldService) ReorderFields(ctx context.Context, clinicID, checkupTypeID uint64, ids []uint64) error {
	return m.reorderFieldsFn(ctx, clinicID, checkupTypeID, ids)
}

func newHandlerWithCheckupFieldSvc(svc CheckupTypeFieldService) *CheckupHandler {
	return NewCheckupHandler(nil, nil, svc)
}

// ---- ListCheckupTypeFields ----

func TestListCheckupTypeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of checkup type fields",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listFieldsFn: func(_ context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), checkupTypeID)
					return []model.CheckupTypeField{{ID: 1, CheckupTypeID: 1, Name: "体重", FieldType: model.CheckupFieldTypeNumber}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id param",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listFieldsFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListCheckupTypeFields(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListCheckupFieldResults ----

func TestListCheckupFieldResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		recordID   string
		checkupID  string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:      "returns list of checkup field results",
			recordID:  "1",
			checkupID: "2",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByCheckupFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					return []model.CheckupFieldResult{{ID: 1, CheckupID: 2, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			recordID:   "1",
			checkupID:  "2",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid medical record id",
			recordID:   "abc",
			checkupID:  "2",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			recordID:   "1",
			checkupID:  "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			recordID:  "1",
			checkupID: "2",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByCheckupFn: func(_ context.Context, _, _, _ uint64) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.recordID}, {Key: "checkupId", Value: tt.checkupID}}
			tt.setupCtx(c)

			h.ListCheckupFieldResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReplaceCheckupFieldResults ----

func TestReplaceCheckupFieldResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fieldID := uint64(10)

	tests := []struct {
		name       string
		recordID   string
		checkupID  string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:      "replaces field results successfully",
			recordID:  "1",
			checkupID: "2",
			body: map[string]any{
				"results": []map[string]any{
					{"checkup_type_field_id": 10, "value_number": 12.5},
				},
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				replaceForCheckupFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					require.Len(t, inputs, 1)
					require.NotNil(t, inputs[0].CheckupTypeFieldID)
					assert.Equal(t, fieldID, *inputs[0].CheckupTypeFieldID)
					return []model.CheckupFieldResult{{ID: 1, CheckupID: 2, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			recordID:   "1",
			checkupID:  "2",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid medical record id",
			recordID:   "abc",
			checkupID:  "2",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			recordID:   "1",
			checkupID:  "abc",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			recordID:   "1",
			checkupID:  "2",
			bodyRaw:    `{"results":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			recordID:  "1",
			checkupID: "2",
			body:      map[string]any{"results": []map[string]any{}},
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				replaceForCheckupFn: func(_ context.Context, _, _, _ uint64, _ *uint64, _ []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.recordID}, {Key: "checkupId", Value: tt.checkupID}}
			tt.setupCtx(c)

			h.ReplaceCheckupFieldResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListPetCheckupResults ----

func TestListPetCheckupResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns pet checkup results",
			query:    "pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByPetFn: func(_ context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), petID)
					return []model.CheckupFieldResult{{ID: 1, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "pet_id=5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pet_id",
			query:      "pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet_id is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByPetFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListPetCheckupResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestCheckupFieldSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		resource model.Resource
		invoke   func(*CheckupHandler, *gin.Context)
		svc      *mockCheckupFieldResultService
	}{
		{
			name:     "ListCheckupTypeFields returns 403 when selected clinic lacks checkups view grant",
			resource: model.ResourceCheckups,
			invoke: func(h *CheckupHandler, c *gin.Context) {
				h.ListCheckupTypeFields(c)
			},
			svc: &mockCheckupFieldResultService{
				listFieldsFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name:     "ListCheckupFieldResults returns 403 when selected clinic lacks medical record view grant",
			resource: model.ResourceMedicalRecords,
			invoke: func(h *CheckupHandler, c *gin.Context) {
				h.ListCheckupFieldResults(c)
			},
			svc: &mockCheckupFieldResultService{
				listByCheckupFn: func(_ context.Context, _, _, _ uint64) ([]model.CheckupFieldResult, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name:     "ListPetCheckupResults returns 403 when selected clinic lacks checkups view grant",
			resource: model.ResourceCheckups,
			invoke: func(h *CheckupHandler, c *gin.Context) {
				h.ListPetCheckupResults(c)
			},
			svc: &mockCheckupFieldResultService{
				listByPetFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?pet_id=5", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}, {Key: "checkupId", Value: "3"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(tt.resource), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

// ---- フィールド定義 write 側（EMR-225） ----
// 権限強制はルート側 perm(model.ResourceCheckups, ...) ミドルウェアの責務（exam-type fields と同型）
// のため、ここではバインド→service 入力変換・ID 経路・fail-closed を固定する。

func newCheckupFieldWriteContext(w *httptest.ResponseRecorder, method, body string, params gin.Params) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	c.Request = httptest.NewRequest(method, "/", reader)
	c.Params = params
	setClinicID(c)
	return c
}

func TestCreateCheckupTypeField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("201 with Location header and converts request to service input", func(t *testing.T) {
		svc := &mockCheckupTypeFieldService{
			createFieldFn: func(_ context.Context, clinicID, checkupTypeID uint64, input *CreateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), checkupTypeID)
				assert.Equal(t, "総合評価", input.Name)
				assert.Equal(t, "single_select", input.FieldType)
				require.Len(t, input.Options, 2)
				assert.Equal(t, CheckupFieldOptionInput{Value: "a", Label: "良好"}, input.Options[0])
				return &model.CheckupTypeField{ID: 42, CheckupTypeID: 7, Name: input.Name, FieldType: model.CheckupFieldTypeSingleSelect}, nil
			},
		}
		h := newHandlerWithCheckupFieldSvc(svc)
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPost,
			`{"name":"総合評価","field_type":"single_select","options":[{"value":"a","label":"良好"},{"value":"b","label":"要注意"}],"sort_order":3}`,
			gin.Params{{Key: "id", Value: "7"}})

		h.CreateCheckupTypeField(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "/v1/masters/checkup-types/7/fields/42", w.Header().Get("Location"))
		assert.Contains(t, w.Body.String(), `"id":42`)
	})

	t.Run("400 on malformed body", func(t *testing.T) {
		h := newHandlerWithCheckupFieldSvc(&mockCheckupTypeFieldService{})
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPost, `{"field_type":"number"}`,
			gin.Params{{Key: "id", Value: "7"}})
		h.CreateCheckupTypeField(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 on non-numeric parent id", func(t *testing.T) {
		h := newHandlerWithCheckupFieldSvc(&mockCheckupTypeFieldService{
			createFieldFn: func(_ context.Context, _, _ uint64, _ *CreateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
				t.Fatal("service must not be reached")
				return nil, nil
			},
		})
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPost, `{"name":"x","field_type":"text"}`,
			gin.Params{{Key: "id", Value: "abc"}})
		h.CreateCheckupTypeField(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("500 when field service is not wired", func(t *testing.T) {
		h := NewCheckupHandler(nil, nil) // fieldService 未注入（旧2引数呼出し互換）
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPost, `{"name":"x","field_type":"text"}`,
			gin.Params{{Key: "id", Value: "7"}})
		h.CreateCheckupTypeField(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateCheckupTypeField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 and passes partial update through to service", func(t *testing.T) {
		svc := &mockCheckupTypeFieldService{
			updateFieldFn: func(_ context.Context, clinicID, checkupTypeID, fieldID uint64, input *UpdateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), checkupTypeID)
				assert.Equal(t, uint64(3), fieldID)
				require.NotNil(t, input.Name)
				assert.Equal(t, "体重（朝）", *input.Name)
				require.NotNil(t, input.FieldType)
				assert.Equal(t, "text", *input.FieldType, "field_type 変更を許可")
				assert.True(t, input.ClearMaxValue)
				return &model.CheckupTypeField{ID: 3, CheckupTypeID: 7, Name: *input.Name, FieldType: model.CheckupFieldTypeText}, nil
			},
		}
		h := newHandlerWithCheckupFieldSvc(svc)
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPatch,
			`{"name":"体重（朝）","field_type":"text","clear_max_value":true}`,
			gin.Params{{Key: "id", Value: "7"}, {Key: "fieldId", Value: "3"}})

		h.UpdateCheckupTypeField(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"field_type":"text"`)
	})

	t.Run("400 on non-numeric field id", func(t *testing.T) {
		h := newHandlerWithCheckupFieldSvc(&mockCheckupTypeFieldService{
			updateFieldFn: func(_ context.Context, _, _, _ uint64, _ *UpdateCheckupTypeFieldInput) (*model.CheckupTypeField, error) {
				t.Fatal("service must not be reached")
				return nil, nil
			},
		})
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPatch, `{"name":"x"}`,
			gin.Params{{Key: "id", Value: "7"}, {Key: "fieldId", Value: "zzz"}})
		h.UpdateCheckupTypeField(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteCheckupTypeField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("204 on success", func(t *testing.T) {
		svc := &mockCheckupTypeFieldService{
			deleteFieldFn: func(_ context.Context, clinicID, checkupTypeID, fieldID uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), checkupTypeID)
				assert.Equal(t, uint64(3), fieldID)
				return nil
			},
		}
		// c.Status(NoContent) のみでボディ書き込みが無いため gin.Engine 経由でヘッダーを
		// フラッシュする（直接呼び出しだと w.Code が 200 のまま — care_plan_item 先例）。
		h := newHandlerWithCheckupFieldSvc(svc)
		r := gin.New()
		r.DELETE("/masters/checkup-types/:id/fields/:fieldId", func(c *gin.Context) {
			setClinicID(c)
		}, h.DeleteCheckupTypeField)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/masters/checkup-types/7/fields/3", http.NoBody)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("404 propagates from service for missing field", func(t *testing.T) {
		svc := &mockCheckupTypeFieldService{
			deleteFieldFn: func(_ context.Context, _, _, _ uint64) error {
				return apperrors.WrapNotFound("checkup_type_field", "3")
			},
		}
		h := newHandlerWithCheckupFieldSvc(svc)
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodDelete, "",
			gin.Params{{Key: "id", Value: "7"}, {Key: "fieldId", Value: "3"}})
		h.DeleteCheckupTypeField(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestReorderCheckupTypeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("204 and forwards id order", func(t *testing.T) {
		svc := &mockCheckupTypeFieldService{
			reorderFieldsFn: func(_ context.Context, clinicID, checkupTypeID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), checkupTypeID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		// 204 のみのレスポンスは gin.Engine 経由でヘッダーをフラッシュする（上記 Delete 同型）。
		h := newHandlerWithCheckupFieldSvc(svc)
		r := gin.New()
		r.PATCH("/masters/checkup-types/:id/fields/reorder", func(c *gin.Context) {
			setClinicID(c)
		}, h.ReorderCheckupTypeFields)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/masters/checkup-types/7/fields/reorder",
			bytes.NewReader([]byte(`{"ids":[3,1,2]}`)))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400 on empty ids", func(t *testing.T) {
		h := newHandlerWithCheckupFieldSvc(&mockCheckupTypeFieldService{
			reorderFieldsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
				t.Fatal("service must not be reached")
				return nil
			},
		})
		w := httptest.NewRecorder()
		c := newCheckupFieldWriteContext(w, http.MethodPatch, `{"ids":[]}`,
			gin.Params{{Key: "id", Value: "7"}})
		h.ReorderCheckupTypeFields(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
