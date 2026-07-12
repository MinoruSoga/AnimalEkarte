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
	appCrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/logger"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

func main() {
	// 設定読み込み（ロガー初期化より先に行い、LOG_LEVEL を含む全設定を config.Config に一元化する）
	cfg := config.Load()

	// ロガー初期化
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

	if err := config.ConfigureTimeZone(); err != nil {
		slog.Error("timezone configuration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// H2: TRUSTED_PROXY_CIDR（release必須）・STORAGE_TYPE=s3 時の S3_BUCKET/S3_REGION 必須検証は
	// cfg.Validate() に集約済み（G9-2）
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

	// 連携設定の暗号化 cipher 初期化（INTEGRATION_ENCRYPTION_KEY 未設定時は dev モード・暗号化なし）。
	// lstep 連携（clinic_integrations）と LINE 予約設定（line_reservation_settings）で共有する（H-4）。
	// NewServices に渡すため NewServices 呼び出しより前に初期化する。
	var integrationCipher *appCrypto.AESGCMCipher
	if cfg.IntegrationEncryptionKey != "" {
		c, err := appCrypto.NewAESGCMCipher(cfg.IntegrationEncryptionKey)
		if err != nil {
			logger.Error("failed to initialize AES-GCM cipher", slog.String("error", err.Error()))
			os.Exit(1)
		}
		integrationCipher = c
		logger.Info("integration cipher: AES-256-GCM enabled")
	} else {
		logger.Info("INTEGRATION_ENCRYPTION_KEY not set: running without encryption (dev mode)")
	}

	// 共有ファイルストレージ初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	// NewServices に渡すため NewServices 呼び出しより前に初期化する（G9-1）。
	var sharedStorage infra.FileStorage
	if cfg.StorageType == "s3" {
		s3fs, err := infra.NewS3FileStorage(context.Background(), cfg.S3SharedBucket, cfg.S3SharedRegion, cfg.S3Endpoint)
		if err != nil {
			logger.Error("failed to initialize S3FileStorage", slog.String("error", err.Error()))
			os.Exit(1)
		}
		sharedStorage = s3fs
		logger.Info("shared file storage: S3", slog.String("bucket", cfg.S3SharedBucket))
	} else {
		sharedStorage = infra.NewLocalFileStorage("/app/uploads/shared", "http://localhost:"+cfg.Port+"/uploads/shared")
		logger.Info("shared file storage: local filesystem")
	}

	// サービス初期化（G9-1: 旧・二段階DIを単一段階に統合。LstepLifecycle/LstepTag/SharedFile/
	// ChronicCondition/LineSend/LineLink/LstepTagSummary/CheckupSync/LstepDeliveryMonitor/
	// LstepTriggerPriority/LstepDeliveryTrigger/MedicalRecord/Checkup/LstepBatch/LstepCsvImport/
	// LstepAnalytics はすべて service.NewServices 内で一括構築される）
	svcs := service.NewServices(repos, &service.ReservationNotificationConfig{
		SMTPHost:    cfg.SMTPHost,
		SMTPPort:    cfg.SMTPPort,
		SMTPUser:    cfg.SMTPUser,
		SMTPPass:    cfg.SMTPPass,
		SMTPFrom:    cfg.SMTPFrom,
		FrontendURL: cfg.FrontendURL,
	}, integrationCipher, sharedStorage)

	// ファイルアップローダー初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	var uploader infra.FileUploader
	if cfg.StorageType == "s3" {
		// S3_BUCKET/S3_REGION 必須検証は cfg.Validate() で起動時に済んでいる（G9-2）
		s3Up, err := infra.NewS3Uploader(context.Background(), cfg.S3Bucket, cfg.S3Region, cfg.S3Endpoint)
		if err != nil {
			logger.Error("failed to initialize S3 uploader", slog.String("error", err.Error()))
			os.Exit(1)
		}
		uploader = s3Up
		logger.Info("file uploader: S3", slog.String("bucket", cfg.S3Bucket), slog.String("region", cfg.S3Region))
	} else {
		uploader = infra.NewLocalUploader("/app/uploads", "/uploads")
		logger.Info("file uploader: local filesystem")
	}

	// ハンドラー初期化
	h := handler.New(cfg, svcs, repos, uploader)

	// アプリケーションライフタイムコンテキスト（バックグラウンドゴルーチン管理用）
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// LSTEP-BE-014: ノーショウ検知バッチ — 毎時0分に起動し 10/15/20 時 (JST) のみ実行
	// time.Local は main() 冒頭の config.ConfigureTimeZone() で Asia/Tokyo 確定済み（G9-3）
	go runScheduled(appCtx, "no-show batch", hourlyTick, func(ctx context.Context) error {
		triggerHours := map[int]bool{10: true, 15: true, 20: true}
		if !triggerHours[time.Now().Hour()] {
			return nil
		}
		return svcs.LstepBatch.RunNoShowCheckAllClinics(ctx)
	})

	// LSTEP-BE-004: 休眠検知バッチ — 毎日 02:00 JST に実行
	go runScheduled(appCtx, "dormant detection batch", dailyAt2AM, svcs.LstepBatch.RunDormantDetectionAllClinics)

	// FEAT-383: 自動配信トリガーバッチ — 毎時0分に起動（10:00 JST 固定）
	go runScheduled(appCtx, "delivery trigger batch", hourlyTick, svcs.LstepBatch.RunDeliveryTriggerBatchAllClinics)

	// ルーター設定
	r := gin.New()
	// H2: Set trusted proxies to prevent rate-limit bypass via X-Forwarded-For spoofing
	// TRUSTED_PROXY_CIDR validation is done earlier via cfg.Validate(), so only build list here
	var trustedProxies []string
	if cfg.GinMode == "release" {
		// Production: trust ALB CIDR
		// e.g., TRUSTED_PROXY_CIDR="10.0.0.0/8" for AWS ALB in private subnet
		if cfg.TrustedProxyCIDR != "" {
			trustedProxies = []string{cfg.TrustedProxyCIDR}
			slog.Info("rate-limit: trusting ALB CIDR")
		}
	} else {
		// Development: trust localhost only
		trustedProxies = []string{"127.0.0.1"}
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		// Log but continue - SetTrustedProxies failure is non-fatal
		slog.Warn("failed to set trusted proxies", slog.Any("error", err))
	}
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigin))
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

	// PERF-FOLLOWUP-05: パスワードリセットメール送信は fire-and-forget goroutine のため、
	// server.Shutdown の HTTP drain だけでは goroutine が孤児化する。ここで明示的に drain する。
	logger.Info("draining in-flight password reset email goroutines...")
	svcs.PasswordReset.Wait()

	logger.Info("server stopped")
}
