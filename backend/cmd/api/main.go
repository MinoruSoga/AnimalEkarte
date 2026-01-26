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
// @description.en Animal Hospital Electronic Medical Record System API
// @termsOfService http://localhost:8080/terms
// @contact.name Animal Ekarte Support
// @contact.url http://localhost:8080/support
// @contact.email support@animal-ekarte.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description API Key for authentication

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

	// マイグレーション（全モデルを依存関係順に登録）
	if err := db.AutoMigrate(
		// 独立テーブル
		&model.Clinic{},
		&model.InventoryItem{},
		&model.Cage{},
		// Clinic依存
		&model.Staff{},
		// InventoryItem依存
		&model.MasterItem{},
		// コアテーブル
		&model.Owner{},
		&model.Pet{},
		// Pet依存
		&model.MedicalRecord{},
		&model.Reservation{},
		&model.Hospitalization{},
		&model.Vaccination{},
		&model.Trimming{},
		&model.Examination{},
		&model.Accounting{},
		// Hospitalization依存
		&model.CarePlanItem{},
		&model.DailyRecord{},
		// DailyRecord依存
		&model.Vital{},
		&model.CareLog{},
		&model.StaffNote{},
		// Accounting依存
		&model.AccountingItem{},
	); err != nil {
		logger.Error("failed to migrate database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database migrated successfully (20 tables)")

	// レイヤー初期化
	repo := repository.New(db)
	medicalRecordRepo := repository.NewMedicalRecordRepository(db)
	svc := service.New(repo, repo, medicalRecordRepo, repo)
	h := handler.New(svc)

	// ルーター設定
	r := gin.Default()
	h.RegisterRoutes(r)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	logger.Info("server stopped")
}
