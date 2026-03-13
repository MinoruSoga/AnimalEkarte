package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/handler"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/animal-ekarte/backend/docs"
)

// @title Animal Ekarte API
// @version 2.0
// @description 動物病院 電子カルテシステム API（45テーブル・専用マスタテーブル版）
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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
	logger.Info("starting Animal Ekarte API v2.0 (45 tables)")

	// 設定読み込み
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("config validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	gin.SetMode(cfg.GinMode)

	// DB接続
	db, err := repository.NewDB(cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected")

	// リポジトリ初期化
	repos := repository.NewRepositories(db)

	// サービス初期化
	svcs := service.NewServices(repos)

	// ハンドラー初期化
	h := handler.New(cfg, svcs)

	// ルーター設定
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLoggingMiddleware())
	h.RegisterRoutes(r)
	// Swagger UI は開発・ステージング環境のみ有効
	if cfg.GinMode != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// HTTPサーバー設定
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting",
		slog.String("port", cfg.Port),
		slog.String("swagger_url", "http://localhost:"+cfg.Port+"/swagger/index.html"),
	)

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// シャットダウン処理（30秒タイムアウト）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
	logger.Info("server stopped")
}
