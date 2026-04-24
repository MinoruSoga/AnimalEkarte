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
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

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
	svcs := service.NewServices(repos, &service.ReservationNotificationConfig{
		SMTPHost:    cfg.SMTPHost,
		SMTPPort:    cfg.SMTPPort,
		SMTPUser:    cfg.SMTPUser,
		SMTPPass:    cfg.SMTPPass,
		SMTPFrom:    cfg.SMTPFrom,
		FrontendURL: cfg.FrontendURL,
	})

	// ファイルアップローダー初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	var uploader infra.FileUploader
	if os.Getenv("STORAGE_TYPE") == "s3" {
		s3Bucket := os.Getenv("S3_BUCKET")
		s3Region := os.Getenv("S3_REGION")
		if s3Bucket == "" || s3Region == "" {
			logger.Error("S3_BUCKET and S3_REGION are required when STORAGE_TYPE=s3")
			os.Exit(1)
		}
		s3Up, err := infra.NewS3Uploader(context.Background(), s3Bucket, s3Region)
		if err != nil {
			logger.Error("failed to initialize S3 uploader", slog.String("error", err.Error()))
			os.Exit(1)
		}
		uploader = s3Up
		logger.Info("file uploader: S3", slog.String("bucket", s3Bucket), slog.String("region", s3Region))
	} else {
		uploader = infra.NewLocalUploader("/app/uploads", "/uploads")
		logger.Info("file uploader: local filesystem")
	}

	// ハンドラー初期化
	h := handler.New(cfg, svcs, repos, uploader)

	// アプリケーションライフタイムコンテキスト（バックグラウンドゴルーチン管理用）
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// ルーター設定
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLoggingMiddleware())
	// BUG-067: POST/PATCH/PUT ボディから NULL バイトを除去（PostgreSQL エラー防止）
	r.Use(middleware.SanitizeNullBytes())
	h.RegisterRoutes(appCtx, r)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting",
		slog.String("port", cfg.Port),
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
