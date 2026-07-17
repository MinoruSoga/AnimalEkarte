package handler

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/config"
)

// TestRegisterRoutes_NoPanic はルート登録時に panic が発生しないことを保証する。
//
// Gin のワイルドカード名衝突（例: :id vs :clinicId）は実行時 panic になるため、
// go build / golangci-lint では検出できない。このテストで CI 上で早期検出する。
//
// 実際に発生した障害: BUG-246 (2026-04-09)
//
//	reservation_line_routes.go の :clinicId/:staffId が既存の :id と衝突して
//	起動時 panic → backend コンテナ unhealthy。
func TestRegisterRoutes_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			JWTSecret: "test-secret-for-route-registration",
		},
	}

	assert.NotPanics(t, func() {
		r := gin.New()
		h.RegisterRoutes(context.Background(), r)
	})
}
