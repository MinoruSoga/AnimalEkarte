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

	// LSTEP 暗号化 cipher 初期化（INTEGRATION_ENCRYPTION_KEY 未設定時は dev モード・暗号化なし）
	var lstepCipher *appCrypto.AESGCMCipher
	if cfg.IntegrationEncryptionKey != "" {
		c, err := appCrypto.NewAESGCMCipher(cfg.IntegrationEncryptionKey)
		if err != nil {
			logger.Error("failed to initialize AES-GCM cipher", slog.String("error", err.Error()))
			os.Exit(1)
		}
		lstepCipher = c
		logger.Info("LSTEP cipher: AES-256-GCM enabled")
	} else {
		logger.Info("INTEGRATION_ENCRYPTION_KEY not set: running without encryption (dev mode)")
	}
	svcs.LstepSettings = service.NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, lstepCipher, svcs.Audit, repos.ClinicSettings)
	svcs.LstepTagSync = service.NewLstepTagSyncService(
		svcs.LstepSettings,
		repos.Owner,
		repos.Vaccination,
		repos.MedicalRecord,
		repos.Accounting,
		repos.LstepTagCache,
		repos.Pet,
		repos.Prescription,
		repos.Checkup,
		repos.Reservation,
		repos.LstepSyncErrorCounter,
		repos.LstepTagCodeMapping,
		repos.BillingItem,
	)
	svcs.LstepLifecycle = service.NewLstepLifecycleService(
		svcs.LstepSettings,
		repos.Owner,
		repos.Pet,
		repos.LstepTagCache,
		svcs.LstepTagSync,
		svcs.Audit,
	)
	svcs.LstepTag = service.NewLstepTagService(
		svcs.LstepSettings,
		repos.Owner,
		repos.LstepTagCache,
		svcs.Audit,
	)

	// 共有ファイルストレージ初期化（STORAGE_TYPE=s3 で S3、それ以外はローカル）
	var sharedStorage infra.FileStorage
	if os.Getenv("STORAGE_TYPE") == "s3" {
		s3fs, err := infra.NewS3FileStorage(context.Background(), cfg.S3SharedBucket, cfg.S3SharedRegion)
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
	svcs.SharedFile = service.NewSharedFileService(repos.SharedFile, repos.Owner, sharedStorage)
	// LSTEP-BE-012: 慢性疾患フラグ
	svcs.ChronicCondition = service.NewChronicConditionService(repos.ChronicCondition, repos.Pet, svcs.LstepTagSync)
	// LSTEP-BE-013: LINE個別送信
	svcs.LineSend = service.NewLineSendService(svcs.LstepSettings, repos.Owner, svcs.SharedFile, repos.LstepTagCache, svcs.Audit, repos.LineSendLog)
	// LSTEP-BE-021: LINE User ID 自動取得・飼い主紐付け
	svcs.LineLink = service.NewLineLinkService(repos.Owner, repos.LineLinkToken, repos.LineReservationSetting, svcs.Audit)
	// LSTEP-BE-020: タグ集計・タグ別飼い主検索
	svcs.LstepTagSummary = service.NewLstepTagSummaryService(repos.LstepTagCache)
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	svcs.CheckupSync = service.NewCheckupSyncService(repos.CheckupSync, repos.Owner, repos.Pet, repos.LstepTagCache, svcs.LstepSettings, svcs.Audit)
	// FEAT-384: 自動配信トリガー監視
	svcs.LstepDeliveryMonitor = service.NewLstepDeliveryMonitorService(repos.LstepDeliveryTriggerLog)
	// Q23: トリガー優先順位設定
	svcs.LstepTriggerPriority = service.NewLstepTriggerPriorityService(repos.LstepTriggerPriority)
	// FEAT-383: 自動配信トリガー（LstepBatch / MedicalRecord / Checkup より先に初期化）
	svcs.LstepDeliveryTrigger = service.NewLstepDeliveryTriggerService(repos.Owner, repos.MedicalRecord, repos.Vaccination, repos.BillingItem, repos.Pet, repos.LstepTagCache, repos.LstepDeliveryTriggerLog, svcs.LstepSettings, svcs.LstepTriggerPriority)
	// FEAT-383: イベントフック注入（LstepDeliveryTrigger 確定後に再初期化）
	svcs.MedicalRecord = service.NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan, repos.LineCustomerMgr, repos.Reservation, svcs.LstepDeliveryTrigger, svcs.Audit)
	svcs.Checkup = service.NewCheckupService(repos.Checkup, svcs.LstepDeliveryTrigger)
	// LSTEP-BE-014: ノーショウ検知バッチ（LstepDeliveryTrigger 確定後に初期化）
	svcs.LstepBatch = service.NewLstepBatchService(repos.Reservation, svcs.LstepTagSync, repos.Clinic, repos.MedicalRecord, svcs.Audit, svcs.LstepSettings, svcs.LstepDeliveryTrigger)
	// FEAT-385: Lステップ CSV インポート・分析
	svcs.LstepCsvImport = service.NewLstepCsvImportService(repos.DB(), repos.LstepCsvImport, repos.LstepFriendAttributeSnapshot, repos.Owner)
	svcs.LstepAnalytics = service.NewLstepAnalyticsService(repos.Owner, repos.LstepDeliveryTriggerLog, repos.LstepFriendAttributeSnapshot)

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

	// LSTEP-BE-014: ノーショウ検知バッチ — 毎時0分に起動し 10/15/20 時 (JST) のみ実行
	go func() {
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
		triggerHours := map[int]bool{10: true, 15: true, 20: true}
		for {
			now := time.Now().In(jst)
			next := now.Truncate(time.Hour).Add(time.Hour)
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-appCtx.Done():
				timer.Stop()
				return
			case <-timer.C:
				h := time.Now().In(jst).Hour()
				if triggerHours[h] {
					if batchErr := svcs.LstepBatch.RunNoShowCheckAllClinics(appCtx); batchErr != nil {
						logger.Error("no-show batch failed", slog.String("error", batchErr.Error()))
					}
				}
			}
		}
	}()

	// LSTEP-BE-004: 休眠検知バッチ — 毎日 02:00 JST に実行
	go func() {
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
		for {
			now := time.Now().In(jst)
			// 翌日の 02:00 JST を計算
			next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, jst)
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-appCtx.Done():
				timer.Stop()
				return
			case <-timer.C:
				if batchErr := svcs.LstepBatch.RunDormantDetectionAllClinics(appCtx); batchErr != nil {
					logger.Error("dormant detection batch failed", slog.String("error", batchErr.Error()))
				}
			}
		}
	}()

	// FEAT-383: 自動配信トリガーバッチ — 毎時0分に起動（10:00 JST 固定）
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(time.Hour).Add(time.Hour)
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-appCtx.Done():
				timer.Stop()
				return
			case <-timer.C:
				if batchErr := svcs.LstepBatch.RunDeliveryTriggerBatchAllClinics(appCtx); batchErr != nil {
					logger.Error("delivery trigger batch failed", slog.String("error", batchErr.Error()))
				}
			}
		}
	}()

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
