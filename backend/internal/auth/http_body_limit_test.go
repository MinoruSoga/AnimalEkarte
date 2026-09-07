package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

const testAuthJSONBodyLimit = 16 * 1024

type bodyLimitPermissionGroupService struct{}

func (bodyLimitPermissionGroupService) List(
	context.Context,
	uint64,
) ([]model.PermissionGroup, error) {
	return nil, nil
}

func (bodyLimitPermissionGroupService) GetByID(
	_ context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
}

func (bodyLimitPermissionGroupService) Create(
	context.Context,
	uint64,
	*CreatePermissionGroupInput,
	PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	return &model.PermissionGroup{ID: 1}, nil
}

func (bodyLimitPermissionGroupService) Update(
	_ context.Context,
	clinicID, id uint64,
	_ *UpdatePermissionGroupInput,
	_ PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
}

func (bodyLimitPermissionGroupService) Delete(
	context.Context,
	uint64,
	uint64,
	PermissionMutationAudit,
) error {
	return nil
}

func (bodyLimitPermissionGroupService) UpdateRules(
	context.Context,
	uint64,
	uint64,
	[]SetPermissionGroupRulesInput,
	uint64,
	PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	return &model.PermissionGroup{}, nil
}

func (bodyLimitPermissionGroupService) Reorder(context.Context, uint64, []uint64) error {
	return nil
}

type bodyLimitAuthCallService struct {
	Service
	calls int
}

func (s *bodyLimitAuthCallService) AuthenticateUser(
	context.Context,
	string,
	string,
) (*model.Account, *model.Staff, error) {
	s.calls++
	return nil, nil, apperrors.WrapUnauthorized("invalid credentials")
}

type bodyLimitPasswordResetCallService struct {
	PasswordResetService
	forgotCalls int
	resetCalls  int
}

func (s *bodyLimitPasswordResetCallService) ForgotPassword(
	context.Context,
	string,
) error {
	s.forgotCalls++
	return nil
}

func (s *bodyLimitPasswordResetCallService) ResetPassword(
	context.Context,
	string,
	string,
) error {
	s.resetCalls++
	return nil
}

func (s *bodyLimitPasswordResetCallService) ResetPasswordWithResult(
	context.Context,
	string,
	string,
	CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	s.resetCalls++
	return &PasswordResetCompletion{AccountID: 1}, nil
}

func (*bodyLimitPasswordResetCallService) Wait() {}

type bodyLimitStaffCallReader struct {
	getCalls int
}

func (s *bodyLimitStaffCallReader) GetByID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	s.getCalls++
	return nil, apperrors.WrapNotFound("staff", "lookup")
}

func (*bodyLimitStaffCallReader) FindByAccountID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return nil, nil
}

type bodyLimitPermissionCallService struct {
	PermissionGroupService
	createCalls      int
	getByIDCalls     int
	updateRulesCalls int
	getByIDError     error
}

func (s *bodyLimitPermissionCallService) Create(
	_ context.Context,
	clinicID uint64,
	input *CreatePermissionGroupInput,
	_ PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	s.createCalls++
	return &model.PermissionGroup{
		ID:       1,
		ClinicID: clinicID,
		Name:     input.Name,
		Color:    input.Color,
	}, nil
}

func (s *bodyLimitPermissionCallService) GetByID(
	_ context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	s.getByIDCalls++
	if s.getByIDError != nil {
		return nil, s.getByIDError
	}
	return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
}

func (s *bodyLimitPermissionCallService) UpdateRules(
	context.Context,
	uint64,
	uint64,
	[]SetPermissionGroupRulesInput,
	uint64,
	PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	s.updateRulesCalls++
	return &model.PermissionGroup{}, nil
}

func streamedOversizedAuthJSONRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()
	payload := `{"padding":"` + strings.Repeat("x", testAuthJSONBodyLimit) + `"}`
	request := httptest.NewRequest(
		method,
		target,
		io.NopCloser(strings.NewReader(payload)),
	)
	request.Header.Set("Content-Type", "application/json")
	require.Equal(t, int64(-1), request.ContentLength)
	return request
}

func TestHTTPHandler_JSONBodiesRejectStreamedPayloadsOver16KiB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHTTPHandler(HTTPDependencies{
		PermissionGroups: bodyLimitPermissionGroupService{},
	}, CookieConfigForProduction(false))

	tests := []struct {
		name      string
		method    string
		target    string
		configure func(*gin.Context)
		handle    gin.HandlerFunc
	}{
		{name: "public login", method: http.MethodPost, target: "/api/v1/login", handle: handler.Login},
		{
			name:   "public forgot password",
			method: http.MethodPost,
			target: "/api/v1/auth/forgot-password",
			handle: handler.ForgotPassword,
		},
		{
			name:   "public reset password",
			method: http.MethodPost,
			target: "/api/v1/auth/reset-password",
			handle: handler.ResetPassword,
		},
		{
			name:   "protected password change",
			method: http.MethodPut,
			target: "/api/v1/users/me/password",
			handle: handler.ChangeMyPassword,
		},
		{
			name:   "permission group create",
			method: http.MethodPost,
			target: "/api/v1/masters/permission-groups",
			configure: func(c *gin.Context) {
				c.Set("clinic_id", "1")
			},
			handle: handler.CreatePermissionGroup,
		},
		{
			name:   "permission group update",
			method: http.MethodPatch,
			target: "/api/v1/masters/permission-groups/7",
			configure: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Params = gin.Params{{Key: "id", Value: "7"}}
			},
			handle: handler.UpdatePermissionGroup,
		},
		{
			name:   "permission group rules",
			method: http.MethodPut,
			target: "/api/v1/masters/permission-groups/7/rules",
			configure: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Params = gin.Params{{Key: "id", Value: "7"}}
			},
			handle: handler.SetPermissionGroupRules,
		},
		{
			name:   "permission group reorder",
			method: http.MethodPatch,
			target: "/api/v1/masters/permission-groups/reorder",
			configure: func(c *gin.Context) {
				c.Set("clinic_id", "1")
			},
			handle: handler.ReorderPermissionGroups,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = streamedOversizedAuthJSONRequest(
				t,
				test.method,
				test.target,
			)
			if test.configure != nil {
				test.configure(c)
			}

			test.handle(c)

			assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
			assert.JSONEq(
				t,
				`{"error":"request body exceeds size limit"}`,
				response.Body.String(),
			)
		})
	}
}

func TestHTTPHandler_JSONBodiesRejectOversizedTrailingDataBeforeDependencies(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	authCalls := &bodyLimitAuthCallService{}
	passwordResetCalls := &bodyLimitPasswordResetCallService{}
	staffCalls := &bodyLimitStaffCallReader{}
	permissionCalls := &bodyLimitPermissionCallService{}
	handler := NewHTTPHandler(HTTPDependencies{
		Auth:             authCalls,
		PasswordReset:    passwordResetCalls,
		Staff:            staffCalls,
		PermissionGroups: permissionCalls,
	}, CookieConfigForProduction(false))
	oversizedTrailingData := strings.Repeat(" ", testAuthJSONBodyLimit)

	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		configure func(*gin.Context)
		handle    gin.HandlerFunc
		callCount func() int
	}{
		{
			name:   "public login",
			method: http.MethodPost,
			target: "/api/v1/login",
			body:   `{"email":"user@example.test","password":"password1"}`,
			handle: handler.Login,
			callCount: func() int {
				return authCalls.calls
			},
		},
		{
			name:   "public forgot password",
			method: http.MethodPost,
			target: "/api/v1/auth/forgot-password",
			body:   `{"email":"user@example.test"}`,
			handle: handler.ForgotPassword,
			callCount: func() int {
				return passwordResetCalls.forgotCalls
			},
		},
		{
			name:   "public reset password",
			method: http.MethodPost,
			target: "/api/v1/auth/reset-password",
			body:   `{"token":"token","password":"password1"}`,
			handle: handler.ResetPassword,
			callCount: func() int {
				return passwordResetCalls.resetCalls
			},
		},
		{
			name:   "protected password change",
			method: http.MethodPut,
			target: "/api/v1/users/me/password",
			body:   `{"current_password":"password1","new_password":"password2"}`,
			configure: func(c *gin.Context) {
				c.Set("user_id", "2")
			},
			handle: handler.ChangeMyPassword,
			callCount: func() int {
				return staffCalls.getCalls
			},
		},
		{
			name:   "permission group create",
			method: http.MethodPost,
			target: "/api/v1/masters/permission-groups",
			body:   `{"name":"group","color":"#123456"}`,
			configure: func(c *gin.Context) {
				c.Set("clinic_id", "1")
				c.Set("user_id", "2")
			},
			handle: handler.CreatePermissionGroup,
			callCount: func() int {
				return permissionCalls.createCalls
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeCalls := test.callCount()
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				test.method,
				test.target,
				io.NopCloser(strings.NewReader(test.body+oversizedTrailingData)),
			)
			c.Request.Header.Set("Content-Type", "application/json")
			require.Equal(t, int64(-1), c.Request.ContentLength)
			if test.configure != nil {
				test.configure(c)
			}

			test.handle(c)

			assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
			assert.JSONEq(
				t,
				`{"error":"request body exceeds size limit"}`,
				response.Body.String(),
			)
			assert.Equal(t, beforeCalls, test.callCount())
		})
	}
}

func TestHTTPHandler_SetPermissionGroupRulesRejectsOversizedBodyBeforeLookup(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	payload := `{"rules":[]}` + strings.Repeat(" ", testAuthJSONBodyLimit)
	tests := []struct {
		name          string
		requestBody   func() io.Reader
		contentLength func(*testing.T, *http.Request)
	}{
		{
			name: "known content length",
			requestBody: func() io.Reader {
				return strings.NewReader(payload)
			},
			contentLength: func(t *testing.T, request *http.Request) {
				t.Helper()
				require.Greater(t, request.ContentLength, int64(testAuthJSONBodyLimit))
			},
		},
		{
			name: "streamed content length",
			requestBody: func() io.Reader {
				return io.NopCloser(strings.NewReader(payload))
			},
			contentLength: func(t *testing.T, request *http.Request) {
				t.Helper()
				require.Equal(t, int64(-1), request.ContentLength)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			permissionCalls := &bodyLimitPermissionCallService{
				getByIDError: apperrors.WrapNotFound("permission_group", "7"),
			}
			handler := NewHTTPHandler(
				HTTPDependencies{PermissionGroups: permissionCalls},
				CookieConfigForProduction(false),
			)
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				http.MethodPut,
				"/api/v1/masters/permission-groups/7/rules",
				test.requestBody(),
			)
			c.Request.Header.Set("Content-Type", "application/json")
			test.contentLength(t, c.Request)
			c.Set("clinic_id", "1")
			c.Set("user_id", "2")
			c.Params = gin.Params{{Key: "id", Value: "7"}}

			handler.SetPermissionGroupRules(c)

			assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
			assert.JSONEq(
				t,
				`{"error":"request body exceeds size limit"}`,
				response.Body.String(),
			)
			assert.Zero(t, permissionCalls.getByIDCalls)
			assert.Zero(t, permissionCalls.updateRulesCalls)
		})
	}
}

func TestHTTPHandler_LoginRejectsASecondJSONValueBeforeAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authCalls := &bodyLimitAuthCallService{}
	handler := NewHTTPHandler(
		HTTPDependencies{Auth: authCalls},
		CookieConfigForProduction(false),
	)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		io.NopCloser(strings.NewReader(
			`{"email":"user@example.test","password":"password1"} {"extra":true}`,
		)),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	require.Equal(t, int64(-1), c.Request.ContentLength)

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{"error":"invalid request body"}`, response.Body.String())
	assert.Zero(t, authCalls.calls)
}

func TestBindAuthJSON_RejectsOverlongAuthFieldsAndCollections(t *testing.T) {
	rules := make([]PermissionGroupRuleInput, 101)
	for i := range rules {
		rules[i].Resource = "owners"
	}
	ids := make([]uint64, 101)
	for i := range ids {
		ids[i] = uint64(i + 1)
	}

	tests := []struct {
		name            string
		body            any
		destination     func() any
		expectedMessage string
	}{
		{
			name: "login email",
			body: LoginInput{
				Email:    strings.Repeat("a", 246) + "@test.com",
				Password: "password1",
			},
			destination:     func() any { return &LoginInput{} },
			expectedMessage: "254 以下",
		},
		{
			name: "login password",
			body: LoginInput{
				Email:    "user@example.test",
				Password: strings.Repeat("a", 72) + "1",
			},
			destination:     func() any { return &LoginInput{} },
			expectedMessage: "72 以下",
		},
		{
			name: "login password byte length",
			body: LoginInput{
				Email:    "user@example.test",
				Password: strings.Repeat("界", 25),
			},
			destination:     func() any { return &LoginInput{} },
			expectedMessage: "72 バイト以下",
		},
		{
			name: "forgot password email",
			body: ForgotPasswordRequest{
				Email: strings.Repeat("a", 246) + "@test.com",
			},
			destination:     func() any { return &ForgotPasswordRequest{} },
			expectedMessage: "254 以下",
		},
		{
			name: "reset token",
			body: ResetPasswordRequest{
				Token:    strings.Repeat("t", 129),
				Password: "password1",
			},
			destination:     func() any { return &ResetPasswordRequest{} },
			expectedMessage: "128 以下",
		},
		{
			name: "reset password",
			body: ResetPasswordRequest{
				Token:    "token",
				Password: strings.Repeat("a", 72) + "1",
			},
			destination:     func() any { return &ResetPasswordRequest{} },
			expectedMessage: "72 以下",
		},
		{
			name: "reset password byte length",
			body: ResetPasswordRequest{
				Token:    "token",
				Password: strings.Repeat("界", 25),
			},
			destination:     func() any { return &ResetPasswordRequest{} },
			expectedMessage: "72 バイト以下",
		},
		{
			name: "change current password",
			body: ChangeMyPasswordRequest{
				CurrentPassword: strings.Repeat("a", 73),
				NewPassword:     "password1",
			},
			destination:     func() any { return &ChangeMyPasswordRequest{} },
			expectedMessage: "72 以下",
		},
		{
			name: "change current password byte length",
			body: ChangeMyPasswordRequest{
				CurrentPassword: strings.Repeat("界", 25),
				NewPassword:     "password1",
			},
			destination:     func() any { return &ChangeMyPasswordRequest{} },
			expectedMessage: "72 バイト以下",
		},
		{
			name: "change new password",
			body: ChangeMyPasswordRequest{
				CurrentPassword: "password1",
				NewPassword:     strings.Repeat("a", 73),
			},
			destination:     func() any { return &ChangeMyPasswordRequest{} },
			expectedMessage: "72 以下",
		},
		{
			name: "change new password byte length",
			body: ChangeMyPasswordRequest{
				CurrentPassword: "password1",
				NewPassword:     strings.Repeat("界", 25),
			},
			destination:     func() any { return &ChangeMyPasswordRequest{} },
			expectedMessage: "72 バイト以下",
		},
		{
			name: "permission group name",
			body: CreatePermissionGroupRequest{
				Name:  strings.Repeat("n", 256),
				Color: "#123456",
			},
			destination:     func() any { return &CreatePermissionGroupRequest{} },
			expectedMessage: "255 以下",
		},
		{
			name: "permission group description",
			body: CreatePermissionGroupRequest{
				Name:        "group",
				Description: strings.Repeat("d", 2001),
				Color:       "#123456",
			},
			destination:     func() any { return &CreatePermissionGroupRequest{} },
			expectedMessage: "2000 以下",
		},
		{
			name: "permission group color",
			body: CreatePermissionGroupRequest{
				Name:  "group",
				Color: "#1234567",
			},
			destination: func() any { return &CreatePermissionGroupRequest{} },
		},
		{
			name: "updated permission group name",
			body: map[string]any{"name": strings.Repeat("n", 256)},
			destination: func() any {
				return &UpdatePermissionGroupRequest{}
			},
			expectedMessage: "255 以下",
		},
		{
			name: "updated permission group description",
			body: map[string]any{"description": strings.Repeat("d", 2001)},
			destination: func() any {
				return &UpdatePermissionGroupRequest{}
			},
			expectedMessage: "2000 以下",
		},
		{
			name: "updated permission group color",
			body: map[string]any{"color": "#1234567"},
			destination: func() any {
				return &UpdatePermissionGroupRequest{}
			},
		},
		{
			name: "permission rule resource",
			body: SetPermissionGroupRulesRequest{
				Rules: []PermissionGroupRuleInput{{
					Resource: strings.Repeat("r", 51),
				}},
			},
			destination:     func() any { return &SetPermissionGroupRulesRequest{} },
			expectedMessage: "50 以下",
		},
		{
			name: "permission rule collection",
			body: SetPermissionGroupRulesRequest{Rules: rules},
			destination: func() any {
				return &SetPermissionGroupRulesRequest{}
			},
			expectedMessage: "100 以下",
		},
		{
			name: "permission reorder collection",
			body: ReorderPermissionGroupsRequest{IDs: ids},
			destination: func() any {
				return &ReorderPermissionGroupsRequest{}
			},
			expectedMessage: "100 以下",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.body)
			require.NoError(t, err)
			require.LessOrEqual(t, int64(len(encoded)), authJSONBodyMaxBytes)
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/auth-field-limit",
				bytes.NewReader(encoded),
			)
			c.Request.Header.Set("Content-Type", "application/json")

			bindError := bindAuthJSON(c, test.destination())

			require.Error(t, bindError)
			assert.ErrorIs(t, bindError, apperrors.ErrInvalidInput)
			assert.NotErrorIs(t, bindError, apperrors.ErrPayloadTooLarge)
			if test.expectedMessage != "" {
				assert.Contains(t, bindError.Error(), test.expectedMessage)
			}
		})
	}
}
