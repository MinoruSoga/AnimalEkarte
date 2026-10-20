package medicalrecord

// EMR-456: 健診種別の保険設定を変更する create/update の HTTP 境界
// （選択医院 grant の再検査・actor 受け渡し・不正 payload・応答の保険設定）。

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func serveCheckupTypeMutation(t *testing.T, h *CheckupTypeHandler, method, body string, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	setup(c)
	if method == http.MethodPost {
		h.CreateCheckupType(c)
	} else {
		h.UpdateCheckupType(c)
	}
	return w
}

func TestCheckupTypeMutation_SelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	unreachable := &mockCheckupTypeService{
		createFn: func(context.Context, uint64, *CreateCheckupTypeInput) (*model.CheckupType, error) {
			t.Fatal("service must not be reached")
			return nil, nil
		},
		updateFn: func(context.Context, uint64, uint64, *UpdateCheckupTypeInput) (*model.CheckupType, error) {
			t.Fatal("service must not be reached")
			return nil, nil
		},
	}
	tests := []struct {
		name   string
		method string
		grant  string
	}{
		{name: "create with only view grant on selected clinic", method: http.MethodPost, grant: "view"},
		{name: "update with only view grant on selected clinic", method: http.MethodPatch, grant: "view"},
		{name: "update with edit grant only on another clinic", method: http.MethodPatch, grant: "edit-other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveCheckupTypeMutation(t, NewCheckupTypeHandler(unreachable), tt.method,
				`{"name":"a","is_non_insurance":false}`, func(c *gin.Context) {
					setClinicID(c)
					c.Set("clinic_id", "2")
					c.Set("is_system_admin", false)
					if tt.grant == "edit-other" {
						setResourcePermissionOnlyClinic(c, 1, string(model.ResourceCheckups), "edit")
						return
					}
					setResourcePermissionOnlyClinic(c, 2, string(model.ResourceCheckups), tt.grant)
				})
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

func TestCheckupTypeMutation_PassesActorAndReturnsSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotCreate *CreateCheckupTypeInput
	var gotUpdate *UpdateCheckupTypeInput
	svc := &mockCheckupTypeService{
		createFn: func(_ context.Context, _ uint64, in *CreateCheckupTypeInput) (*model.CheckupType, error) {
			gotCreate = in
			return &model.CheckupType{ID: 9, IsNonInsurance: boolPtr(true), Version: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, in *UpdateCheckupTypeInput) (*model.CheckupType, error) {
			gotUpdate = in
			return &model.CheckupType{ID: 9, IsNonInsurance: boolPtr(false), Version: 5}, nil
		},
	}
	withStaff := func(c *gin.Context) {
		setClinicID(c)
		c.Set("user_id", "42")
	}

	w := serveCheckupTypeMutation(t, NewCheckupTypeHandler(svc), http.MethodPost, `{"name":"健診"}`, withStaff)
	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, gotCreate.ActorID)
	assert.Equal(t, uint64(42), *gotCreate.ActorID)
	assert.Nil(t, gotCreate.IsNonInsurance, "absent create value is left to the service default")
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, true, created["is_non_insurance"])
	assert.EqualValues(t, 1, created["version"])

	w = serveCheckupTypeMutation(t, NewCheckupTypeHandler(svc), http.MethodPatch, `{"is_non_insurance":false,"version":4}`, withStaff)
	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, gotUpdate.ActorID)
	assert.Equal(t, uint64(42), *gotUpdate.ActorID)
	assert.False(t, *gotUpdate.IsNonInsurance)
	assert.Equal(t, int64(4), *gotUpdate.Version)
	var updated map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, false, updated["is_non_insurance"])
	assert.EqualValues(t, 5, updated["version"])
}

func TestCheckupTypeMutation_RejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	unreachable := &mockCheckupTypeService{
		createFn: func(context.Context, uint64, *CreateCheckupTypeInput) (*model.CheckupType, error) {
			t.Fatal("service must not be reached")
			return nil, nil
		},
		updateFn: func(context.Context, uint64, uint64, *UpdateCheckupTypeInput) (*model.CheckupType, error) {
			t.Fatal("service must not be reached")
			return nil, nil
		},
	}
	tests := []struct {
		name   string
		method string
		body   string
	}{
		{name: "create with string insurance value", method: http.MethodPost, body: `{"name":"a","is_non_insurance":"yes"}`},
		{name: "create with explicit null insurance", method: http.MethodPost, body: `{"name":"a","is_non_insurance":null}`},
		{name: "update with numeric insurance value", method: http.MethodPatch, body: `{"is_non_insurance":1,"version":4}`},
		{name: "update with explicit null insurance", method: http.MethodPatch, body: `{"is_non_insurance":null,"version":4}`},
		{name: "update with zero version", method: http.MethodPatch, body: `{"is_non_insurance":true,"version":0}`},
		{name: "update with negative version", method: http.MethodPatch, body: `{"is_non_insurance":true,"version":-1}`},
		{name: "update without version is rejected (required)", method: http.MethodPatch, body: `{"is_non_insurance":true}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveCheckupTypeMutation(t, NewCheckupTypeHandler(unreachable), tt.method, tt.body, setClinicID)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
