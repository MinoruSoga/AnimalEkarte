package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/service"
)

// Handler はHTTPハンドラーのルートコンテナ
type Handler struct {
	cfg *config.Config
	svc *service.Services
}

// New はHandlerを初期化して返す
func New(cfg *config.Config, svc *service.Services) *Handler {
	return &Handler{cfg: cfg, svc: svc}
}

// PaginatedResponse はページネーション付きレスポンスの共通構造
type PaginatedResponse[T any] struct {
	Data  T     `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// newPaginatedResponse はPaginatedResponseを型推論で生成するヘルパー
func newPaginatedResponse[T any](data T, total int64, page, limit int) PaginatedResponse[T] {
	return PaginatedResponse[T]{Data: data, Total: total, Page: page, Limit: limit}
}

// Health はサーバーの稼働状態を返す
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RegisterRoutes はすべてのルートを登録する
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/health", h.Health)
	api.POST("/login", h.Login)
	api.POST("/logout", h.Logout)

	protected := api.Group("")
	protected.Use(middleware.Auth(h.cfg.JWTSecret))

	protected.GET("/me", h.GetMe)

	h.RegisterOwnerRoutes(protected)
	h.RegisterPetRoutes(protected)
	h.RegisterReservationRoutes(protected)
	h.RegisterMedicalRecordRoutes(protected)
	h.RegisterHospitalizationRoutes(protected)
	h.RegisterAccountingRoutes(protected)
	h.RegisterTrimmingRoutes(protected)
	h.RegisterExaminationRoutes(protected)
	h.RegisterVaccinationRoutes(protected)
	h.RegisterInventoryRoutes(protected)
	h.RegisterMasterRoutes(protected)
	h.RegisterClinicRoutes(protected)
}
