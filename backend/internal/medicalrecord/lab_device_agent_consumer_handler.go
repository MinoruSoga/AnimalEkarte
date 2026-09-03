package medicalrecord

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const labDeviceAgentConsumerTokenEnv = "LAB_DEVICE_AGENT_CONSUMER_TOKEN"

// GetLabDeviceAgentConsumer returns the configured credential for the local agent.
// GET /api/v1/lab-device/agent-consumer
func (h *LabImportHandler) GetLabDeviceAgentConsumer(c *gin.Context) {
	if _, ok := httpapi.ExtractClinicID(c); !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceLabImport), "create") {
		return
	}
	token := os.Getenv(labDeviceAgentConsumerTokenEnv)
	if token == "" {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	c.JSON(http.StatusOK, gin.H{"agent_consumer_token": token})
}
