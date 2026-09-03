package medicalrecord

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestGetLabDeviceAgentConsumer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		token      string
		setupCtx   func(*gin.Context)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing configuration fails closed",
			setupCtx:   setClinicID,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "returns configured token",
			token:      "consumer-token",
			setupCtx:   setClinicID,
			wantStatus: http.StatusOK,
			wantBody:   `{"agent_consumer_token":"consumer-token"}`,
		},
		{
			name:  "returns 403 when selected clinic lacks lab-import create grant",
			token: "consumer-token",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceLabImport), "create")
			},
			wantStatus: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(labDeviceAgentConsumerTokenEnv, test.token)
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device/agent-consumer", http.NoBody)
			test.setupCtx(context)

			(&LabImportHandler{}).GetLabDeviceAgentConsumer(context)

			assert.Equal(t, test.wantStatus, response.Code)
			if test.wantBody != "" {
				assert.JSONEq(t, test.wantBody, response.Body.String())
			}
		})
	}
}
