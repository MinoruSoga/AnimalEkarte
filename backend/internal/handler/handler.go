package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// Service defines the interface for the business logic used by the handler.
type Service interface {
	service.PetService
	service.OwnerService
	GetDB() (interface{ DB() *gorm.DB }, error)
}

// Handler contains the HTTP handlers.
type Handler struct {
	svc Service
}

// New creates a new Handler with the given service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ErrorResponse defines the standard error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// ミドルウェア設定
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())

	// ヘルスチェック
	r.GET("/health", h.Health)

	// API v1
	v1 := r.Group("/api/v1")
	v1.GET("/", h.Welcome)

	// Pets CRUD
	v1.GET("/pets", h.GetPets)
	v1.GET("/pets/:id", h.GetPet)
	v1.POST("/pets", h.CreatePet)
	v1.PUT("/pets/:id", h.UpdatePet)
	v1.DELETE("/pets/:id", h.DeletePet)
}

// Health godoc
// @Summary ヘルスチェック
// @Description APIの稼働状態を確認します
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"message":   "Animal Ekarte API is running",
	}

	// DB接続確認
	if db, err := h.svc.GetDB(); err == nil {
		if sqlDB, err := db.DB().DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				status["database"] = "connected"
			} else {
				status["database"] = "disconnected"
				status["database_error"] = err.Error()
			}
		} else {
			status["database"] = "error"
			status["database_error"] = err.Error()
		}
	} else {
		status["database"] = "error"
		status["database_error"] = err.Error()
	}

	c.JSON(http.StatusOK, status)
}

// Welcome godoc
// @Summary ウェルカムメッセージ
// @Description APIのウェルカムメッセージを返します
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func (h *Handler) Welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Animal Ekarte API",
	})
}
