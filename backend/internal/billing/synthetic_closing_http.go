package billing

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

const syntheticClosingCleanupHeader = "X-UAT-Cleanup-Token"

// SyntheticClosingHandler はローカル専用の S09 fixture HTTP。staging/production では 404。
type SyntheticClosingHandler struct {
	DB       *gorm.DB
	AppEnv   string
	DBHost   string
	Password string
}

type createSyntheticClosingHTTPRequest struct {
	TargetDate         string   `json:"targetDate"`
	ExistingBillingIDs []uint64 `json:"existingBillingIds"`
}

type syntheticClosingHTTPResponse struct {
	ClinicID     uint64   `json:"clinicId"`
	LoginEmail   string   `json:"loginEmail"`
	BillingIDs   []uint64 `json:"billingIds"`
	CompletedAt  []string `json:"completedAt"`
	CleanupToken string   `json:"cleanupToken"`
}

// RegisterUATRoutes は認証なしの S09 helper を /api/v1/uat に載せる。
func RegisterUATRoutes(rg *gin.RouterGroup, h *SyntheticClosingHandler) {
	if rg == nil || h == nil {
		return
	}
	uat := rg.Group("/uat")
	uat.POST("/synthetic-closings", h.CreateSyntheticClosings)
	uat.DELETE("/synthetic-closings/:clinic_id", h.DeleteSyntheticClosings)
}

// CreateSyntheticClosings は POST /api/v1/uat/synthetic-closings。
func (h *SyntheticClosingHandler) CreateSyntheticClosings(c *gin.Context) {
	if h == nil || !h.allowHTTP(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var req createSyntheticClosingHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	password := h.password()
	if password == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("password is required"))
		return
	}
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	day, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(req.TargetDate), jst)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("targetDate must be YYYY-MM-DD"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	result, err := CreateSyntheticClosingFixture(c.Request.Context(), h.DB, SyntheticClosingRequest{
		AppEnv:             h.appEnv(),
		DBHost:             h.dbHost(),
		TargetDate:         day,
		PasswordHash:       string(hash),
		ExistingBillingIDs: req.ExistingBillingIDs,
	})
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	completed := make([]string, 0, len(result.CompletedAt))
	for _, at := range result.CompletedAt {
		completed = append(completed, at.Format(time.RFC3339))
	}
	c.Header("Location", "/api/v1/uat/synthetic-closings/"+strconv.FormatUint(result.ClinicID, 10))
	c.JSON(http.StatusCreated, syntheticClosingHTTPResponse{
		ClinicID:     result.ClinicID,
		LoginEmail:   result.LoginEmail,
		BillingIDs:   result.BillingIDs,
		CompletedAt:  completed,
		CleanupToken: result.CleanupToken,
	})
}

// DeleteSyntheticClosings は DELETE /api/v1/uat/synthetic-closings/:clinic_id。
func (h *SyntheticClosingHandler) DeleteSyntheticClosings(c *gin.Context) {
	if h == nil || !h.allowHTTP(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	clinicID, err := strconv.ParseUint(c.Param("clinic_id"), 10, 64)
	if err != nil || clinicID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("clinic_id is required"))
		return
	}
	if err := DeleteSyntheticClosingFixture(
		c.Request.Context(),
		h.DB,
		h.appEnv(),
		h.dbHost(),
		clinicID,
		c.GetHeader(syntheticClosingCleanupHeader),
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SyntheticClosingHandler) allowHTTP(c *gin.Context) bool {
	if err := AllowUATSyntheticClosing(h.appEnv(), h.dbHost()); err != nil {
		return false
	}
	if err := AllowUATSyntheticClosingHTTPHost(c.Request.Host); err != nil {
		return false
	}
	return true
}

func (h *SyntheticClosingHandler) appEnv() string {
	if h != nil && strings.TrimSpace(h.AppEnv) != "" {
		return h.AppEnv
	}
	return os.Getenv("APP_ENV")
}

func (h *SyntheticClosingHandler) dbHost() string {
	if h != nil && strings.TrimSpace(h.DBHost) != "" {
		return h.DBHost
	}
	return os.Getenv("DB_HOST")
}

func (h *SyntheticClosingHandler) password() string {
	if h != nil && h.Password != "" {
		return h.Password
	}
	return os.Getenv("UAT_SYNTHETIC_CLOSING_PASSWORD")
}
