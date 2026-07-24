package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/infra"
	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/middleware"
)

const (
	lineWebhookRequestsPerSecond = 50
	lineWebhookBurst             = 200
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg := config.Load()
	initializeLogger(cfg)
	if err := initializeRuntimeConfig(cfg); err != nil {
		return 1
	}

	db, sqlDB, err := openRuntimeDatabase(cfg)
	if err != nil {
		return 1
	}
	databaseOwnedByRunner := false
	defer func() {
		if databaseOwnedByRunner {
			return
		}
		closeRuntimeDatabase(sqlDB)
	}()

	execution, err := prepareRuntimeExecution(cfg, db, sqlDB)
	if err != nil {
		return 1
	}
	databaseOwnedByRunner = true
	if err := execution.run(); err != nil {
		logger.Error(
			"server lifecycle failed",
			slog.String("error", err.Error()),
		)
		return 1
	}
	logger.Info("server stopped")
	return 0
}

func initializeLogger(cfg *config.Config) {
	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	logger.Init(logger.Config{
		Level:  logLevel,
		Format: "json",
		Output: os.Stdout,
	})
	logger.Info("starting Animal Ekarte API v2.0 (45 tables)")
}

func initializeRuntimeConfig(cfg *config.Config) error {
	if err := config.ConfigureTimeZone(); err != nil {
		slog.Error(
			"timezone configuration failed",
			slog.String("error", err.Error()),
		)
		return err
	}
	if err := cfg.Validate(); err != nil {
		slog.Error(
			"config validation failed",
			slog.String("error", err.Error()),
		)
		return err
	}
	gin.SetMode(cfg.GinMode)
	return nil
}

func initializeIntegrationCipher(
	cfg *config.Config,
) (*appcrypto.AESGCMCipher, error) {
	if cfg.IntegrationEncryptionKey == "" {
		logger.Info(
			"INTEGRATION_ENCRYPTION_KEY not set: running without encryption (dev mode)",
		)
		return nil, nil
	}
	cipher, err := appcrypto.NewAESGCMCipher(cfg.IntegrationEncryptionKey)
	if err != nil {
		logger.Error(
			"failed to initialize AES-GCM cipher",
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	logger.Info("integration cipher: AES-256-GCM enabled")
	return cipher, nil
}

func initializeSharedStorage(
	cfg *config.Config,
) (infra.FileStorage, error) {
	if cfg.StorageType != "s3" {
		logger.Info("shared file storage: local filesystem")
		return infra.NewLocalFileStorage(
			"/app/uploads/shared",
			"http://localhost:"+cfg.Port+"/uploads/shared",
		), nil
	}
	storage, err := infra.NewS3FileStorage(
		context.Background(),
		cfg.S3SharedBucket,
		cfg.S3SharedRegion,
		cfg.S3Endpoint,
	)
	if err != nil {
		logger.Error(
			"failed to initialize S3FileStorage",
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	logger.Info(
		"shared file storage: S3",
		slog.String("bucket", cfg.S3SharedBucket),
	)
	return storage, nil
}

func initializeUploader(
	cfg *config.Config,
) (infra.FileUploader, error) {
	if cfg.StorageType != "s3" {
		logger.Info("file uploader: local filesystem")
		return infra.NewLocalUploader("/app/uploads", "/uploads"), nil
	}
	uploader, err := infra.NewS3Uploader(
		context.Background(),
		cfg.S3Bucket,
		cfg.S3Region,
		cfg.S3Endpoint,
		cfg.S3PublicBaseURL,
	)
	if err != nil {
		logger.Error(
			"failed to initialize S3 uploader",
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	logger.Info(
		"file uploader: S3",
		slog.String("bucket", cfg.S3Bucket),
		slog.String("region", cfg.S3Region),
	)
	if cfg.S3Endpoint != "" && cfg.S3PublicBaseURL == "" {
		logger.Warn(
			"S3_PUBLIC_BASE_URL 未設定: オブジェクト公開 URL が S3 API ホストを指しブラウザから参照できません。R2 の公開ドメイン(custom domain / *.r2.dev)を S3_PUBLIC_BASE_URL に設定してください",
			slog.String("s3_endpoint", cfg.S3Endpoint),
		)
	}
	return uploader, nil
}

func newRouter(cfg *config.Config) (*gin.Engine, error) {
	router := gin.New()
	trustedProxies := []string{"127.0.0.1"}
	if cfg.GinMode == "release" {
		trustedProxies = nil
		if cfg.TrustedProxyCIDR != "" {
			trustedProxies = []string{cfg.TrustedProxyCIDR}
			slog.Info("rate-limit: trusting configured proxy CIDR")
		}
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		logger.Error(
			"failed to set trusted proxies",
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigin))
	router.Use(middleware.RequestLoggingMiddleware())
	router.Use(middleware.SanitizeNullBytes())
	return router, nil
}

func newHTTPServer(
	cfg *config.Config,
	router *gin.Engine,
) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
}

func runtimeDrainers(composition runtimeComposition) []func() {
	return []func(){
		func() {
			logger.Info(
				"draining in-flight password reset email goroutines...",
			)
			composition.auth.DrainPasswordReset()
		},
		func() {
			logger.Info(
				"draining in-flight reservation notification goroutines...",
			)
			composition.reservation.DrainNotifications()
		},
		func() {
			logger.Info(
				"draining in-flight checkup followup trigger goroutines...",
			)
			composition.medicalRecord.DrainCheckups()
		},
	}
}
