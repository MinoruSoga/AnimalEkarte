package staff

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestBindStaffJSONRejectsOversizedBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("declared content length", func(t *testing.T) {
		context := newStaffBindingContext(t, `{"name":"staff"}`)
		context.Request.ContentLength = staffJSONBodyMaxBytes + 1

		err := bindStaffJSON(context, &createStaffRequest{})

		require.Error(t, err)
		assert.True(t, apperrors.IsPayloadTooLarge(err))
	})

	t.Run("chunked body", func(t *testing.T) {
		body := `{"name":"` + strings.Repeat("x", int(staffJSONBodyMaxBytes)) + `"}`
		context := newStaffBindingContext(t, body)
		context.Request.ContentLength = -1

		err := bindStaffJSON(context, &createStaffRequest{})

		require.Error(t, err)
		assert.True(t, apperrors.IsPayloadTooLarge(err))
	})
}

func TestBindStaffJSONRejectsTrailingJSONValue(t *testing.T) {
	context := newStaffBindingContext(t, `{"name":"staff"} {"name":"second"}`)

	err := bindStaffJSON(context, &createStaffRequest{})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestBindStaffJSONEnforcesBcryptByteMaximum(t *testing.T) {
	password := strings.Repeat("あ", 24) + "1"
	context := newStaffBindingContext(
		t,
		`{"name":"staff","email":"staff@example.com","password":"`+password+`"}`,
	)

	err := bindStaffJSON(context, &createStaffRequest{})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestBindStaffJSONAcceptsValidRequest(t *testing.T) {
	context := newStaffBindingContext(
		t,
		`{"staff_id":7,"date":"2026-07-24","shift_type":"full","notes":"ok"}`,
	)
	request := &createShiftRequest{}

	require.NoError(t, bindStaffJSON(context, request))
	assert.Equal(t, uint64(7), request.StaffID)
	assert.Equal(t, "ok", request.Notes)
}

func TestStaffJSONRequestStringsHaveExplicitMaximums(t *testing.T) {
	requestTypes := []reflect.Type{
		reflect.TypeOf(createStaffRequest{}),
		reflect.TypeOf(updateStaffRequest{}),
		reflect.TypeOf(createOccupationRequest{}),
		reflect.TypeOf(updateOccupationRequest{}),
		reflect.TypeOf(shiftBreakRequest{}),
		reflect.TypeOf(createShiftRequest{}),
		reflect.TypeOf(updateShiftRequest{}),
		reflect.TypeOf(shiftTemplateBreakRequest{}),
		reflect.TypeOf(createShiftTemplateRequest{}),
		reflect.TypeOf(updateShiftTemplateRequest{}),
	}

	for _, requestType := range requestTypes {
		for index := 0; index < requestType.NumField(); index++ {
			field := requestType.Field(index)
			fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.String {
				continue
			}
			t.Run(requestType.Name()+"."+field.Name, func(t *testing.T) {
				assert.Contains(t, field.Tag.Get("binding"), "max=")
			})
		}
	}
}

func TestStaffHandlersUseBoundedJSONBinder(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	directory := filepath.Dir(currentFile)

	for _, fileName := range []string{
		"staff_handler.go",
		"occupation_handler.go",
		"shift_handler.go",
		"shift_template_handler.go",
	} {
		t.Run(fileName, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join(directory, fileName))
			require.NoError(t, err)
			assert.NotContains(t, string(source), "ShouldBindJSON")
			assert.Contains(t, string(source), "bindStaffJSON")
		})
	}
}

func TestUpdateStaffMapsOversizedBodyTo413(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/staffs/7",
		bytes.NewBufferString(`{"name":"`+strings.Repeat("x", int(staffJSONBodyMaxBytes))+`"}`),
	)
	context.Request.ContentLength = -1
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{{Key: "id", Value: "7"}}
	context.Set("clinic_id", "1")

	handler.UpdateStaff(context)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}

func newStaffBindingContext(t *testing.T, body string) *gin.Context {
	t.Helper()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return context
}
