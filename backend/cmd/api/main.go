package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/handler"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/animal-ekarte/backend/docs"
)

// @title Animal Ekarte API
// @version 1.0
// @description 動物病院 電子カルテシステム API
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// ロガー初期化
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger.Init(logger.Config{
		Level:  logLevel,
		Format: "json",
		Output: os.Stdout,
	})

	logger.Info("starting Animal Ekarte API")

	// 設定読み込み
	cfg := config.Load()

	// DB接続
	db, err := repository.NewDB(cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected successfully")

	// マイグレーション
	if err := db.AutoMigrate(&model.Pet{}); err != nil {
		logger.Error("failed to migrate database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database migrated successfully")

	// レイヤー初期化
	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	// ルーター設定
	r := gin.Default()
	h.RegisterRoutes(r)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	logger.Info("server starting",
		slog.String("port", cfg.Port),
		slog.String("swagger_url", "http://localhost:8080/swagger/index.html"),
	)

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + cfg.Port); err != nil {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// シャットダウン処理（5秒タイムアウト）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	logger.Info("server stopped")
}
