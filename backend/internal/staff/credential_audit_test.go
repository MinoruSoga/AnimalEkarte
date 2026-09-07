package staff

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

type credentialAuditService struct {
	Service
	result    *model.Staff
	err       error
	calls     *int
	lastInput *UpdateStaffInput
}

func (s *credentialAuditService) Update(
	_ context.Context,
	_, _ uint64,
	input *UpdateStaffInput,
) (*model.Staff, error) {
	if s.calls != nil {
		*s.calls++
	}
	s.lastInput = input
	return s.result, s.err
}

func executeCredentialUpdate(
	t *testing.T,
	handler *Handler,
	body map[string]any,
	configure ...func(*gin.Context),
) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/masters/staffs/29",
		bytes.NewReader(encoded),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "staff-credential-audit-test")
	c.Request.RemoteAddr = "192.0.2.29:1234"
	c.Params = gin.Params{{Key: "id", Value: "29"}}
	c.Set("clinic_id", "23")
	c.Set("clinic_ids", []uint64{23})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	for _, configureContext := range configure {
		configureContext(c)
	}
	handler.UpdateStaff(c)
	return response
}

func TestHandler_UpdateStaffPasswordPassesExplicitCredentialAuditMetadata(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	service := &credentialAuditService{
		result: &model.Staff{
			ID:        29,
			ClinicID:  23,
			AccountID: &accountID,
			Name:      "Updated Staff",
		},
	}
	handler := NewHandlerWithPermissionChecker(
		service, nil, nil, nil, nil, nil,
		func(_ *gin.Context, _, _ string) bool { return true },
	)

	response := executeCredentialUpdate(
		t,
		handler,
		map[string]any{"password": "newPassw0rd"},
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NotNil(t, service.lastInput)
	require.NotNil(t, service.lastInput.CredentialAudit)
	audit := service.lastInput.CredentialAudit
	assert.Equal(t, uint64(23), audit.ClinicID)
	assert.Equal(t, uint64(17), audit.ActorStaffID)
	assert.Equal(t, uint64(29), audit.TargetStaffID)
	assert.Equal(t, "192.0.2.29", audit.IPAddress)
	assert.Equal(t, "staff-credential-audit-test", audit.UserAgent)

	encoded, err := json.Marshal(audit)
	require.NoError(t, err)
	for _, secret := range []string{"newPassw0rd", "token", "email"} {
		assert.NotContains(t, string(encoded), secret)
	}
}

func TestHandler_UpdateStaffWithoutPasswordOmitsCredentialAuditMetadata(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	service := &credentialAuditService{
		result: &model.Staff{ID: 29, ClinicID: 23, Name: "Updated Staff"},
	}
	handler := NewHandler(service, nil, nil, nil, nil, nil)

	response := executeCredentialUpdate(
		t,
		handler,
		map[string]any{"name": "Updated Staff"},
	)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NotNil(t, service.lastInput)
	assert.Nil(t, service.lastInput.CredentialAudit)
}

func TestHandler_UpdateStaffPasswordRejectsMissingActorBeforeMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updateCalls := 0
	service := &credentialAuditService{
		result: &model.Staff{ID: 29},
		calls:  &updateCalls,
	}
	handler := NewHandler(service, nil, nil, nil, nil, nil)

	response := executeCredentialUpdate(
		t,
		handler,
		map[string]any{"password": "newPassw0rd"},
		func(c *gin.Context) {
			c.Set("user_id", "")
		},
	)

	assert.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Zero(t, updateCalls)
}
