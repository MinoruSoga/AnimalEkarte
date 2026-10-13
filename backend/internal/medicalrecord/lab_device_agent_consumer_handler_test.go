package medicalrecord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/labdevicecap"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestGetLabDeviceAgentConsumer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		token      string
		setupCtx   func(*gin.Context)
		wantStatus int
		wantCode   string
	}{
		{
			// PR647-F1: 未設定 503 は機械可読 code を持ち、FE が他原因の
			// 503（Worker/Container 障害）と区別できる。
			name:       "missing configuration fails closed",
			setupCtx:   setClinicID,
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "lab_device_agent_consumer_unconfigured",
		},
		{
			name:       "returns clinic-bound signed capability",
			token:      "consumer-token",
			setupCtx:   setClinicID,
			wantStatus: http.StatusOK,
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
			if test.wantCode != "" {
				var body struct {
					Code  string `json:"code"`
					Error string `json:"error"`
				}
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				assert.Equal(t, test.wantCode, body.Code)
				assert.NotEmpty(t, body.Error)
			}
			if test.wantStatus == http.StatusOK {
				var body struct {
					Capability string `json:"agent_consumer_token"`
					ExpiresIn  int    `json:"expires_in"`
				}
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.NotEmpty(t, body.Capability)
				assert.Equal(t, int(labdevicecap.TTL.Seconds()), body.ExpiresIn)
				// 選択中 clinic ("1") に束縛された capability — 他院 agent は拒否する。
				require.NoError(t, labdevicecap.Verify(test.token, "1", body.Capability, time.Now()))
				require.ErrorIs(t, labdevicecap.Verify(test.token, "2", body.Capability, time.Now()), labdevicecap.ErrClinicMismatch)
				// 生の共有シークレットがレスポンスに出ないこと。
				assert.NotContains(t, response.Body.String(), test.token)
			}
		})
	}
}
