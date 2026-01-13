package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// Handler contains the HTTP handlers.
type Handler struct {
	svc service.PetService
}

// New creates a new Handler with the given service.
func New(svc service.PetService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// CORS設定
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

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
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Animal Ekarte API is running",
	})
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
