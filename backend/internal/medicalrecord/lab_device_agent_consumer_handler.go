package medicalrecord

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/labdevicecap"
	"github.com/animal-ekarte/backend/internal/model"
)

const labDeviceAgentConsumerTokenEnv = "LAB_DEVICE_AGENT_CONSUMER_TOKEN"

// labDeviceAgentConsumerUnconfiguredCode は秘密未設定の 503 にのみ付く
// 機械可読 code。FE は status+code の一致でのみ「未設定」表示へ変換し、
// Worker/Container 起因の code なし 503 とは区別する（PR647-F1）。
const labDeviceAgentConsumerUnconfiguredCode = "lab_device_agent_consumer_unconfigured"

// labDeviceAgentConsumerResponse mirrors the OpenAPI LabDeviceAgentConsumer
// schema (enforced by TestLabDeviceOpenAPIResponseParity).
type labDeviceAgentConsumerResponse struct {
	AgentConsumerToken string `json:"agent_consumer_token"`
	ExpiresIn          int    `json:"expires_in"`
}

// GetLabDeviceAgentConsumer returns a short-lived capability for the local agent.
// codex-security ad9e0152 finding-1: the shared env value is an HMAC signing
// secret, never returned. The capability is bound to the selected clinic that
// the lab-import grant was verified for, so a credential issued for clinic A
// is rejected by clinic B's workstation agent.
// GET /api/v1/lab-device/agent-consumer
func (h *LabImportHandler) GetLabDeviceAgentConsumer(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "create") {
		return
	}
	secret := os.Getenv(labDeviceAgentConsumerTokenEnv)
	if secret == "" {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error": "lab device agent consumer is not configured",
			"code":  labDeviceAgentConsumerUnconfiguredCode,
		})
		return
	}
	capability, err := labdevicecap.Issue(secret, strconv.FormatUint(clinicID, 10), time.Now())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, labDeviceAgentConsumerResponse{
		AgentConsumerToken: capability,
		ExpiresIn:          int(labdevicecap.TTL.Seconds()),
	})
}
